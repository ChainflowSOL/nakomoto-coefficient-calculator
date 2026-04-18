package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xenowits/nakamoto-coefficient-calculator/core/chains"
)

const DBPath = "/app/data/nc_history.db"

const defaultBaseURL = "https://nakaflow.io"

type JsonResponse struct {
	ChainName     string `json:"chain_name"`
	ChainToken    string `json:"chain_token"`
	NakaCoPrevVal int    `json:"naka_co_prev_val"`
	NakaCoCurrVal int    `json:"naka_co_curr_val"`
	Change        int    `json:"naka_co_change_val"`
}

func main() {
	var mu sync.Mutex
	chainState := chains.NewState()
	lastUpdated := time.Now().UTC()

	db, err := InitDB(DBPath)
	if err != nil {
		log.Printf("Failed to initialize history DB: %v. History will be unavailable.", err)
	}

	if db != nil {
		coefficients := getListOfCoefficients(chainState)
		if err := SaveNCSnapshot(db, coefficients); err != nil {
			log.Printf("Failed to save initial NC snapshot: %v", err)
		}
	}

	ticker := time.NewTicker(6 * time.Hour)
	quit := make(chan struct{})
	defer close(quit)

	go func(state chains.ChainState) {
		for {
			select {
			case <-ticker.C:
				log.Println("Ticker ticked")
				newState := chains.RefreshChainState(chainState)

				mu.Lock()
				chainState = newState
				lastUpdated = time.Now().UTC()
				mu.Unlock()

				if db != nil {
					coefficients := getListOfCoefficients(newState)
					if err := SaveNCSnapshot(db, coefficients); err != nil {
						log.Printf("Failed to save NC snapshot: %v", err)
					}
				}

				fmt.Println(chainState)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(chainState)

	baseURL := strings.TrimRight(os.Getenv("NAKAFLOW_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/naka-coeffs", func(c *gin.Context) {
		mu.Lock()
		coefficients := getListOfCoefficients(chainState)
		ts := lastUpdated
		mu.Unlock()
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(200, gin.H{
			"coefficients": coefficients,
			"last_updated": ts.Format(time.RFC3339),
		})
	})

	r.GET("/solana-details", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(200, gin.H{
			"entities": chains.SolanaNakamotoDetails,
		})
	})

	// --- NC History endpoint ---
	//   GET /nc-history?chain=SOL          → all history for Solana
	//   GET /nc-history?chain=SOL&days=90  → last 90 days for Solana
	//   GET /nc-history?days=30            → last 30 days for all chains
	r.GET("/nc-history", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")

		if db == nil {
			c.JSON(503, gin.H{"error": "history database not available"})
			return
		}

		chainToken := c.Query("chain")
		daysStr := c.DefaultQuery("days", "0")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			days = 0
		}

		if chainToken != "" {
			records, err := GetNCHistory(db, chainToken, days)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			if records == nil {
				records = []HistoryRecord{}
			}
			c.JSON(200, gin.H{
				"chain":   chainToken,
				"days":    days,
				"records": records,
			})
		} else {
			if days == 0 {
				days = 30 
			}
			records, err := GetAllChainsLatestHistory(db, days)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			if records == nil {
				records = []HistoryRecord{}
			}
			c.JSON(200, gin.H{
				"days":    days,
				"records": records,
			})
		}
	})

	r.GET("/feed.xml", func(c *gin.Context) {
		mu.Lock()
		coefficients := getListOfCoefficients(chainState)
		ts := lastUpdated
		mu.Unlock()

		body, err := renderRSS(baseURL, coefficients, ts)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "public, max-age=600")
		c.Data(200, "application/rss+xml; charset=utf-8", body)
	})

	r.GET("/embed/badge/:chain", func(c *gin.Context) {
		token := chains.Token(strings.ToUpper(c.Param("chain")))
		mu.Lock()
		chain, ok := chainState[token]
		mu.Unlock()
		if !ok {
			c.JSON(404, gin.H{"error": "unknown chain"})
			return
		}
		svg := renderBadgeSVG(token.ChainName(), chain.CurrNCVal)
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "public, max-age=300")
		c.Data(200, "image/svg+xml; charset=utf-8", []byte(svg))
	})

	r.GET("/embed/widget/:chain", func(c *gin.Context) {
		token := chains.Token(strings.ToUpper(c.Param("chain")))
		mu.Lock()
		chain, ok := chainState[token]
		mu.Unlock()
		if !ok {
			c.JSON(404, gin.H{"error": "unknown chain"})
			return
		}
		page := renderWidgetHTML(widgetData{
			ChainName:  token.ChainName(),
			ChainToken: string(token),
			NC:         chain.CurrNCVal,
			Prev:       chain.PrevNCVal,
			Change:     chain.CurrNCVal - chain.PrevNCVal,
			BaseURL:    baseURL,
		})
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "public, max-age=300")
		c.Data(200, "text/html; charset=utf-8", []byte(page))
	})

	r.GET("/openapi.json", func(c *gin.Context) {
		body, err := openAPISpec(baseURL)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Cache-Control", "public, max-age=3600")
		c.Data(200, "application/json; charset=utf-8", body)
	})

	r.GET("/docs", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(swaggerHTML(baseURL)))
	})

	r.Run(":8080")
}

func getListOfCoefficients(state chains.ChainState) []JsonResponse {
	var coeffs []JsonResponse
	for token, chain := range state {
		coeffs = append(coeffs, JsonResponse{
			ChainName:     token.ChainName(),
			ChainToken:    string(token),
			NakaCoPrevVal: chain.PrevNCVal,
			NakaCoCurrVal: chain.CurrNCVal,
			Change:        chain.CurrNCVal - chain.PrevNCVal,
		})
	}

	sort.Slice(coeffs, func(i, j int) bool {
		if coeffs[i].ChainToken < coeffs[j].ChainToken {
			return true
		}

		return false
	})

	return coeffs
}