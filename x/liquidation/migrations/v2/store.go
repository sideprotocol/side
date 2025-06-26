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

// migrateLiquidations performs the liquidation migration
func migrateLiquidations(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	liquidationIterator := storetypes.KVStorePrefixIterator(store, types.LiquidationKeyPrefix)
	defer liquidationIterator.Close()

	for ; liquidationIterator.Valid(); liquidationIterator.Next() {
		// delete liquidation
		store.Delete(liquidationIterator.Key())
	}

	liquidationRecordIterator := storetypes.KVStorePrefixIterator(store, types.LiquidationRecordKeyPrefix)
	defer liquidationRecordIterator.Close()

	for ; liquidationRecordIterator.Valid(); liquidationRecordIterator.Next() {
		// delete liquidation record
		store.Delete(liquidationRecordIterator.Key())
	}

	recordByLiquidationIterator := storetypes.KVStorePrefixIterator(store, types.LiquidationRecordByLiquidationKeyPrefix)
	defer recordByLiquidationIterator.Close()

	for ; recordByLiquidationIterator.Valid(); recordByLiquidationIterator.Next() {
		// delete liquidation record by liquidation
		store.Delete(recordByLiquidationIterator.Key())
	}
}
