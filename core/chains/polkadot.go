package chains

// Polkadot uses Nominated Proof-of-Stake (NPoS), and its consensus security
// model differs from stake-weighted PoS chains (e.g. Ethereum):
//
// Stake gates entry to the 600-slot active set (via the Phragmén election)
// and creates slashing exposure, but does not translate to voting power once
// elected. Consensus capture is therefore a function of how many validator
// slots a single operator controls, not how much stake is concentrated. We
// compute the Nakamoto coefficient as the minimum number of operators whose
// combined validator count reaches >= 1/3 of the active set, grouping
// validators by their on-chain identity (sub-identities rolled into the
// parent; operators that register multiple parent identities with numeric
// suffixes are collapsed; anonymous validators are treated as independent
// singletons).

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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

// Polkadot's NPoS active set is currently 600. Phragmén elects the validators
// with the highest backed stake, so the top N by bonded_total approximates
// the active set returned by Subscan.
const polkadotActiveSetSize = 600

type polkadotPeople struct {
	Display  string `json:"display"`
	Identity bool   `json:"identity"`
	Parent   *struct {
		Display string `json:"display"`
	} `json:"parent"`
}

type polkadotValidator struct {
	BondedTotal         string `json:"bonded_total"`
	StashAccountDisplay struct {
		Address string          `json:"address"`
		People  *polkadotPeople `json:"people"`
	} `json:"stash_account_display"`
}

type polkadotResponse struct {
	Data struct {
		List []polkadotValidator `json:"list"`
	} `json:"data"`
}

var (
	polkadotNormalizeRe   = regexp.MustCompile(`[^a-z0-9]+`)
	polkadotNumericSuffix = regexp.MustCompile(`^(.{3,}?)(\d+)$`)
	polkadotTLDTokens     = []string{"com", "io", "org", "xyz", "net", "foundation", "tech", "app", "dev"}
)

func Polkadot() (int, error) {
	validators, err := fetchPolkadotValidators()
	if err != nil {
		return 0, err
	}

	// Sort by bonded_total desc and take the top N as the active set.
	sort.Slice(validators, func(i, j int) bool {
		bi, _ := strconv.ParseInt(validators[i].BondedTotal, 10, 64)
		bj, _ := strconv.ParseInt(validators[j].BondedTotal, 10, 64)
		return bi > bj
	})
	if len(validators) > polkadotActiveSetSize {
		validators = validators[:polkadotActiveSetSize]
	}
	if len(validators) == 0 {
		return 0, fmt.Errorf("no validators found for polkadot")
	}

	counts := groupPolkadotValidatorsByOperator(validators)

	// Count-based Nakamoto: treat each entity's validator count as its
	// "voting power" so the existing 33% utility computes the right number.
	var votingPowers []int64
	for _, c := range counts {
		votingPowers = append(votingPowers, int64(c))
	}
	total := utils.CalculateTotalVotingPower(votingPowers)
	nakamotoCoefficient := utils.CalcNakamotoCoefficient(total, votingPowers)

	log.Printf(
		"Polkadot count-based NC: %d entities control >= 1/3 of %d active validators (across %d operator groups)",
		nakamotoCoefficient, total, len(counts),
	)
	return nakamotoCoefficient, nil
}

// fetchPolkadotValidators paginates Subscan's /scan/staking/validators endpoint
// until it has enough rows to cover the active set. The previous version capped
// at page 3 (max 400 rows), which truncated the tail of the active set; we now
// stop when we have at least polkadotActiveSetSize rows or the API returns an
// empty page.
func fetchPolkadotValidators() ([]polkadotValidator, error) {
	var all []polkadotValidator
	for page := 0; ; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		payload := bytes.NewBuffer([]byte(fmt.Sprintf(
			`{"order":"desc","order_field":"bonded_total","row":100,"page":%d,"key":"validator"}`, page,
		)))

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://assethub-polkadot.api.subscan.io/api/scan/staking/validators", payload)
		if err != nil {
			cancel()
			return nil, errors.New("create post request for polkadot")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", os.Getenv("SUBSCAN_API_KEY"))

		resp, err := new(http.Client).Do(req)
		if err != nil {
			cancel()
			return nil, errors.New("post request unsuccessful")
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("subscan API returned status %d: %.200s", resp.StatusCode, string(body))
		}

		var response polkadotResponse
		if err = json.Unmarshal(body, &response); err != nil {
			return nil, err
		}
		if len(response.Data.List) == 0 {
			break
		}
		all = append(all, response.Data.List...)
		if len(response.Data.List) < 100 || len(all) >= polkadotActiveSetSize {
			break
		}
	}
	return all, nil
}

// groupPolkadotValidatorsByOperator returns the number of active validators
// controlled by each operator, where an operator is identified by:
//
//  1. The parent identity, if this validator is registered as a sub-identity
//     (e.g. Coinbase/v04 -> "Coinbase").
//  2. The validator's own on-chain display name, if it has one.
//  3. The stash address, otherwise (anonymous validator is its own singleton).
//
// Operators that register multiple parent identities instead of using
// sub-identities (e.g. BINANCE_STAKE_1..BINANCE_STAKE_15, figment1..figment8)
// are collapsed via two deterministic rules applied after lowercasing and
// stripping non-alphanumerics:
//
//   - numeric-suffix collapse: trailing digits are stripped from any name with
//     a >=3-character prefix.
//   - TLD collapse: trailing TLD-like tokens (com, io, org, ...) are stripped
//     only when the stripped form is also a known operator in the set, to
//     avoid false positives.
func groupPolkadotValidatorsByOperator(validators []polkadotValidator) map[string]int {
	// Pass 1: extract a raw identity string for each validator.
	raw := make([]string, len(validators))
	for i, v := range validators {
		raw[i] = identityForValidator(v)
	}

	// Pass 2: build the normalized universe and apply numeric-suffix collapse.
	postNumeric := map[string]string{}
	universe := map[string]bool{}
	for _, r := range raw {
		if r == "" {
			continue
		}
		n := normalizePolkadotName(r)
		if n == "" {
			continue
		}
		root := numericRoot(n)
		postNumeric[n] = root
		universe[root] = true
	}

	// Pass 3: apply TLD collapse only where the stripped form is in the set.
	finalLabel := map[string]string{}
	for n, root := range postNumeric {
		finalLabel[n] = tldCollapse(root, universe)
	}

	// Pass 4: count validators per final group label.
	counts := map[string]int{}
	for i, v := range validators {
		r := raw[i]
		if r == "" {
			counts[v.StashAccountDisplay.Address]++
			continue
		}
		n := normalizePolkadotName(r)
		if n == "" {
			counts[v.StashAccountDisplay.Address]++
			continue
		}
		group, ok := finalLabel[n]
		if !ok {
			group = n
		}
		counts[group]++
	}
	return counts
}

func identityForValidator(v polkadotValidator) string {
	p := v.StashAccountDisplay.People
	if p == nil {
		return ""
	}
	if p.Parent != nil && p.Parent.Display != "" {
		return p.Parent.Display
	}
	return p.Display
}

func normalizePolkadotName(s string) string {
	return polkadotNormalizeRe.ReplaceAllString(strings.ToLower(s), "")
}

func numericRoot(s string) string {
	if m := polkadotNumericSuffix.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

func tldCollapse(s string, universe map[string]bool) string {
	for _, tld := range polkadotTLDTokens {
		if strings.HasSuffix(s, tld) && len(s) > len(tld)+2 {
			stripped := s[:len(s)-len(tld)]
			if universe[stripped] {
				return stripped
			}
		}
	}
	return s
}
