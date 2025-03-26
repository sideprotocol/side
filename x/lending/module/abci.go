package lending

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
)

// BeginBlocker called at the beginning of each block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	updatePools(ctx, k)
}

// EndBlocker called at the end of each block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleActiveLoans(ctx, k)

	handleLiquidatedLoans(ctx, k)
	handleDefaultedLoans(ctx, k)

	handleRepayments(ctx, k)
}

// handleActiveLoans handles active loans
func handleActiveLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all active loans
	loans := k.GetLoans(ctx, types.LoanStatus_Open)

	for _, loan := range loans {
		var liquidationCet string
		var triggeredEventId uint64

		currentPrice, err := k.GetPrice(ctx, "BTC-USD")
		if err != nil {
			k.Logger(ctx).Info("failed to get price", "err", err)
		}

		poolConfig := k.GetPool(ctx, loan.VaultAddress).Config
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)

		// check if the loan has defaulted
		if ctx.BlockTime().Unix() >= types.GetDefaultLiquidationDate(loan.MaturityTime) {
			loan.Status = types.LoanStatus_Defaulted

			liquidationCet = dlcMeta.DefaultLiquidationCet.Tx
			triggeredEventId = loan.DefaultLiquidationEventId

			// get default liquidation cet sig hashes; no error
			defaultLiquidationCetSigHashes, _ := types.GetDefaultLiquidationCetSigHashes(dlcMeta)

			// emit default event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeDefault,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
					sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(defaultLiquidationCetSigHashes, types.AttributeValueSeparator)),
				),
			)
		} else if !currentPrice.IsZero() {
			liquidationPrice := types.GetLiquidationPrice(loan.CollateralAmount, loan.BorrowAmount.Amount, loan.MaturityTime-loan.CreateAt.Unix(), poolConfig.BorrowAPR, poolConfig.LiquidationThreshold)

			// check if the loan is to be liquidated
			if currentPrice.LTE(liquidationPrice) {
				loan.Status = types.LoanStatus_Liquidated

				liquidationCet = dlcMeta.LiquidationCet.Tx
				triggeredEventId = loan.LiquidationEventId

				// get liquidation cet sig hashes; no error
				liquidationCetSigHashes, _ := types.GetLiquidationCetSigHashes(dlcMeta)

				// emit liquidation event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeLiquidate,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
						sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
						sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(liquidationCetSigHashes, types.AttributeValueSeparator)),
					),
				)
			}
		}

		// create liquidation if defaulted or liquidated
		if loan.Status == types.LoanStatus_Defaulted || loan.Status == types.LoanStatus_Liquidated {
			liquidation := k.LiquidationKeeper().CreateLiquidation(ctx, &liquidationtypes.Liquidation{
				LoanId:               loan.VaultAddress,
				Borrower:             loan.Borrower,
				DCM:                  loan.DCM,
				LiquidatedCollateral: sdk.NewCoin("sat", loan.CollateralAmount),
				LiquidatedPrice:      currentPrice.Int64(),
				LiquidatedTime:       ctx.BlockTime(),
				DebtAmount:           sdk.NewCoin(k.GetPool(ctx, loan.PoolId).Supply.Denom, loan.BorrowAmount.Amount.Add(loan.Interest)),
				LiquidationCet:       liquidationCet,
			})
			loan.LiquidationId = liquidation.Id

			// trigger dlc event if not triggered yet
			if !k.DLCKeeper().GetEvent(ctx, triggeredEventId).HasTriggered {
				k.DLCKeeper().TriggerDLCEvent(ctx, triggeredEventId, 0)
			}

			// update loan
			k.SetLoan(ctx, loan)
		}
	}
}

// handleLiquidatedLoans handles liquidated loans
func handleLiquidatedLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all liquidated loans
	loans := k.GetLoans(ctx, types.LoanStatus_Liquidated)

	for _, loan := range loans {
		// check if the liquidation cet has been signed
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.LiquidationCet.SignedTxHex) != 0 {
			continue
		}

		// check if the borrower adapted signatures already exist
		if len(dlcMeta.LiquidationCet.BorrowerAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.LiquidationEventId)
			if attestation == nil {
				continue
			}

			// decrypt the adaptor signatures
			for _, adaptorSignature := range dlcMeta.LiquidationCet.BorrowerAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptorSecret, _ := hex.DecodeString(attestation.Signature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.LiquidationCet.BorrowerAdaptedSignatures = append(
					dlcMeta.LiquidationCet.BorrowerAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed liquidation cet if both borrower adapted signatures(obviously exist) and DCM signatures already exist
		if len(dlcMeta.LiquidationCet.DCMSignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedCet(dlcMeta.LiquidationCet.Tx, loan.BorrowerPubKey, dlcMeta.LiquidationCet.BorrowerAdaptedSignatures, loan.DCM, dlcMeta.LiquidationCet.DCMSignatures)
			if err != nil {
				k.Logger(ctx).Info("failed to build signed liquidation cet", "loan id", loan.VaultAddress, "err", err)
			} else {
				dlcMeta.LiquidationCet.SignedTxHex = hex.EncodeToString(signedTx)

				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeGenerateSignedCet,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
						sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", types.CetType_LIQUIDATION)),
						sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
					),
				)
			}
		}

		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}

