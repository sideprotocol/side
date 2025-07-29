package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 2 to
// version 3
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, oracleStoreKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)

	return nil
}

// migrateParams performs the params migration
func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get current params
	var paramsV2 types.ParamsV2
	bz := store.Get(types.ParamsStoreKey)
	cdc.MustUnmarshal(bz, &paramsV2)

	// build new params
	params := &types.Params{
		DepositConfirmationDepth:  types.DefaultDepositConfirmationDepth,
		WithdrawConfirmationDepth: types.DefaultWithdrawConfirmationDepth,
		MaxAcceptableBlockDepth:   paramsV2.MaxAcceptableBlockDepth,
		BtcVoucherDenom:           paramsV2.BtcVoucherDenom,
		DepositEnabled:            paramsV2.DepositEnabled,
		WithdrawEnabled:           paramsV2.WithdrawEnabled,
		TrustedNonBtcRelayers:     paramsV2.TrustedNonBtcRelayers,
		TrustedFeeProviders:       paramsV2.TrustedFeeProviders,
		FeeRateValidityPeriod:     paramsV2.FeeRateValidityPeriod,
		Vaults:                    paramsV2.Vaults,
		WithdrawParams:            paramsV2.WithdrawParams,
		ProtocolLimits:            paramsV2.ProtocolLimits,
		ProtocolFees:              paramsV2.ProtocolFees,
		TssParams:                 paramsV2.TssParams,
		IbcParams: types.IBCParams{
			TimeoutHeightOffset: types.DefaultIBCTimeoutHeightOffset,
			TimeoutDuration:     types.DefaultIBCTimeoutDuration,
		},
		RateLimitParams: paramsV2.RateLimitParams,
	}

	bz = cdc.MustMarshal(params)
	store.Set(types.ParamsStoreKey, bz)
}
