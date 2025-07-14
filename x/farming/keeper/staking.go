package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sideprotocol/side/x/farming/types"
)

// GetStakingId gets the current staking id
func (k Keeper) GetStakingId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.StakingIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementStakingId increments the staking id
func (k Keeper) IncrementStakingId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetStakingId(ctx) + 1
	store.Set(types.StakingIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasStaking returns true if the given staking exists, false otherwise
func (k Keeper) HasStaking(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.StakingKey(id))
}

// GetStaking gets the staking by the given id
func (k Keeper) GetStaking(ctx sdk.Context, id uint64) *types.Staking {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.StakingKey(id))
	var staking types.Staking
	k.cdc.MustUnmarshal(bz, &staking)

	return &staking
}

// SetStaking sets the given staking
func (k Keeper) SetStaking(ctx sdk.Context, staking *types.Staking) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(staking)

	k.SetStakingStatus(ctx, staking.Id, staking.Status)

	store.Set(types.StakingKey(staking.Id), bz)
}

// SetStakingByAddress sets the staking by the given address
func (k Keeper) SetStakingByAddress(ctx sdk.Context, address string, staking *types.Staking) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.StakingByAddressKey(address, staking.Id), []byte{})
}

// SetStakingStatus sets the status store of the given staking
func (k Keeper) SetStakingStatus(ctx sdk.Context, id uint64, status types.StakingStatus) {
	store := ctx.KVStore(k.storeKey)

	if k.HasStaking(ctx, id) {
		k.RemoveStakingStatus(ctx, id)
	}

	store.Set(types.StakingByStatusKey(status, id), []byte{})
}

// RemoveStakingStatus removes the status store of the given staking
func (k Keeper) RemoveStakingStatus(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)

	staking := k.GetStaking(ctx, id)

	store.Delete(types.StakingByStatusKey(staking.Status, id))
}

// GetAllStakings gets all stakings
func (k Keeper) GetAllStakings(ctx sdk.Context) []*types.Staking {
	stakings := make([]*types.Staking, 0)

	k.IterateStakings(ctx, func(staking *types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// GetStakingsByStatus gets stakings by the given status
func (k Keeper) GetStakingsByStatus(ctx sdk.Context, status types.StakingStatus) []*types.Staking {
	stakings := make([]*types.Staking, 0)

	k.IterateStakingsByStatus(ctx, status, func(staking *types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// GetStakingsByStatusWithPagination gets stakings by the given status with pagination
func (k Keeper) GetStakingsByStatusWithPagination(ctx sdk.Context, status types.StakingStatus, pagination *query.PageRequest) ([]*types.Staking, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	stakingStatusStore := prefix.NewStore(store, append(types.StakingByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...))

	var stakings []*types.Staking

	pageRes, err := query.Paginate(stakingStatusStore, pagination, func(key []byte, value []byte) error {
		id := sdk.BigEndianToUint64(key)
		staking := k.GetStaking(ctx, id)

		stakings = append(stakings, staking)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return stakings, pageRes, nil
}

// GetStakingsByAddress gets stakings according to the specified address
func (k Keeper) GetStakingsByAddress(ctx sdk.Context, address string) []*types.Staking {
	stakings := make([]*types.Staking, 0)

	k.IterateStakingsByAddress(ctx, address, func(staking *types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// IterateStakings iterates through all stakings
func (k Keeper) IterateStakings(ctx sdk.Context, cb func(staking *types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.StakingKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var staking types.Staking
		k.cdc.MustUnmarshal(iterator.Value(), &staking)

		if cb(&staking) {
			break
		}
	}
}

// IterateStakingsByStatus iterates through stakings by the given status
func (k Keeper) IterateStakingsByStatus(ctx sdk.Context, status types.StakingStatus, cb func(staking *types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	keyPrefix := append(types.StakingByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		id := sdk.BigEndianToUint64(key[len(keyPrefix):])
		staking := k.GetStaking(ctx, id)

		if cb(staking) {
			break
		}
	}
}

// IterateStakingsByAddress iterates through stakings by the given address
func (k Keeper) IterateStakingsByAddress(ctx sdk.Context, address string, cb func(staking *types.Staking) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	keyPrefix := append(types.StakingByAddressKeyPrefix, []byte(address)...)

	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		id := sdk.BigEndianToUint64(key[len(keyPrefix):])
		staking := k.GetStaking(ctx, id)

		if cb(staking) {
			break
		}
	}
}

// HasTotalStaking returns true if total staking exists for the given denom, false otherwise
func (k Keeper) HasTotalStaking(ctx sdk.Context, denom string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.TotalStakingKey(denom))
}

// GetTotalStaking gets total staking by the given phase and denom
func (k Keeper) GetTotalStaking(ctx sdk.Context, denom string) *types.TotalStaking {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TotalStakingKey(denom))
	var totalStaking types.TotalStaking
	k.cdc.MustUnmarshal(bz, &totalStaking)

	return &totalStaking
}

// SetTotalStaking sets total staking
func (k Keeper) SetTotalStaking(ctx sdk.Context, totalStaking *types.TotalStaking) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(totalStaking)

	store.Set(types.TotalStakingKey(totalStaking.Denom), bz)
}

// IncreaseTotalStaking increases total staking according to the given staking
func (k Keeper) IncreaseTotalStaking(ctx sdk.Context, staking *types.Staking) {
	totalStaking := &types.TotalStaking{}
	if !k.HasTotalStaking(ctx, staking.Amount.Denom) {
		totalStaking.Denom = staking.Amount.Denom
		totalStaking.Amount = sdk.NewInt64Coin(staking.Amount.Denom, 0)
		totalStaking.EffectiveAmount = sdk.NewInt64Coin(staking.Amount.Denom, 0)
	} else {
		totalStaking = k.GetTotalStaking(ctx, staking.Amount.Denom)
	}

	totalStaking.Amount = totalStaking.Amount.Add(staking.Amount)
	totalStaking.EffectiveAmount = totalStaking.EffectiveAmount.Add(staking.EffectiveAmount)

	k.SetTotalStaking(ctx, totalStaking)
}

// DecreaseTotalStaking decreases total staking according to the given staking
func (k Keeper) DecreaseTotalStaking(ctx sdk.Context, staking *types.Staking) {
	totalStaking := k.GetTotalStaking(ctx, staking.Amount.Denom)

	totalStaking.Amount = totalStaking.Amount.Sub(staking.Amount)
	totalStaking.EffectiveAmount = totalStaking.EffectiveAmount.Sub(staking.EffectiveAmount)

	k.SetTotalStaking(ctx, totalStaking)
}
