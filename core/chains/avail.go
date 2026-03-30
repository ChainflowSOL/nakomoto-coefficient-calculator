package chains

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"sort"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

type AvailResponse struct {
	Data struct {
		List []struct {
			BondedTotal string `json:"bonded_total"`
		} `json:"list"`
	} `json:"data"`
}

func Avail() (int, error) {
	var votingPowers []*big.Int
	page := 0

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		payload := bytes.NewBuffer([]byte(fmt.Sprintf(`{"order":"desc","order_field":"bonded_total","row":100,"page":%d}`, page)))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://avail.api.subscan.io/api/scan/staking/validators", payload)
		if err != nil {
			cancelFunc()
			return 0, errors.New("create post request for avail")
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
			return 0, fmt.Errorf("subscan avail returned status %d: %.200s", resp.StatusCode, string(body))
		}

		var response AvailResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return 0, err
		}

		if len(response.Data.List) == 0 {
			break
		}

		for _, ele := range response.Data.List {
			bondedTotal := new(big.Int)
			if _, ok := bondedTotal.SetString(ele.BondedTotal, 10); !ok {
				log.Println("Error parsing bonded total:", ele.BondedTotal)
				continue
			}
			votingPowers = append(votingPowers, bondedTotal)
		}

		if len(response.Data.List) < 100 {
			break
		}
		page++
	}

	if len(votingPowers) == 0 {
		return 0, fmt.Errorf("no validators found for avail")
	}

	log.Printf("Avail: fetched %d validators", len(votingPowers))

	sort.Slice(votingPowers, func(i, j int) bool { return votingPowers[i].Cmp(votingPowers[j]) > 0 })

	totalVotingPower := utils.CalculateTotalVotingPowerBigInt(votingPowers)
	nakamotoCoefficient := utils.CalcNakamotoCoefficientBigInt(totalVotingPower, votingPowers)
	log.Printf("The Nakamoto coefficient for Avail is %d", nakamotoCoefficient)

	return nakamotoCoefficient, nil
}
