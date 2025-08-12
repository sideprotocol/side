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
func (k Keeper) AfterPoolBorrowed(ctx sdk.Context, poolId string, maturity int64, amount sdk.Coin) {
	pool := k.GetPool(ctx, poolId)

	pool.AvailableAmount = pool.AvailableAmount.Sub(amount.Amount)
	pool.BorrowedAmount = pool.BorrowedAmount.Add(amount.Amount)
	pool.TotalBorrowed = pool.TotalBorrowed.Add(amount.Amount)

	for i, tranche := range pool.Tranches {
		if tranche.Maturity == maturity {
			pool.Tranches[i].TotalBorrowed = pool.Tranches[i].TotalBorrowed.Add(amount.Amount)
			break
		}
	}

	k.SetPool(ctx, pool)
}

// AfterPoolRepaid is the hook which is invoked after the loan is repaid
func (k Keeper) AfterPoolRepaid(ctx sdk.Context, poolId string, maturity int64, amount sdk.Coin, interest sdkmath.Int, protocolFee sdkmath.Int, actualProtocolFee sdkmath.Int) {
	pool := k.GetPool(ctx, poolId)

	repaidAmount := amount.Amount.Add(interest)
	actualRepaidAmount := repaidAmount.Sub(protocolFee)

	pool.Supply = pool.Supply.AddAmount(interest).SubAmount(protocolFee)
	pool.AvailableAmount = pool.AvailableAmount.Add(actualRepaidAmount)
	pool.BorrowedAmount = pool.BorrowedAmount.Sub(amount.Amount)
	pool.TotalBorrowed = pool.TotalBorrowed.Sub(repaidAmount)
	pool.ReserveAmount = pool.ReserveAmount.Add(actualProtocolFee)
	pool.TotalReserve = pool.TotalReserve.Sub(protocolFee)

	for i, tranche := range pool.Tranches {
		if tranche.Maturity == maturity {
			pool.Tranches[i].TotalBorrowed = pool.Tranches[i].TotalBorrowed.Sub(repaidAmount)
			break
		}
	}

	k.NormalizePool(ctx, pool)

	k.SetPool(ctx, pool)
}

// RebalancePool rebalances the given pool by means of decreasing total borrowed and total reserve by the respective amounts
// Note: This is used to adjust the pool after liquidation is completed
func (k Keeper) RebalancePool(ctx sdk.Context, poolId string, maturity int64, totalBorrowedDelta sdkmath.Int, totalReserveDelta sdkmath.Int) {
	pool := k.GetPool(ctx, poolId)

	pool.TotalBorrowed = pool.TotalBorrowed.Sub(totalBorrowedDelta)
	pool.TotalReserve = pool.TotalReserve.Sub(totalReserveDelta)

	for i, tranche := range pool.Tranches {
		if tranche.Maturity == maturity {
			pool.Tranches[i].TotalBorrowed = pool.Tranches[i].TotalBorrowed.Sub(totalBorrowedDelta)
			pool.Tranches[i].TotalReserve = pool.Tranches[i].TotalReserve.Sub(totalReserveDelta)
			break
		}
	}

	k.NormalizePool(ctx, pool)

	k.SetPool(ctx, pool)
}

// UpdatePoolTranches updates total borrowed amount for each tranche at the beginning of each block
//
// Formula:
//
// borrowRatePerBlock = borrowAPR / blocksPerYear
// borrowIndexRatio = 1 + borrowRatePerBlock
// borrowIndex_new = borrowIndex_old * borrowIndexRatio
// totalBorrowed_new = totalBorrowed_old * borrowIndexRatio
func (k Keeper) UpdatePoolTranches(ctx sdk.Context, pool *types.LendingPool) {
	// get blocks per year
	blocksPerYear := k.GetBlocksPerYear(ctx)

	for i, tranche := range pool.Tranches {
		trancheConfig, _ := types.GetTrancheConfig(pool.Config.Tranches, tranche.Maturity)

		borrowRatePerBlock := trancheConfig.BorrowAPR.Quo(sdkmath.LegacyNewDec(int64(blocksPerYear)))
		borrowIndexRatio := sdkmath.LegacyOneDec().Add(borrowRatePerBlock)

		reserveDelta := pool.Tranches[i].TotalBorrowed.ToLegacyDec().Mul(borrowRatePerBlock).Mul(pool.Config.ReserveFactor).TruncateInt()

		pool.Tranches[i].BorrowIndex = pool.Tranches[i].BorrowIndex.Mul(borrowIndexRatio)
		pool.Tranches[i].TotalBorrowed = pool.Tranches[i].TotalBorrowed.ToLegacyDec().Mul(borrowIndexRatio).TruncateInt()

		pool.Tranches[i].TotalReserve = pool.Tranches[i].TotalReserve.Add(reserveDelta)
	}
}

