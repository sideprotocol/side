package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 2 to
// version 3
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
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
		// unmarshal loan to v2
		var loanV2 types.LoanV2
		cdc.MustUnmarshal(iterator.Value(), &loanV2)

		// build new loan
		loan := &types.Loan{
			VaultAddress:       loanV2.VaultAddress,
			Borrower:           loanV2.Borrower,
			BorrowerPubKey:     loanV2.BorrowerPubKey,
			BorrowerAuthPubKey: loanV2.BorrowerAuthPubKey,
			DCM:                loanV2.DCM,
			MaturityTime:       loanV2.MaturityTime,
			FinalTimeout:       loanV2.FinalTimeout,
			PoolId:             loanV2.PoolId,
			BorrowAmount:       loanV2.BorrowAmount,
			RequestFee:         loanV2.RequestFee,
			OriginationFee:     loanV2.OriginationFee,
			Interest:           loanV2.Interest,
			ProtocolFee:        loanV2.ProtocolFee,
			Maturity:           loanV2.Maturity,
			BorrowAPR:          loanV2.BorrowAPR,
			StartBorrowIndex:   loanV2.StartBorrowIndex,
			LiquidationPrice:   loanV2.LiquidationPrice,
			DlcEventId:         loanV2.DlcEventId,
			Authorizations:     loanV2.Authorizations,
			CollateralAmount:   loanV2.CollateralAmount,
			LiquidationId:      loanV2.LiquidationId,
			ReferralCode:       "",
			CreateAt:           loanV2.CreateAt,
			DisburseAt:         loanV2.DisburseAt,
			Status:             loanV2.Status,
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
		// unmarshal pool to v1
		var poolV1 types.LendingPoolV1
		cdc.MustUnmarshal(iterator.Value(), &poolV1)

		// build new pool
		pool := &types.LendingPool{
			Id:              poolV1.Id,
			Supply:          poolV1.Supply,
			AvailableAmount: poolV1.AvailableAmount,
			BorrowedAmount:  poolV1.BorrowedAmount,
			TotalBorrowed:   poolV1.TotalBorrowed,
			ReserveAmount:   poolV1.ReserveAmount,
			TotalReserve:    poolV1.TotalReserve,
			TotalYTokens:    poolV1.TotalYTokens,
			Tranches:        poolV1.Tranches,
			Config: types.PoolConfig{
				CollateralAsset:      poolV1.Config.CollateralAsset,
				LendingAsset:         poolV1.Config.LendingAsset,
				SupplyCap:            poolV1.Config.SupplyCap,
				BorrowCap:            poolV1.Config.BorrowCap,
				MinBorrowAmount:      poolV1.Config.MinBorrowAmount,
				MaxBorrowAmount:      poolV1.Config.MaxBorrowAmount,
				Tranches:             poolV1.Config.Tranches,
				RequestFee:           poolV1.Config.RequestFee,
				OriginationFeeFactor: 0,
				ReserveFactor:        poolV1.Config.ReserveFactor,
				MaxLtv:               poolV1.Config.MaxLtv,
				LiquidationThreshold: poolV1.Config.LiquidationThreshold,
				Paused:               poolV1.Config.Paused,
			},
			Status: poolV1.Status,
		}

		// update pool
		store.Set(iterator.Key(), cdc.MustMarshal(pool))
	}
}
