package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// SetPool sets the given pool
func (k Keeper) SetPool(ctx sdk.Context, pool *types.LendingPool) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(pool)

	store.Set(types.PoolKey(pool.Id), bz)
}

// HasPool returns true if the given pool exists, false otherwise
func (k Keeper) HasPool(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.PoolKey(id))
}

// GetPool gets the given pool
func (k Keeper) GetPool(ctx sdk.Context, id string) *types.LendingPool {
	store := ctx.KVStore(k.storeKey)

	var pool types.LendingPool
	bz := store.Get(types.PoolKey(id))
	k.cdc.MustUnmarshal(bz, &pool)

	return &pool
}

// GetPools gets pools by the given status
func (k Keeper) GetPools(ctx sdk.Context, status types.PoolStatus) []*types.LendingPool {
	var pools []*types.LendingPool

	k.IteratePools(ctx, func(pool *types.LendingPool) (stop bool) {
		if pool.Status == status {
			pools = append(pools, pool)
		}

		return false
	})

	return pools
}

// GetAllPools gets all pools
func (k Keeper) GetAllPools(ctx sdk.Context) []*types.LendingPool {
	var pools []*types.LendingPool

	k.IteratePools(ctx, func(pool *types.LendingPool) (stop bool) {
		pools = append(pools, pool)
		return false
	})

	return pools
}

// IteratePools iterates through all pools
func (k Keeper) IteratePools(ctx sdk.Context, cb func(pool *types.LendingPool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var pool types.LendingPool
		k.cdc.MustUnmarshal(iterator.Value(), &pool)

		if cb(&pool) {
			break
		}
	}
}

// AfterPoolBorrowed is the hook which is invoked after the loan is disbursed
func (k Keeper) AfterPoolBorrowed(ctx sdk.Context, poolId string, amount sdk.Coin) {
	pool := k.GetPool(ctx, poolId)

	pool.AvailableAmount = pool.AvailableAmount.Sub(amount.Amount)
	pool.TotalBorrowed = pool.TotalBorrowed.Add(amount.Amount)

	k.SetPool(ctx, pool)
}

// AfterPoolRepaid is the hook which is invoked after the loan is repaid
func (k Keeper) AfterPoolRepaid(ctx sdk.Context, poolId string, amount sdk.Coin, interest sdkmath.Int, protocolFee sdkmath.Int) {
	pool := k.GetPool(ctx, poolId)

	pool.Supply = pool.Supply.AddAmount(interest).SubAmount(protocolFee)
	pool.AvailableAmount = pool.AvailableAmount.Add(amount.Amount).Add(interest).Sub(protocolFee)
	pool.TotalBorrowed = pool.TotalBorrowed.Sub(amount.Amount)
	pool.TotalReserves = pool.TotalReserves.Add(protocolFee)

	k.SetPool(ctx, pool)
}

// GetSTokenAmount calculates the sToken amount from the given deposit amount
func (k Keeper) GetSTokenAmount(ctx sdk.Context, pool *types.LendingPool, depositAmount sdkmath.Int) sdkmath.Int {
	return depositAmount.Mul(pool.TotalSTokens.Amount).Quo(pool.AvailableAmount.Add(pool.TotalBorrowed))
}

// GetUnderlyingAssetAmount calculates the underlying asset amount from the given sToken amount
func (k Keeper) GetUnderlyingAssetAmount(ctx sdk.Context, pool *types.LendingPool, sTokenAmount sdkmath.Int) sdkmath.Int {
	return sTokenAmount.Mul(pool.AvailableAmount.Add(pool.TotalBorrowed)).Quo(pool.TotalSTokens.Amount)
}
