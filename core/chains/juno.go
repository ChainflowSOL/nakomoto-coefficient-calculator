package chains

func Juno() (int, error) {
	validatorsURL := "https://juno-api.polkachu.com/cosmos/staking/v1beta1/validators?pagination.limit=500&status=BOND_STATUS_BONDED"
	stakingPoolURL := "https://juno-api.polkachu.com/cosmos/staking/v1beta1/pool"

	return FetchCosmosSDKNakaCoeff("juno", validatorsURL, stakingPoolURL)
}