// UpdatePool updates total borrowed amount for the given pool at the beginning of each block
func (k Keeper) UpdatePool(ctx sdk.Context, pool *types.LendingPool) {
	// update all tranches
	k.UpdatePoolTranches(ctx, pool)

	// reset total borrowed and total reserve
	pool.TotalBorrowed = sdkmath.ZeroInt()
	pool.TotalReserve = sdkmath.ZeroInt()

	for _, tranche := range pool.Tranches {
		pool.TotalBorrowed = pool.TotalBorrowed.Add(tranche.TotalBorrowed)
		pool.TotalReserve = pool.TotalReserve.Add(tranche.TotalReserve)
	}
}

// UpdatePoolStatus updates the pool status with the given new config
func (k Keeper) UpdatePoolStatus(ctx sdk.Context, pool *types.LendingPool, newConfig *types.PoolConfig) {
	switch {
	case newConfig.Paused && !pool.Config.Paused:
		pool.Status = types.PoolStatus_PAUSED

	case !newConfig.Paused && pool.Config.Paused:
		if pool.Supply.IsZero() {
			pool.Status = types.PoolStatus_INACTIVE
		} else {
			pool.Status = types.PoolStatus_ACTIVE
		}

	default:
		return
	}
}

// OnPoolTranchesConfigChanged is called when the pool tranches config changes
func (k Keeper) OnPoolTranchesConfigChanged(ctx sdk.Context, pool *types.LendingPool, newTranchesConfig []types.PoolTrancheConfig) {
	for _, newConfig := range newTranchesConfig {
		_, found := types.GetTrancheConfig(pool.Config.Tranches, newConfig.Maturity)
		if !found {
			// add the new tranche to pool if the tranche does not exist
			pool.Tranches = append(pool.Tranches, types.NewTranche(newConfig))
		}
	}
}

// NormalizePool normalizes the given pool
func (k Keeper) NormalizePool(ctx sdk.Context, pool *types.LendingPool) {
	if pool.TotalBorrowed.IsNegative() {
		pool.TotalBorrowed = sdkmath.ZeroInt()
	}

	if pool.TotalReserve.IsNegative() {
		pool.TotalReserve = sdkmath.ZeroInt()
	}

	for i := range pool.Tranches {
		if pool.Tranches[i].TotalBorrowed.IsNegative() {
			pool.Tranches[i].TotalBorrowed = sdkmath.ZeroInt()
		}

		if pool.Tranches[i].TotalReserve.IsNegative() {
			pool.Tranches[i].TotalReserve = sdkmath.ZeroInt()
		}
	}
}

// GetYTokenAmount calculates the yToken amount from the given deposit amount
func (k Keeper) GetYTokenAmount(ctx sdk.Context, pool *types.LendingPool, depositAmount sdkmath.Int) sdkmath.Int {
	return depositAmount.Mul(pool.TotalYTokens.Amount).Quo(pool.AvailableAmount.Add(pool.TotalBorrowed).Sub(pool.TotalReserve))
}

// GetUnderlyingAssetAmount calculates the underlying asset amount from the given yToken amount
func (k Keeper) GetUnderlyingAssetAmount(ctx sdk.Context, pool *types.LendingPool, yTokenAmount sdkmath.Int) sdkmath.Int {
	return yTokenAmount.Mul(pool.AvailableAmount.Add(pool.TotalBorrowed).Sub(pool.TotalReserve)).Quo(pool.TotalYTokens.Amount)
}
