package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// MigrateStore migrates the x/liquidation module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateLiquidations(ctx, storeKey, cdc)

	return nil
}

// migrateLiquidations performs the liquidations migration
func migrateLiquidations(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidationKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var liquidation types.Liquidation
		cdc.MustUnmarshal(iterator.Value(), &liquidation)

		// set liquidation by status
		store.Set(types.LiquidationByStatusKey(liquidation.Status, liquidation.Id), []byte{})
	}
}
