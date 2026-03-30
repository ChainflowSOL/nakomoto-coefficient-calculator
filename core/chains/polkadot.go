package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type PolkadotResponse struct {
	Data struct {
		List []struct {
			BondedTotal string `json:"bonded_total"`
		} `json:"list"`
	} `json:"data"`
}

func Polkadot() (int, error) {
	var votingPowers []int64
	page := 0

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		payload := bytes.NewBuffer([]byte(fmt.Sprintf(`{"order":"desc","order_field":"bonded_total","row":100,"page":%d}`, page)))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://polkadot.api.subscan.io/api/scan/staking/validators", payload)
		if err != nil {
			cancelFunc()
			return 0, errors.New("create post request for polkadot")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", os.Getenv("SUBSCAN_API_KEY"))

		resp, err := new(http.Client).Do(req)
		if err != nil {
			cancelFunc()
			return 0, errors.New("post request unsuccessful")
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancelFunc()
		if err != nil {
			return 0, err
		}

		if resp.StatusCode != 200 {
			return 0, fmt.Errorf("subscan API returned status %d: %.200s", resp.StatusCode, string(body))
		}

		var response PolkadotResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return 0, err
		}

		if len(response.Data.List) == 0 {
			break // no more pages
		}

		for _, ele := range response.Data.List {
			bondedTotal, err := strconv.ParseInt(ele.BondedTotal, 10, 64)
			if err != nil {
				log.Println(err)
				continue
			}
			votingPowers = append(votingPowers, bondedTotal)
		}

		// Polkadot has ~300 active validators, stop after 4 pages to be safe
		if len(response.Data.List) < 100 || page >= 3 {
			break
		}
		page++
	}

	if len(votingPowers) == 0 {
		return 0, fmt.Errorf("no voting powers found for polkadot")
	}

	log.Printf("Polkadot: fetched %d validators", len(votingPowers))

	sort.Slice(votingPowers, func(i, j int) bool { return votingPowers[i] > votingPowers[j] })

	totalVotingPower := utils.CalculateTotalVotingPower(votingPowers)
	nakamotoCoefficient := utils.CalcNakamotoCoefficient(totalVotingPower, votingPowers)
	log.Printf("The Nakamoto coefficient for Polkadot is %d", nakamotoCoefficient)

	return nakamotoCoefficient, nil
}
