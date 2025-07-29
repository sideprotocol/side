package v4

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 3 to
// version 4
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateLoans(ctx, storeKey, cdc)

	return nil
}

// migrateLoans performs the loan migration
func migrateLoans(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// unmarshal loan to v3
		var loanV3 types.LoanV3
		cdc.MustUnmarshal(iterator.Value(), &loanV3)

		var referrer *types.Referrer
		bz := store.Get(types.ReferrerKey(loanV3.ReferralCode))
		if bz != nil {
			var r types.Referrer
			cdc.MustUnmarshal(bz, &r)
			referrer = &r
		}

		// build new loan
		loan := &types.LoanV4{
			VaultAddress:       loanV3.VaultAddress,
			Borrower:           loanV3.Borrower,
			BorrowerPubKey:     loanV3.BorrowerPubKey,
			BorrowerAuthPubKey: loanV3.BorrowerAuthPubKey,
			DCM:                loanV3.DCM,
			MaturityTime:       loanV3.MaturityTime,
			FinalTimeout:       loanV3.FinalTimeout,
			PoolId:             loanV3.PoolId,
			BorrowAmount:       loanV3.BorrowAmount,
			RequestFee:         loanV3.RequestFee,
			OriginationFee:     loanV3.OriginationFee,
			Interest:           loanV3.Interest,
			ProtocolFee:        loanV3.ProtocolFee,
			Maturity:           loanV3.Maturity,
			BorrowAPR:          loanV3.BorrowAPR,
			StartBorrowIndex:   loanV3.StartBorrowIndex,
			LiquidationPrice:   loanV3.LiquidationPrice,
			DlcEventId:         loanV3.DlcEventId,
			Authorizations:     loanV3.Authorizations,
			CollateralAmount:   loanV3.CollateralAmount,
			LiquidationId:      loanV3.LiquidationId,
			Referrer:           referrer,
			CreateAt:           loanV3.CreateAt,
			DisburseAt:         loanV3.DisburseAt,
			Status:             loanV3.Status,
		}

		// update loan
		store.Set(iterator.Key(), cdc.MustMarshal(loan))
	}
}
