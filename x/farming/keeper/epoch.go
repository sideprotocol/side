package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

// GetEpochId gets the current epoch id
func (k Keeper) GetEpochId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementEpochId increments the epoch id
func (k Keeper) IncrementEpochId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetEpochId(ctx) + 1
	store.Set(types.EpochIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasEpoch returns true if the given epoch exists, false otherwise
func (k Keeper) HasEpoch(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.EpochKey(id))
}

// GetEpoch gets the epoch by the given id
func (k Keeper) GetEpoch(ctx sdk.Context, id uint64) *types.Epoch {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochKey(id))
	var epoch types.Epoch
	k.cdc.MustUnmarshal(bz, &epoch)

	return &epoch
}

// SetEpoch sets the given epoch
func (k Keeper) SetEpoch(ctx sdk.Context, epoch *types.Epoch) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(epoch)

	store.Set(types.EpochKey(epoch.Id), bz)
}

// GetCurrentEpoch gets the current epoch
func (k Keeper) GetCurrentEpoch(ctx sdk.Context) *types.Epoch {
	id := k.GetEpochId(ctx)
	if id == 0 {
		return nil
	}

	return k.GetEpoch(ctx, id)
}

// NewEpoch creates a new epoch
func (k Keeper) NewEpoch(ctx sdk.Context) {
	epoch := &types.Epoch{
		Id:        k.IncrementEpochId(ctx),
		StartTime: ctx.BlockTime(),
		EndTime:   ctx.BlockTime().Add(k.EpochDuration(ctx)),
		Status:    types.EpochStatus_EPOCH_STATUS_STARTED,
	}

	k.SetEpoch(ctx, epoch)
}
