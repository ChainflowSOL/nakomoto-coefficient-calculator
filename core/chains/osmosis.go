package chains

func Osmosis() (int, error) {
	validatorURL := "https://osmosis-api.polkachu.com/cosmos/staking/v1beta1/validators?pagination.limit=500&status=BOND_STATUS_BONDED"
	stakingPoolURL := "https://osmosis-api.polkachu.com/cosmos/staking/v1beta1/pool"

	return FetchCosmosSDKNakaCoeff("osmosis", validatorURL, stakingPoolURL)
}
