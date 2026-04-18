package chains

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"sort"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type HyperliquidValidator struct {
	Validator string  `json:"validator"` 
	Name      string  `json:"name"`
	Stake     float64 `json:"stake"`  
	IsActive  bool    `json:"isActive"` 
}

type HyperliquidResponse []HyperliquidValidator

var hyperliquidClient = http11Client(10 * time.Second)

func Hyperliquid() (int, error) {
	url := "https://api.hyperliquid.xyz/info"
	payload := []byte(`{"type": "validatorSummaries"}`)

	var resp *http.Response
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		req, reqErr := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if reqErr != nil {
			return 0, reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) Chainflow/nc-calc")

		resp, err = hyperliquidClient.Do(req)
		if err == nil {
			break
		}
		log.Printf("hyperliquid: request attempt %d failed: %v", attempt+1, err)
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var validators HyperliquidResponse
	err = json.Unmarshal(body, &validators)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hyperliquid response: %v", err)
	}

	var votingPowers []*big.Int
	totalVotingPower := big.NewInt(0)
	activeCount := 0

	for _, v := range validators {
		if !v.IsActive {
			continue
		}

		vp := big.NewInt(int64(v.Stake))
		
		votingPowers = append(votingPowers, vp)
		totalVotingPower.Add(totalVotingPower, vp)
		activeCount++
	}

	if len(votingPowers) == 0 {
		return 0, fmt.Errorf("no active validators found for Hyperliquid")
	}

	sort.Slice(votingPowers, func(i, j int) bool {
		return votingPowers[i].Cmp(votingPowers[j]) > 0
	})

	log.Printf("Hyperliquid: Fetched %d active validators. Total Stake: %s", activeCount, totalVotingPower.String())

	nakamotoCoefficient := utils.CalcNakamotoCoefficientBigInt(totalVotingPower, votingPowers)
	log.Printf("The Nakamoto coefficient for Hyperliquid is %d", nakamotoCoefficient)

	return nakamotoCoefficient, nil
}