package chains

import (
	"fmt"
	"log"
)

// Chain contains details of a particular Chain.
type Chain struct {
	PrevNCVal int
	CurrNCVal int
}

// Token represents the name of token for a blockchain.
// For example, ATOM for cosmos.
// It is used to identify a particular Chain.
type Token string

// ChainState contains complete NC information for all supported chains.
type ChainState map[Token]Chain

// Append new chains in alphabetical order only.
const (
	ADA   Token = "ADA"
	ALGO  Token = "ALGO"
	APT   Token = "APT"
	ATOM  Token = "ATOM"
	AVAIL Token = "AVAIL"
	AVAX  Token = "AVAX"
	BASE  Token = "BASE"
	BLD   Token = "BLD"
	BNB   Token = "BNB"
	DOT   Token = "DOT"
	EGLD  Token = "EGLD"
	ETH   Token = "ETH"
	GRT   Token = "GRT"
	HBAR  Token = "HBAR"
	HYPE  Token = "HYPE"
	JUNO  Token = "JUNO"
	MATIC Token = "MATIC"
	MINA  Token = "MINA"
	MON   Token = "MON"
	NAM   Token = "NAM"
	NIM   Token = "NIM"
	NEAR  Token = "NEAR"
	OSMO  Token = "OSMO"
	PLS   Token = "PLS"
	PLUME Token = "PLUME"
	REGEN Token = "REGEN"
	RUNE  Token = "RUNE"
	SEI   Token = "SEI"
	SOL   Token = "SOL"
	STARS Token = "STARS"
	STORY Token = "STORY"
	SUI   Token = "SUI"
	TIA   Token = "TIA"
	XNO   Token = "XNO"
)

// ChainName returns the name of the chain given the token name.
func (t Token) ChainName() string {
	switch t {
	case ADA:
		return "Cardano"
	case ALGO:
		return "Algo"
	case APT:
		return "Aptos"
	case ATOM:
		return "Cosmos"
	case AVAIL:
		return "Avail DA"
	case AVAX:
		return "Avalanche"
	case BASE:
		return "Base"
	case BLD:
		return "Agoric"
	case BNB:
		return "BNB Smart Chain"
	case DOT:
		return "Polkadot"
	case EGLD:
		return "MultiversX"
	case ETH:
		return "Ethereum"
	case GRT:
		return "Graph Protocol"
	case HBAR:
		return "Hedera"
	case HYPE:
		return "Hype"
	case JUNO:
		return "Juno"
	case MATIC:
		return "Polygon"
	case MINA:
		return "Mina Protocol"
	case MON:
		return "Monad"
	case NAM:
		return "Namada"
	case NIM:
		return "Nimiq"
	case NEAR:
		return "Near Protocol"
	case OSMO:
		return "Osmosis"
	case PLS:
		return "Pulsechain"
	case PLUME:
		return "Plume"
	case REGEN:
		return "Regen Network"
	case RUNE:
		return "Thorchain"
	case SEI:
		return "Sei"
	case SOL:
		return "Solana"
	case STARS:
		return "Stargaze"
	case STORY:
		return "Story Protocol"
	case SUI:
		return "Sui Protocol"
	case TIA:
		return "Celestia"
	case XNO:
		return "Nano"
	default:
		return "Unknown"
	}
}

var Tokens = []Token{ADA, ALGO, APT, ATOM, AVAIL, AVAX, BASE, BLD, BNB, DOT, EGLD, ETH, GRT, HBAR, HYPE, JUNO, MATIC, MINA, MON, NAM, NIM, NEAR, OSMO, PLS, PLUME, REGEN, RUNE, SEI, SOL, STARS, STORY, SUI, TIA, XNO}

// NewState returns a new fresh state.
func NewState() ChainState {
	state := make(ChainState)

	return RefreshChainState(state)
}

func RefreshChainState(prevState ChainState) ChainState {
	newState := make(ChainState)
	for _, token := range Tokens {
		currVal, err := newValues(token)
		if err != nil {
			log.Println("Failed to update chain info:", token, err)
			continue
		}

		newState[token] = Chain{
			PrevNCVal: prevState[token].CurrNCVal,
			CurrNCVal: currVal,
		}
	}

	return newState
}

func newValues(token Token) (int, error) {
	var (
		currVal int
		err     error
	)

	log.Printf("Calculating Nakamoto coefficient for %s", token.ChainName())

	switch token {
	case ADA:
		currVal, err = Cardano()
	case ALGO:
		currVal, err = Algorand()
	case APT:
		currVal, err = Aptos()
	case ATOM:
		currVal, err = Cosmos()
	case AVAIL:
		currVal, err = Avail()
	case AVAX:
		currVal, err = Avalanche()
	case BASE:
		currVal, err = Base()
	case BLD:
		currVal, err = Agoric()
	case BNB:
		currVal, err = BSC()
	case DOT:
		currVal, err = Polkadot()
	case EGLD:
		currVal, err = MultiversX()
	case ETH:
		currVal, err = Ethereum()
	case GRT:
		currVal, err = Graph()
	case HBAR:
		currVal, err = Hedera()
	case HYPE:
		currVal, err = Hyperliquid()
	case JUNO:
		currVal, err = Juno()
	case MATIC:
		currVal, err = Polygon()
	case MINA:
		currVal, err = Mina()
	case MON:
		log.Println("Attempting to calculate Monad Nakamoto coefficient...")
		currVal, err = Monad()
		if err != nil {
			log.Printf("Error calculating Monad Nakamoto coefficient: %v", err)
		}
	case NAM:
		currVal, err = Namada()
	case NIM:
		currVal, err = Nimiq()
	case NEAR:
		currVal, err = Near()
	case OSMO:
		currVal, err = Osmosis()
	case PLS:
		currVal, err = Pulsechain()
	case PLUME:
		currVal, err = Plume()
	case REGEN:
		currVal, err = Regen()
	case RUNE:
		currVal, err = Thorchain()
	case SEI:
		log.Println("Attempting to calculate Sei Nakamoto coefficient...")
		currVal, err = Sei()
		if err != nil {
			log.Printf("Error calculating Sei Nakamoto coefficient: %v", err)
		}
	case SOL:
		currVal, err = Solana()
	case STARS:
		log.Println("Attempting to calculate Stargaze Nakamoto coefficient...")
		currVal, err = Stargaze()
		if err != nil {
			log.Printf("Error calculating Stargaze Nakamoto coefficient: %v", err)
		}
	case STORY:
		currVal, err = Story()
	case SUI:
		currVal, err = Sui()
	case TIA:
		currVal, err = Celestia()
	case XNO:
		currVal, err = Nano()
	default:
		return 0, fmt.Errorf("chain not found: %s", token)
	}

	if err != nil {
		log.Printf("Error in chain %s: %v", token.ChainName(), err)
	} else {
		log.Printf("Successfully calculated Nakamoto coefficient for %s: %d", token.ChainName(), currVal)
	}

	return currVal, err
}
