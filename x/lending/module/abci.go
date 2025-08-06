package lending

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
)

// BeginBlocker called at the beginning of each block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	updatePools(ctx, k)

	return nil
}

// EndBlocker called at the end of each block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// handle approvals for pending loans
	if err := handleApprovals(ctx, k); err != nil {
		return err
	}

	// handle liquidations for active loans
	handleLiquidations(ctx, k)

	// handle liquidated loans
	handleLiquidatedLoans(ctx, k)

	// handle repayments
	return handleRepayments(ctx, k)
}

// handleApprovals performs approvals for pending loans
func handleApprovals(ctx sdk.Context, k keeper.Keeper) error {
	var err error

	// requested loans
	k.IterateLoansByStatus(ctx, types.LoanStatus_Requested, func(loan *types.Loan) (stop bool) {
		err = k.HandleApproval(ctx, loan)
		return err != nil
	})

	if err != nil {
		return err
	}

	// authorized loans
	k.IterateLoansByStatus(ctx, types.LoanStatus_Authorized, func(loan *types.Loan) (stop bool) {
		err = k.HandleApproval(ctx, loan)
		return err != nil
	})

	return err
}

// handleLiquidations performs possible liquidations for active loans
func handleLiquidations(ctx sdk.Context, k keeper.Keeper) {
	k.IterateLoansByStatus(ctx, types.LoanStatus_Open, func(loan *types.Loan) (stop bool) {
		k.HandleLiquidation(ctx, loan)
		return false
	})
}

// handleLiquidatedLoans handles liquidated loans
func handleLiquidatedLoans(ctx sdk.Context, k keeper.Keeper) {
	k.IterateLiquidationQueue(ctx, func(loan *types.Loan) (stop bool) {
		// get dlc meta
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)

		// get liquidation cet and type
		cet, cetType := types.GetLiquidationCetAndType(dlcMeta, loan.Status)

		// check if the borrower adapted signatures already exist
		if len(cet.BorrowerAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.DlcEventId)
			if attestation == nil {
				return false
			}

			eventSignature, _ := hex.DecodeString(attestation.Signature)
			adaptorSecret := eventSignature[32:]

			// decrypt the adaptor signatures
			for _, adaptorSignature := range cet.BorrowerAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				cet.BorrowerAdaptedSignatures = append(
					cet.BorrowerAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed liquidation cet if both borrower adapted signatures(obviously exist) and DCM signatures already exist
		if len(cet.DCMSignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedCet(cet.Tx, loan.BorrowerAuthPubKey, cet.BorrowerAdaptedSignatures, loan.DCM, cet.DCMSignatures, cetType)
			if err != nil {
				k.Logger(ctx).Error("failed to build signed liquidation cet", "loan id", loan.VaultAddress, "err", err)
			} else {
				cet.SignedTxHex = hex.EncodeToString(signedTx)

				// remove from the liquidation queue
				k.RemoveFromLiquidationQueue(ctx, loan.VaultAddress)

				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeGenerateSignedCet,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
						sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", cetType)),
						sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
					),
				)
			}
		}

		// update liquidation cet
		types.UpdateLiquidationCet(dlcMeta, cetType, cet)

		// update dlc meta
		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

		return false
	})
}

// handleRepayments handles repayments
func handleRepayments(ctx sdk.Context, k keeper.Keeper) error {
	var unexpectedErr error

	k.IterateLoansByStatus(ctx, types.LoanStatus_Repaid, func(loan *types.Loan) (stop bool) {
		// trigger dlc event if not triggered yet
		if !k.DLCKeeper().GetEvent(ctx, loan.DlcEventId).HasTriggered {
			k.DLCKeeper().TriggerDLCEvent(ctx, loan.DlcEventId, types.RepaidOutcomeIndex)
			return false
		}

		// get dlc meta
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)

		// check if the DCM adapted signatures already exist
		if len(dlcMeta.RepaymentCet.DCMAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.DlcEventId)
			if attestation == nil {
				return false
			}

			eventSignature, _ := hex.DecodeString(attestation.Signature)
			adaptorSecret := eventSignature[32:]

			// decrypt the DCM adaptor signatures
			for _, adaptorSignature := range dlcMeta.RepaymentCet.DCMAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.RepaymentCet.DCMAdaptedSignatures = append(
					dlcMeta.RepaymentCet.DCMAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed repayment cet
		signedTx, txHash, err := types.BuildSignedCet(dlcMeta.RepaymentCet.Tx, loan.BorrowerPubKey, dlcMeta.RepaymentCet.BorrowerSignatures, loan.DCM, dlcMeta.RepaymentCet.DCMAdaptedSignatures, types.CetType_REPAYMENT)
		if err != nil {
			k.Logger(ctx).Error("failed to build signed repayment cet", "loan id", loan.VaultAddress, "err", err)
		} else {
			dlcMeta.RepaymentCet.SignedTxHex = hex.EncodeToString(signedTx)

			// complete repayment
			if err := k.CompleteRepayment(ctx, loan); err != nil {
				// unexpected error
				unexpectedErr = err
				k.Logger(ctx).Error("failed to complete repayment", "loan id", loan.VaultAddress, "err", err)

				return true
			}

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeGenerateSignedCet,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", types.CetType_REPAYMENT)),
					sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
				),
			)
		}

		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

		return false
	})

	return unexpectedErr
}

// updatePools updates all active pools at the beginning of each block
func updatePools(ctx sdk.Context, k keeper.Keeper) {
	// get all active pools
	pools := k.GetPools(ctx, types.PoolStatus_ACTIVE)

	for _, pool := range pools {
		k.UpdatePool(ctx, pool)

		k.SetPool(ctx, pool)
	}
}
