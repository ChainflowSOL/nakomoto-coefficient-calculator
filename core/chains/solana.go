package chains

import (
	"encoding/json"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	utils "github.com/xenowits/nakamoto-coefficient-calculator/core/utils"
)

// --- CONFIGURATION ---
// Try multiple paths for the entities file (Docker path, then local paths)
var EntitiesFilePaths = []string{
	"/app/solana_entities.json",
	"/app/keybase_ids.json",
	"solana_entities.json",
	"keybase_ids.json",
	"core/chains/solana_entities.json",
	"core/chains/keybase_ids.json",
}

// HardcodedEntityStakes contains total stake (in SOL) for entities whose full
// validator set cannot be tracked via the API. Any validators detected from the
// API that match these entity names will be SKIPPED to avoid double-counting.
var HardcodedEntityStakes = map[string]float64{
	"Coinbase": 38.66e6, // 38.66M SOL — source: on-chain analysis
}

// --- DATA STRUCTURES ---

type EntityOverrideMap map[string][]string

type EntityDetail struct {
	Rank       int     `json:"rank"`
	Name       string  `json:"name"`
	StakeSOL   float64 `json:"stake_sol"`
	Percent    float64 `json:"percent"`
	Cumulative float64 `json:"cumulative"`
	Commission float64 `json:"commission"`
	SkipRate   float64 `json:"skip_rate"`
	IsVerified bool    `json:"is_verified"`
}

var SolanaNakamotoDetails []EntityDetail

type SolanaResponse []struct {
	Name         string  `json:"name"`
	Account      string  `json:"keybase_id"`
	Active_stake int64   `json:"active_stake"`
	Delinquent   bool    `json:"delinquent"`
	Pubkey       string  `json:"account"`
	Commission   int     `json:"commission"`
	SkipRate     float64 `json:"skip_rate"`
}

type EntityAggregator struct {
	TotalStake   *big.Int
	WeightedComm *big.Float
	WeightedSkip *big.Float
	IsVerified   bool
}

// --- HELPER FUNCTIONS ---

