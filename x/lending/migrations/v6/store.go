package v6

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 5 to
// version 6
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, dlcKeeper types.DLCKeeper, cdc codec.BinaryCodec) error {
	migrateLoans(ctx, storeKey, cdc)
	migratePools(ctx, storeKey, cdc)

	return nil
}

// migrateLoans performs the loan migration
func migrateLoans(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// unmarshal loan to v4
		var loanV4 types.LoanV4
		cdc.MustUnmarshal(iterator.Value(), &loanV4)

		// build new loan
		loan := &types.Loan{
			VaultAddress:       loanV4.VaultAddress,
			Borrower:           loanV4.Borrower,
			BorrowerPubKey:     loanV4.BorrowerPubKey,
			BorrowerAuthPubKey: loanV4.BorrowerAuthPubKey,
			DCM:                loanV4.DCM,
			MaturityTime:       loanV4.MaturityTime,
			FinalTimeout:       loanV4.FinalTimeout,
			PoolId:             loanV4.PoolId,
			BorrowAmount:       loanV4.BorrowAmount,
			RequestFee:         loanV4.RequestFee,
			OriginationFee:     loanV4.OriginationFee,
			Interest:           loanV4.Interest,
			ProtocolFee:        loanV4.ProtocolFee,
			Maturity:           loanV4.Maturity,
			BorrowAPR:          sdkmath.LegacyNewDec(int64(loanV4.BorrowAPR)).QuoInt64(1000),
			StartBorrowIndex:   loanV4.StartBorrowIndex,
			LiquidationPrice:   loanV4.LiquidationPrice,
			DlcEventId:         loanV4.DlcEventId,
			Authorizations:     loanV4.Authorizations,
			CollateralAmount:   loanV4.CollateralAmount,
			LiquidationId:      loanV4.LiquidationId,
			Referrer:           loanV4.Referrer,
			CreateAt:           loanV4.CreateAt,
			DisburseAt:         loanV4.DisburseAt,
			Status:             loanV4.Status,
		}

		// update loan
		store.Set(iterator.Key(), cdc.MustMarshal(loan))
	}
}

// migratePools performs the pool migration
func migratePools(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.PoolKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// unmarshal pool to v2
		var poolV2 types.LendingPoolV2
		cdc.MustUnmarshal(iterator.Value(), &poolV2)

		// build new pool
		pool := &types.LendingPool{
			Id:              poolV2.Id,
			Supply:          poolV2.Supply,
			AvailableAmount: poolV2.AvailableAmount,
			BorrowedAmount:  poolV2.BorrowedAmount,
			TotalBorrowed:   poolV2.TotalBorrowed,
			ReserveAmount:   poolV2.ReserveAmount,
			TotalReserve:    poolV2.TotalReserve,
			TotalYTokens:    poolV2.TotalYTokens,
			Tranches:        poolV2.Tranches,
			Config: types.PoolConfig{
				CollateralAsset:      poolV2.Config.CollateralAsset,
				LendingAsset:         poolV2.Config.LendingAsset,
				SupplyCap:            poolV2.Config.SupplyCap,
				BorrowCap:            poolV2.Config.BorrowCap,
				MinBorrowAmount:      poolV2.Config.MinBorrowAmount,
				MaxBorrowAmount:      poolV2.Config.MaxBorrowAmount,
				RequestFee:           poolV2.Config.RequestFee,
				OriginationFeeFactor: sdkmath.LegacyNewDec(int64(poolV2.Config.OriginationFeeFactor)).QuoInt64(1000),
				ReserveFactor:        sdkmath.LegacyNewDec(int64(poolV2.Config.ReserveFactor)).QuoInt64(1000),
				MaxLtv:               sdkmath.LegacyNewDec(int64(poolV2.Config.MaxLtv)).QuoInt64(100),
				LiquidationThreshold: sdkmath.LegacyNewDec(int64(poolV2.Config.LiquidationThreshold)).QuoInt64(100),
				Paused:               poolV2.Config.Paused,
			},
			Status: poolV2.Status,
		}

		for _, tranche := range poolV2.Config.Tranches {
			pool.Config.Tranches = append(pool.Config.Tranches, types.PoolTrancheConfig{
				Maturity:  tranche.Maturity,
				BorrowAPR: sdkmath.LegacyNewDec(int64(tranche.BorrowAPR)).QuoInt64(1000),
			})
		}

		// update pool
		store.Set(iterator.Key(), cdc.MustMarshal(pool))
	}
}
