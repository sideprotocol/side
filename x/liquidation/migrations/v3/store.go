package v3

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// MigrateStore migrates the x/liquidation module state from the consensus version 2 to
// version 3
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)

	return nil
}

// migrateParams performs the params migration
func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get current params
	var paramsV1 types.ParamsV1
	bz := store.Get(types.ParamsKey)
	cdc.MustUnmarshal(bz, &paramsV1)

	// build new params
	params := &types.Params{
		MinLiquidationFactor:            sdkmath.LegacyNewDec(int64(paramsV1.MinLiquidationFactor)).QuoInt64(1000),
		LiquidationBonusFactor:          sdkmath.LegacyNewDec(int64(paramsV1.LiquidationBonusFactor)).QuoInt64(1000),
		ProtocolLiquidationFeeFactor:    sdkmath.LegacyNewDec(int64(paramsV1.ProtocolLiquidationFeeFactor)).QuoInt64(1000),
		ProtocolLiquidationFeeCollector: paramsV1.ProtocolLiquidationFeeCollector,
	}

	// update params
	bz = cdc.MustMarshal(params)
	store.Set(types.ParamsKey, bz)
}
