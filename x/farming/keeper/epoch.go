package keeper

import (
	storetypes "cosmossdk.io/store/types"
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

// AddToCurrentEpochStakingQueue adds the given staking to the staking queue for the current epoch
func (k Keeper) AddToCurrentEpochStakingQueue(ctx sdk.Context, stakingId uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.CurrentEpochStakingQueueKey(stakingId), []byte{})
}

// RemoveFromCurrentEpochStakingQueue removes the given staking from the staking queue for the current epoch
func (k Keeper) RemoveFromCurrentEpochStakingQueue(ctx sdk.Context, stakingId uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.CurrentEpochStakingQueueKey(stakingId))
}

// HasStakingForCurrentEpoch returns true if the given staking exists for the current epoch, false otherwise
func (k Keeper) HasStakingForCurrentEpoch(ctx sdk.Context, stakingId uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.CurrentEpochStakingQueueKey(stakingId))
}

// IterateCurrentEpochStakingQueue iterates through the staking queue for the current epoch
func (k Keeper) IterateCurrentEpochStakingQueue(ctx sdk.Context, cb func(staking *types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.CurrentEpochStakingQueueKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		stakingId := sdk.BigEndianToUint64(iterator.Key()[1:])
		staking := k.GetStaking(ctx, stakingId)

		if cb(staking) {
			break
		}
	}
}

// NewEpoch creates a new epoch
func (k Keeper) NewEpoch(ctx sdk.Context) {
	epoch := &types.Epoch{
		Id:        k.IncrementEpochId(ctx),
		StartTime: ctx.BlockTime(),
		EndTime:   ctx.BlockTime().Add(k.EpochDuration(ctx)),
		Status:    types.EpochStatus_EPOCH_STATUS_STARTED,
	}

	// set the new epoch
	k.SetEpoch(ctx, epoch)

	// call handler on the new epoch started
	k.OnEpochStarted(ctx)
}

// OnEpochStarted is called when the current epoch is started
func (k Keeper) OnEpochStarted(ctx sdk.Context) {
	// get the current epoch
	currentEpoch := k.GetCurrentEpoch(ctx)

	// get all stakings
	stakings := k.GetAllStakings(ctx)

	for _, staking := range stakings {
		if staking.Status != types.StakingStatus_STAKING_STATUS_STAKED {
			continue
		}

		if staking.StartTime.Add(staking.LockDuration).Before(currentEpoch.EndTime) {
			continue
		}

		// add to staking queue for the current epoch
		k.AddToCurrentEpochStakingQueue(ctx, staking.Id)

		// update total staking for the current epoch
		types.UpdateEpochTotalStaking(currentEpoch, staking)
	}

	// update the current epoch
	k.SetEpoch(ctx, currentEpoch)
}

// OnEpochEnded is called when the current epoch ends
func (k Keeper) OnEpochEnded(ctx sdk.Context) {
	k.IterateCurrentEpochStakingQueue(ctx, func(staking *types.Staking) (stop bool) {
		// calculate the pending reward
		pendingReward := k.GetPendingReward(ctx, staking)

		// distribute reward
		staking.PendingRewards = staking.PendingRewards.Add(pendingReward)
		k.SetStaking(ctx, staking)

		// remove from the staking queue for the current epoch
		k.RemoveFromCurrentEpochStakingQueue(ctx, staking.Id)

		return false
	})
}
