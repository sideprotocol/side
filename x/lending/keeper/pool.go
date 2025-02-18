package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

func (k Keeper) SetPool(ctx sdk.Context, pool types.LendingPool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&pool)
	store.Set(types.PoolStoreKey(pool.Id), bz)
}

func (k Keeper) HasPool(ctx sdk.Context, pool_id string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.PoolStoreKey(pool_id))
}

func (k Keeper) GetPool(ctx sdk.Context, pool_id string) types.LendingPool {
	store := ctx.KVStore(k.storeKey)
	var pool types.LendingPool
	bz := store.Get(types.PoolStoreKey(pool_id))
	k.cdc.MustUnmarshal(bz, &pool)
	return pool
}

// GetAllPools returns all block headers
func (k Keeper) GetAllPools(ctx sdk.Context) []*types.LendingPool {
	var pools []*types.LendingPool
	k.IteratePools(ctx, func(pool types.LendingPool) (stop bool) {
		pools = append(pools, &pool)
		return false
	})
	return pools
}

// IteratePools iterates through all block headers
func (k Keeper) IteratePools(ctx sdk.Context, process func(header types.LendingPool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.PoolStorePrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.LendingPool
		k.cdc.MustUnmarshal(iterator.Value(), &header)
		if process(header) {
			break
		}
	}
}