func loadEntityOverrides() map[string]string {
	var file *os.File
	var err error
	var loadedPath string

	for _, path := range EntitiesFilePaths {
		file, err = os.Open(path)
		if err == nil {
			loadedPath = path
			break
		}
	}

	if file == nil {
		log.Printf("⚠️ No manual entities file found at any of %v. Using auto-grouping only.", EntitiesFilePaths)
		return make(map[string]string)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var manualMap EntityOverrideMap
	if err := json.Unmarshal(byteValue, &manualMap); err != nil {
		log.Printf("❌ Error parsing entities file %s: %v", loadedPath, err)
		return make(map[string]string)
	}

	pubkeyToEntity := make(map[string]string)
	for entityName, pubkeys := range manualMap {
		for _, pubkey := range pubkeys {
			pubkeyToEntity[pubkey] = entityName
		}
	}

	log.Printf("✅ Loaded %d manual validator overrides from %s.", len(pubkeyToEntity), loadedPath)
	return pubkeyToEntity
}

// normalizeEntityName maps validator names/keybase_ids to canonical entity names.
// This handles cases where entities run multiple validators with slightly different names.
func normalizeEntityName(name string) string {
	lower := strings.ToLower(name)

	// --- Major exchanges & custodians ---
	if strings.Contains(lower, "coinbase") {
		return "Coinbase"
	}
	if strings.Contains(lower, "binance") {
		return "Binance"
	}
	if strings.Contains(lower, "kraken") {
		return "Kraken"
	}
	if strings.Contains(lower, "upbit") {
		return "Upbit Staking"
	}
	if strings.Contains(lower, "okx") || strings.Contains(lower, "okex") {
		return "OKX"
	}
	if strings.Contains(lower, "bybit") {
		return "Bybit"
	}
	if strings.Contains(lower, "bitfinex") {
		return "Bitfinex"
	}
	if strings.Contains(lower, "crypto.com") || strings.Contains(lower, "cryptocom") {
		return "Crypto.com"
	}
	if strings.Contains(lower, "hashkey") {
		return "HashKey"
	}

	// --- Major staking providers ---
	if strings.Contains(lower, "ledger by figment") {
		return "Figment"
	}
	if strings.Contains(lower, "figment") {
		return "Figment"
	}
	if strings.Contains(lower, "everstake") {
		return "Everstake"
	}
	if strings.Contains(lower, "chorus") && strings.Contains(lower, "one") {
		return "Bitwise Onchain Solutions" // Chorus One acquired by Bitwise (Feb 2026)
	}
	if strings.Contains(lower, "chorusone") {
		return "Bitwise Onchain Solutions" // Chorus One acquired by Bitwise (Feb 2026)
	}
	if strings.Contains(lower, "p2p") && (strings.Contains(lower, "validator") || strings.Contains(lower, "staking") || strings.Contains(lower, ".org")) {
		return "P2P Validator"
	}
	if lower == "p2p" || strings.HasPrefix(lower, "p2p.") {
		return "P2P Validator"
	}
	if strings.Contains(lower, "staking facilities") || strings.Contains(lower, "stakingfacilities") {
		return "Staking Facilities"
	}
	if strings.Contains(lower, "twinstake") {
		return "Twinstake"
	}
	if strings.Contains(lower, "blockdaemon") {
		return "Blockdaemon"
	}
	if strings.Contains(lower, "allnodes") {
		return "Allnodes"
	}
	if strings.Contains(lower, "certus") && strings.Contains(lower, "one") {
		return "Certus One"
	}
	if strings.Contains(lower, "solflare") {
		return "SolFlare"
	}

	// --- Solana ecosystem entities ---
	if strings.Contains(lower, "helius") {
		return "Helius"
	}
	if strings.Contains(lower, "jupiter") || strings.Contains(lower, "jup.ag") {
		return "Jupiter"
	}
	if strings.Contains(lower, "jito") {
		return "Jito"
	}
	if strings.HasPrefix(lower, "mrgn") || strings.Contains(lower, "marginfi") {
		return "mrgn"
	}
	if strings.Contains(lower, "marinade") {
		return "Marinade"
	}
	if strings.Contains(lower, "blazestake") || strings.Contains(lower, "blaze") {
		return "BlazeStake"
	}
	if strings.Contains(lower, "lido") {
		return "Lido"
	}

	// --- Institutional ---
	if strings.Contains(lower, "galaxy") {
		return "Galaxy"
	}
	if strings.Contains(lower, "bitwise") {
		return "Bitwise Onchain Solutions"
	}
	if strings.Contains(lower, "sol strategies") || strings.Contains(lower, "solstrategies") {
		return "Sol Strategies"
	}
	if strings.Contains(lower, "forward") && strings.Contains(lower, "industr") {
		return "Forward Industries"
	}

	// --- Infrastructure providers ---
	if strings.Contains(lower, "kiln") {
		return "Kiln"
	}
	if strings.Contains(lower, "triton") {
		return "Triton"
	}
	if strings.Contains(lower, "syndica") {
		return "Syndica"
	}

	// --- Generic cleanup: strip trailing numbers, pipe/colon suffixes ---
	re := regexp.MustCompile(`[|(:].*`)
	clean := re.ReplaceAllString(name, "")
	reNum := regexp.MustCompile(`\s*\d+$`)
	clean = reNum.ReplaceAllString(clean, "")
	return strings.TrimSpace(clean)
}

// --- MAIN FUNCTION ---

func Solana() (int, error) {
	// 1. LOAD MANUAL OVERRIDES
	overrideMap := loadEntityOverrides()

	// 2. FETCH DATA FROM API
	url := "https://www.validators.app/api/v1/validators/mainnet.json"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	apiKey := os.Getenv("VALIDATORS_APP_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SOLANA_API_KEY")
	}
	req.Header.Add("Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response SolanaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}

	// 3. GROUP VALIDATORS BY ENTITY
	groups := make(map[string]*EntityAggregator)
	skippedStake := make(map[string]int64) // track how much API stake we skipped for hardcoded entities

	for _, ele := range response {
		if ele.Delinquent {
			continue
		}

		var entityID string
		isManual := false

		// Priority 1: Manual override by pubkey
		if manualName, found := overrideMap[ele.Pubkey]; found {
			entityID = manualName
			isManual = true
		} else {
			// Priority 2: Normalize the keybase_id or name
			if ele.Account != "" {
				entityID = normalizeEntityName(ele.Account)
			} else if ele.Name != "" {
				entityID = normalizeEntityName(ele.Name)
			} else {
				// Priority 3: Check if the pubkey itself has a vanity prefix
				entityID = normalizeVanityPubkey(ele.Pubkey)
			}
		}

		// Skip validators belonging to hardcoded entities (to avoid double-counting)
		if _, isHardcoded := HardcodedEntityStakes[entityID]; isHardcoded {
			skippedStake[entityID] += ele.Active_stake
			continue
		}

		if _, exists := groups[entityID]; !exists {
			groups[entityID] = &EntityAggregator{
				TotalStake:   big.NewInt(0),
				WeightedComm: big.NewFloat(0),
				WeightedSkip: big.NewFloat(0),
				IsVerified:   false,
			}
		}

		groups[entityID].IsVerified = groups[entityID].IsVerified || isManual

		stake := big.NewInt(ele.Active_stake)
		stakeFloat := new(big.Float).SetInt(stake)
		groups[entityID].TotalStake.Add(groups[entityID].TotalStake, stake)

		commFloat := big.NewFloat(float64(ele.Commission))
		weightedComm := new(big.Float).Mul(stakeFloat, commFloat)
		groups[entityID].WeightedComm.Add(groups[entityID].WeightedComm, weightedComm)

		skipFloat := big.NewFloat(ele.SkipRate)
		weightedSkip := new(big.Float).Mul(stakeFloat, skipFloat)
		groups[entityID].WeightedSkip.Add(groups[entityID].WeightedSkip, weightedSkip)
	}

	// 3b. INJECT HARDCODED ENTITIES
	for entityName, stakeSOL := range HardcodedEntityStakes {
		stakeLamports := new(big.Int).SetUint64(uint64(stakeSOL * 1e9))
		groups[entityName] = &EntityAggregator{
			TotalStake:   stakeLamports,
			WeightedComm: big.NewFloat(0),
			WeightedSkip: big.NewFloat(0),
			IsVerified:   true,
		}
		skippedLamports := skippedStake[entityName]
		log.Printf("💎 Hardcoded %s: %.2fM SOL (replaced %.2fM SOL detected from API)",
			entityName, stakeSOL/1e6, float64(skippedLamports)/1e9/1e6)
	}

	// 4. PREPARE FINAL LIST
	var detailsList []EntityDetail
	var votingPowers []big.Int
	totalNetworkStake := new(big.Float).SetFloat64(0)

	for name, data := range groups {
		votingPowers = append(votingPowers, *data.TotalStake)
		sFloat := new(big.Float).SetInt(data.TotalStake)
		totalNetworkStake.Add(totalNetworkStake, sFloat)

		sVal, _ := sFloat.Float64()
		var c, sk float64
		if sVal > 0 {
			avgComm := new(big.Float).Quo(data.WeightedComm, sFloat)
			avgSkip := new(big.Float).Quo(data.WeightedSkip, sFloat)
			c, _ = avgComm.Float64()
			sk, _ = avgSkip.Float64()
		}

		detailsList = append(detailsList, EntityDetail{
			Name:       name,
			StakeSOL:   sVal / 1e9,
			Commission: c,
			SkipRate:   sk,
			IsVerified: data.IsVerified,
		})
	}

	sort.Slice(detailsList, func(i, j int) bool { return detailsList[i].StakeSOL > detailsList[j].StakeSOL })
	sort.Slice(votingPowers, func(i, j int) bool { return votingPowers[i].Cmp(&votingPowers[j]) == 1 })

	totalVP := utils.CalculateTotalVotingPowerBigNums(votingPowers)
	nakamotoCoefficient := utils.CalcNakamotoCoefficientBigNums(totalVP, votingPowers)

	tStake, _ := totalNetworkStake.Float64()
	var cumulative float64 = 0
	for i := range detailsList {
		detailsList[i].Rank = i + 1
		percent := (detailsList[i].StakeSOL * 1e9 / tStake) * 100
		detailsList[i].Percent = percent
		cumulative += percent
		detailsList[i].Cumulative = cumulative
	}


	if len(detailsList) > 300 {
		detailsList = detailsList[:300]
	}
	SolanaNakamotoDetails = detailsList

	log.Printf("Calculated Solana NC: %d", nakamotoCoefficient)
	return nakamotoCoefficient, nil
}

// normalizeVanityPubkey checks if a pubkey has a known vanity prefix
// (e.g., mrgn validators use pubkeys starting with "mrgn")
func normalizeVanityPubkey(pubkey string) string {
	lower := strings.ToLower(pubkey)

	if strings.HasPrefix(lower, "mrgn") {
		return "mrgn"
	}
	if strings.HasPrefix(lower, "jito") || strings.HasPrefix(lower, "j1to") {
		return "Jito"
	}

	// No vanity match, return the raw pubkey
	return pubkey
}