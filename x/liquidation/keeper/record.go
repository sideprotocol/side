package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// GetLiquidationRecordId gets the current liquidation record id
func (k Keeper) GetLiquidationRecordId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.LiquidationRecordIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementLiquidationRecordId increments the liquidation record id and returns the new id
func (k Keeper) IncrementLiquidationRecordId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetLiquidationRecordId(ctx) + 1
	store.Set(types.LiquidationRecordIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasLiquidationRecord returns true if the given liquidation record exists, false otherwise
func (k Keeper) HasLiquidationRecord(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.LiquidationRecordKey(id))
}

// GetLiquidationRecord gets the liquidation record by the given id
func (k Keeper) GetLiquidationRecord(ctx sdk.Context, id uint64) *types.LiquidationRecord {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.LiquidationRecordKey(id))
	var record types.LiquidationRecord
	k.cdc.MustUnmarshal(bz, &record)

	return &record
}

// SetLiquidationRecord sets the given liquidation record
func (k Keeper) SetLiquidationRecord(ctx sdk.Context, record *types.LiquidationRecord) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(record)
	store.Set(types.LiquidationRecordKey(record.Id), bz)

	store.Set(types.LiquidationRecordByLiquidationKey(record.LiquidationId, record.Id), []byte{})
}

// GetAllLiquidationRecords gets all liquidation records
func (k Keeper) GetAllLiquidationRecords(ctx sdk.Context) []*types.LiquidationRecord {
	records := make([]*types.LiquidationRecord, 0)

	k.IterateLiquidationRecords(ctx, func(record *types.LiquidationRecord) (stop bool) {
		records = append(records, record)
		return false
	})

	return records
}

// GetLiquidationRecords gets liquidation records of the given liquidation
func (k Keeper) GetLiquidationRecords(ctx sdk.Context, liquidationId uint64) []*types.LiquidationRecord {
	records := make([]*types.LiquidationRecord, 0)

	k.IterateLiquidationRecordsByLiquidation(ctx, liquidationId, func(record *types.LiquidationRecord) (stop bool) {
		records = append(records, record)
		return false
	})

	return records
}

// IterateLiquidationRecords iterates through all liquidation records
func (k Keeper) IterateLiquidationRecords(ctx sdk.Context, cb func(record *types.LiquidationRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidationRecordKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var record types.LiquidationRecord
		k.cdc.MustUnmarshal(iterator.Value(), &record)

		if cb(&record) {
			break
		}
	}
}

// IterateBidsByAuction iterates through bids by the specified auction and status
func (k Keeper) IterateLiquidationRecordsByLiquidation(ctx sdk.Context, liquidationId uint64, cb func(record *types.LiquidationRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.LiquidationRecordByLiquidationKeyPrefix, sdk.Uint64ToBigEndian(liquidationId)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		record := k.GetLiquidationRecord(ctx, sdk.BigEndianToUint64(key[1+8:]))

		if cb(record) {
			break
		}
	}
}
