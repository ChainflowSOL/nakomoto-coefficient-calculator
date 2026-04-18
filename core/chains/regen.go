package chains

func Regen() (int, error) {
	validatorURL := "https://rest.cosmos.directory/regen/cosmos/staking/v1beta1/validators?pagination.limit=500&status=BOND_STATUS_BONDED"
	poolURL := "https://rest.cosmos.directory/regen/cosmos/staking/v1beta1/pool"

	return FetchCosmosSDKNakaCoeff("regen", validatorURL, poolURL)
}
