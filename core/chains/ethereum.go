package chains

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"
)

type RatedResponse struct {
	Data []RatedOperator `json:"data"`
}

type RatedOperator struct {
	ID                 string  `json:"id"`          
	NetworkPenetration float64 `json:"networkPenetration"` 
	ValidatorCount     int     `json:"validatorCount"`
}

func Ethereum() (int, error) {
	// Rated Network API
	url := "https://api.rated.network/v0/eth/operators?window=1d"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	apiKey := os.Getenv("RATED_API_KEY")
	if apiKey == "" {
		return 0, fmt.Errorf("RATED_API_KEY is missing")
	}
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("X-Rated-Network", "mainnet")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("rated api returned status %d: %s", resp.StatusCode, string(body))
	}

	var response RatedResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to parse eth response: %v", err)
	}

	operators := response.Data
	if len(operators) == 0 {
		return 0, fmt.Errorf("no operators found in rated response")
	}

	sort.Slice(operators, func(i, j int) bool {
		return operators[i].NetworkPenetration > operators[j].NetworkPenetration
	})

	totalShare := 0.0
	nakamotoCoefficient := 0
	threshold := 33.33

	for _, op := range operators {
		sharePercent := op.NetworkPenetration * 100
		
		totalShare += sharePercent
		nakamotoCoefficient++

		if totalShare > threshold {
			break
		}
	}

	log.Printf("The Nakamoto coefficient for Ethereum is %d", nakamotoCoefficient)
	return nakamotoCoefficient, nil
}