package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)

	return nil
}

// migrateParams migrates the params
func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get current params
	var paramsV1 types.ParamsV1
	bz := store.Get(types.ParamsStoreKey)
	cdc.MustUnmarshal(bz, &paramsV1)

	// build new params
	params := &types.Params{
		Confirmations:           paramsV1.Confirmations,
		MaxAcceptableBlockDepth: paramsV1.MaxAcceptableBlockDepth,
		BtcVoucherDenom:         paramsV1.BtcVoucherDenom,
		DepositEnabled:          paramsV1.DepositEnabled,
		WithdrawEnabled:         paramsV1.WithdrawEnabled,
		TrustedNonBtcRelayers:   paramsV1.TrustedNonBtcRelayers,
		Vaults:                  paramsV1.Vaults,
		WithdrawParams:          paramsV1.WithdrawParams,
		ProtocolLimits:          paramsV1.ProtocolLimits,
		ProtocolFees:            paramsV1.ProtocolFees,
		TssParams: types.TSSParams{
			DkgTimeoutPeriod:                  paramsV1.TssParams.DkgTimeoutPeriod,
			ParticipantUpdateTransitionPeriod: paramsV1.TssParams.ParticipantUpdateTransitionPeriod,
		},
	}

	params.TrustedFeeProviders = paramsV1.TrustedOracles
	params.TrustedBtcRelayers = []string{}

	bz = cdc.MustMarshal(params)
	store.Set(types.ParamsStoreKey, bz)
}
