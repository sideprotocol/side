package v3

import (
	sdkmath "cosmossdk.io/math"
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
	registerReferrer(ctx, storeKey, cdc)

	return nil
}

// migrateLoans performs the loan migration
func migrateLoans(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// unmarshal loan to v1
		var loanV1 types.LoanV1
		cdc.MustUnmarshal(iterator.Value(), &loanV1)

		// build new loan
		loan := &types.Loan{
			VaultAddress:       loanV1.VaultAddress,
			Borrower:           loanV1.Borrower,
			BorrowerPubKey:     loanV1.BorrowerPubKey,
			BorrowerAuthPubKey: loanV1.BorrowerAuthPubKey,
			DCM:                loanV1.DCM,
			MaturityTime:       loanV1.MaturityTime,
			FinalTimeout:       loanV1.FinalTimeout,
			PoolId:             loanV1.PoolId,
			BorrowAmount:       loanV1.BorrowAmount,
			RequestFee:         loanV1.RequestFee,
			OriginationFee:     loanV1.OriginationFee,
			Interest:           loanV1.Interest,
			ProtocolFee:        loanV1.ProtocolFee,
			Maturity:           loanV1.Maturity,
			BorrowAPR:          loanV1.BorrowAPR,
			StartBorrowIndex:   loanV1.StartBorrowIndex,
			LiquidationPrice:   loanV1.LiquidationPrice,
			DlcEventId:         loanV1.DlcEventId,
			Authorizations:     loanV1.Authorizations,
			CollateralAmount:   loanV1.CollateralAmount,
			LiquidationId:      loanV1.LiquidationId,
			Referrer:           nil,
			CreateAt:           loanV1.CreateAt,
			DisburseAt:         loanV1.DisburseAt,
			Status:             loanV1.Status,
		}

		// update status
		if loan.Status >= types.LoanStatus(2) {
			loan.Status = loan.Status + 1
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
				OriginationFeeFactor: 1, // 0.1%
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

// registerReferrer performs the referrer pre-registration
func registerReferrer(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	referrer := &types.Referrer{
		Name:              "side",
		ReferralCode:      "SIDE1234",
		Address:           "tb1q37m9xfqscwral5km478geh4xzgsyw7nevmhaxx",
		ReferralFeeFactor: sdkmath.LegacyMustNewDecFromStr("0.2"),
	}

	store.Set(types.ReferrerKey(referrer.ReferralCode), cdc.MustMarshal(referrer))
}
