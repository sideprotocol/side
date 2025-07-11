package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

	store.Set(types.StakingKey(staking.Id), bz)
}

// SetStakingByAddress sets the staking by the given address
func (k Keeper) SetStakingByAddress(ctx sdk.Context, address string, staking *types.Staking) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.StakingByAddressKey(address, staking.Id), []byte{})
}

// HasTotalStakings returns true if total staking stats exists by the given denom, false otherwise
func (k Keeper) HasTotalStakings(ctx sdk.Context, denom string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.TotalStakingsKey(denom))
}

// GetTotalStakings gets total staking stats by the given denom
func (k Keeper) GetTotalStakings(ctx sdk.Context, denom string) *types.TotalStakings {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.TotalStakingsKey(denom))
	var totalStakings types.TotalStakings
	k.cdc.MustUnmarshal(bz, &totalStakings)

	return &totalStakings
}

// SetTotalStakings sets total staking stats
func (k Keeper) SetTotalStakings(ctx sdk.Context, totalStakings *types.TotalStakings) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(totalStakings)

	store.Set(types.TotalStakingsKey(totalStakings.Denom), bz)
}

// IncreaseTotalStakings increases total staking stats according to the given staking
func (k Keeper) IncreaseTotalStakings(ctx sdk.Context, staking *types.Staking) {
	totalStakings := k.GetTotalStakings(ctx, staking.Amount.Denom)

	totalStakings.Amount = staking.Amount.AddAmount(totalStakings.Amount.Amount)
	totalStakings.EffectiveAmount = staking.EffectiveAmount.AddAmount(totalStakings.EffectiveAmount.Amount)

	k.SetTotalStakings(ctx, totalStakings)
}

// DecreaseTotalStakings decreases total staking stats according to the given staking
func (k Keeper) DecreaseTotalStakings(ctx sdk.Context, staking *types.Staking) {
	totalStakings := k.GetTotalStakings(ctx, staking.Amount.Denom)

	totalStakings.Amount = totalStakings.Amount.Sub(staking.Amount)
	totalStakings.EffectiveAmount = totalStakings.EffectiveAmount.Sub(staking.EffectiveAmount)

	k.SetTotalStakings(ctx, totalStakings)
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

// GetStakingsByAddress gets stakings according to the specified address
func (k Keeper) GetStakingsByAddress(ctx sdk.Context, address string) []*types.Staking {
	stakings := make([]*types.Staking, 0)

	k.IterateStakingsByAddress(ctx, address, func(staking *types.Staking) (stop bool) {
		stakings = append(stakings, staking)
		return false
	})

	return stakings
}

// IterateStakignsByAddress iterates through stakings by the given address
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
