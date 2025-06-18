package v2

import (
	"encoding/hex"

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
		// decode loan to v1
		var loanV1 types.LoanV1
		cdc.MustUnmarshal(iterator.Value(), &loanV1)

		// build new loan
		loan := &types.Loan{
			VaultAddress:              loanV1.VaultAddress,
			Borrower:                  loanV1.Borrower,
			BorrowerPubKey:            loanV1.BorrowerPubKey,
			BorrowerAuthPubKey:        loanV1.BorrowerPubKey,
			DCM:                       loanV1.DCM,
			MaturityTime:              loanV1.MaturityTime,
			FinalTimeout:              loanV1.FinalTimeout,
			PoolId:                    loanV1.PoolId,
			BorrowAmount:              loanV1.BorrowAmount,
			RequestFee:                loanV1.RequestFee,
			OriginationFee:            loanV1.OriginationFee,
			Interest:                  loanV1.Interest,
			ProtocolFee:               loanV1.ProtocolFee,
			Maturity:                  loanV1.Maturity,
			BorrowAPR:                 loanV1.BorrowAPR,
			MinMaturity:               loanV1.MinMaturity,
			StartBorrowIndex:          loanV1.StartBorrowIndex,
			LiquidationPrice:          loanV1.LiquidationPrice,
			LiquidationEventId:        loanV1.LiquidationEventId,
			DefaultLiquidationEventId: loanV1.DefaultLiquidationEventId,
			RepaymentEventId:          loanV1.RepaymentEventId,
			Authorizations:            loanV1.Authorizations,
			CollateralAmount:          loanV1.CollateralAmount,
			LiquidationId:             loanV1.LiquidationId,
			Referrer:                  loanV1.Referrer,
			CreateAt:                  loanV1.CreateAt,
			DisburseAt:                loanV1.DisburseAt,
			Status:                    loanV1.Status,
		}

		// update loan
		store.Set(types.LoanKey(loan.VaultAddress), cdc.MustMarshal(loan))

		// decode dlc meta to v1
		dlcMetaBz := store.Get(types.DLCMetaKey(loan.VaultAddress))
		var dlcMetaV1 types.DLCMetaV1
		cdc.MustUnmarshal(dlcMetaBz, &dlcMetaV1)

		internalKey, _ := hex.DecodeString(dlcMetaV1.InternalKey)

		// build new dlc meta
		dlcMeta := &types.DLCMeta{
			LiquidationCet:        dlcMetaV1.LiquidationCet,
			DefaultLiquidationCet: dlcMetaV1.DefaultLiquidationCet,
			RepaymentCet:          dlcMetaV1.RepaymentCet,
			TimeoutRefundTx:       dlcMetaV1.TimeoutRefundTx,
			VaultUtxos:            dlcMetaV1.VaultUtxos,
			InternalKey:           hex.EncodeToString(internalKey[1:]),
			LiquidationScript:     dlcMetaV1.MultisigScript,
			RepaymentScript:       dlcMetaV1.MultisigScript,
			TimeoutRefundScript:   dlcMetaV1.TimeoutRefundScript,
		}

		// update dlc meta
		store.Set(types.DLCMetaKey(loan.VaultAddress), cdc.MustMarshal(dlcMeta))
	}
}