// handleDefaultedLoans handles defaulted loans
func handleDefaultedLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all defaulted loans
	loans := k.GetLoans(ctx, types.LoanStatus_Defaulted)

	for _, loan := range loans {
		// check if the default liquidation cet has been signed
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.DefaultLiquidationCet.SignedTxHex) != 0 {
			continue
		}

		// check if the borrower adapted signatures already exist
		if len(dlcMeta.DefaultLiquidationCet.BorrowerAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.DefaultLiquidationEventId)
			if attestation == nil {
				continue
			}

			// decrypt the adaptor signatures
			for _, adaptorSignature := range dlcMeta.DefaultLiquidationCet.BorrowerAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptorSecret, _ := hex.DecodeString(attestation.Signature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.DefaultLiquidationCet.BorrowerAdaptedSignatures = append(
					dlcMeta.DefaultLiquidationCet.BorrowerAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed default liquidation cet if both borrower adapted signatures(obviously exist) and DCM signatures already exist
		if len(dlcMeta.DefaultLiquidationCet.DCMSignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedCet(dlcMeta.DefaultLiquidationCet.Tx, loan.BorrowerPubKey, dlcMeta.DefaultLiquidationCet.BorrowerAdaptedSignatures, loan.DCM, dlcMeta.DefaultLiquidationCet.DCMSignatures)
			if err != nil {
				k.Logger(ctx).Info("failed to build signed default liquidation cet", "loan id", loan.VaultAddress, "err", err)
			} else {
				dlcMeta.DefaultLiquidationCet.SignedTxHex = hex.EncodeToString(signedTx)

				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeGenerateSignedCet,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
						sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", types.CetType_DEFAULT_LIQUIDATION)),
						sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
					),
				)
			}
		}

		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}

// handleRepayments handles repayments
func handleRepayments(ctx sdk.Context, k keeper.Keeper) {
	// get all repaid loans
	loans := k.GetLoans(ctx, types.LoanStatus_Repaid)

	for _, loan := range loans {
		// // trigger dlc event if not triggered yet
		// if !k.DLCKeeper().GetEvent(ctx, triggeredEventId).HasTriggered {
		// 	k.DLCKeeper().TriggerDLCEvent(ctx, triggeredEventId, 0)
		// }
		// check if the repayment cet has been signed
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.RepaymentCet.SignedTxHex) != 0 {
			continue
		}

		// check if the DCM adaptor signatures have been submitted
		if len(dlcMeta.RepaymentCet.DCMAdaptorSignatures) == 0 {
			continue
		}

		if len(dlcMeta.RepaymentCet.DCMAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.RepaymentEventId)
			if attestation == nil {
				continue
			}

			// decrypt the DCM adaptor signatures
			for _, adaptorSignature := range dlcMeta.RepaymentCet.DCMAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptorSecret, _ := hex.DecodeString(attestation.Signature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.RepaymentCet.DCMAdaptedSignatures = append(
					dlcMeta.RepaymentCet.DCMAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed repayment cet
		signedTx, txHash, err := types.BuildSignedCet(dlcMeta.RepaymentCet.Tx, loan.BorrowerPubKey, dlcMeta.RepaymentCet.BorrowerSignatures, loan.DCM, dlcMeta.RepaymentCet.DCMAdaptedSignatures)
		if err != nil {
			k.Logger(ctx).Info("failed to build signed repayment cet", "loan id", loan.VaultAddress, "err", err)
		} else {
			dlcMeta.RepaymentCet.SignedTxHex = hex.EncodeToString(signedTx)

			if err := k.CompleteRepayment(ctx, loan); err != nil {
				k.Logger(ctx).Info("failed to complete repayment", "loan id", loan.VaultAddress, "err", err)
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
	}
}

// updatePools updates all active pools
func updatePools(ctx sdk.Context, k keeper.Keeper) {
	// get all active pools
	pools := k.GetPools(ctx, types.PoolStatus_ACTIVE)

	// get blocks per year
	blocksPerYear := k.GetBlocksPerYear(ctx)

	for _, pool := range pools {
		// update total borrowed amount every block
		//
		// Formula:
		//
		// borrowIndex_new = borrowIndex_old * (1 + borrowAPR/blocksPerYear)
		// totalBorrowed_new = totalBorrowed_old * borrowIndex_new/borrowIndex_old

		borrowIndexRatioNumerator := int64(1000*blocksPerYear) + int64(pool.Config.BorrowAPR)
		borrowIndexRatioDenominator := int64(1000 * blocksPerYear)

		pool.TotalBorrowed = pool.TotalBorrowed.Mul(sdkmath.NewInt(borrowIndexRatioNumerator).Quo(sdkmath.NewInt(borrowIndexRatioDenominator)))
	}
}
