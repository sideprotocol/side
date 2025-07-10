package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 1 to
// version 2
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
		// unmarshal loan to v1
		var loanV1 types.LoanV1
		cdc.MustUnmarshal(iterator.Value(), &loanV1)

		// build new loan
		loan := &types.LoanV2{
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
			Referrer:           loanV1.Referrer,
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
