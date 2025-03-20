package lending

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	auctiontypes "github.com/sideprotocol/side/x/auction/types"
	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleActiveLoans(ctx, k)
	handleLiquidatedLoans(ctx, k)
	handleRepayments(ctx, k)
}

// handleActiveLoans handles active loans
func handleActiveLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all active loans
	loans := k.GetLoans(ctx, types.LoanStatus_Open)

	for _, loan := range loans {
		// check if the loan has defaulted
		if ctx.BlockTime().Unix() >= loan.MaturityTime {
			loan.Status = types.LoanStatus_Defaulted
			k.SetLoan(ctx, loan)

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeDefault,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
				),
			)

			continue
		}

		liquidationPrice := types.GetLiquidationPrice(loan.CollateralAmount, loan.BorrowAmount.Amount, sdkmath.NewInt(int64(k.GetPool(ctx, loan.PoolId).Config.LiquidationThreshold)))

		price, err := k.GetPrice(ctx, fmt.Sprintf("BTC-%s", loan.BorrowAmount.Denom))
		if err != nil {
			k.Logger(ctx).Info("failed to get oracle price", "err", err)
			continue
		}

		// check if the loan is to be liquidated
		if price.LTE(liquidationPrice) {
			loan.Status = types.LoanStatus_Liquidated

			// get liquidation cet sig hashes; no error
			liquidationCetSigHashes, _ := types.GetLiquidationCetSigHashes(k.GetDLCMeta(ctx, loan.VaultAddress))

			// emit liquidation event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLiquidate,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
					sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(liquidationCetSigHashes, types.AttributeValueSeparator)),
				),
			)

			// create auction
			auction := k.AuctionKeeper().CreateAuction(ctx, &auctiontypes.Auction{
				LoanId:             loan.VaultAddress,
				Borrower:           loan.Borrower,
				Agency:             loan.Agency,
				DepositedAsset:     sdk.NewCoin("sat", loan.CollateralAmount),
				LiquidatedPrice:    liquidationPrice.Int64(),
				LiquidatedTime:     ctx.BlockTime(),
				ExpectedValue:      sdk.NewCoin(k.GetPool(ctx, loan.PoolId).Supply.Denom, loan.BorrowAmount.Amount.Add(loan.Interest)),
				LiquidationPenalty: k.GetPool(ctx, loan.PoolId).Config.LiquidationPenalty,
				LiquidationCet:     k.GetDLCMeta(ctx, loan.VaultAddress).LiquidationCet.Tx,
			})
			loan.AuctionId = auction.Id

			// trigger price event
			k.DLCKeeper().TriggerDLCEvent(ctx, loan.LiquidationEventId, 0)

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

		// build signed liquidation cet if both borrower adapted signatures(obviously exist) and agency signatures already exist
		if len(dlcMeta.LiquidationCet.AgencySignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedCet(dlcMeta.LiquidationCet.Tx, loan.BorrowerPubKey, dlcMeta.LiquidationCet.BorrowerAdaptedSignatures, loan.Agency, dlcMeta.LiquidationCet.AgencySignatures)
			if err != nil {
				k.Logger(ctx).Info("failed to build signed liquidation cet", "loan id", loan.VaultAddress, "err", err)
			} else {
				dlcMeta.LiquidationCet.SignedTxHex = hex.EncodeToString(signedTx)

				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeGenerateSignedLiquidationCet,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
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
		// check if the repayment cet has been signed
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.RepaymentCet.SignedTxHex) != 0 {
			continue
		}

		// check if the agency adaptor signatures have been submitted
		if len(dlcMeta.RepaymentCet.AgencyAdaptorSignatures) == 0 {
			continue
		}

		if len(dlcMeta.RepaymentCet.AgencyAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.RepaymentEventId)
			if attestation == nil {
				continue
			}

			// decrypt the agency adaptor signatures
			for _, adaptorSignature := range dlcMeta.RepaymentCet.AgencyAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptorSecret, _ := hex.DecodeString(attestation.Signature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.RepaymentCet.AgencyAdaptedSignatures = append(
					dlcMeta.RepaymentCet.AgencyAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed repayment cet
		signedTx, txHash, err := types.BuildSignedCet(dlcMeta.RepaymentCet.Tx, loan.BorrowerPubKey, dlcMeta.RepaymentCet.BorrowerSignatures, loan.Agency, dlcMeta.RepaymentCet.AgencyAdaptedSignatures)
		if err != nil {
			k.Logger(ctx).Info("failed to build signed repayment cet", "loan id", loan.VaultAddress, "err", err)
		} else {
			dlcMeta.RepaymentCet.SignedTxHex = hex.EncodeToString(signedTx)

			if err := k.CompleteRepayment(ctx, loan); err != nil {
				k.Logger(ctx).Info("failed to complete repayment", "loan id", loan.VaultAddress, "err", err)
			}

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeGenerateSignedRepaymentCet,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
				),
			)
		}

		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}
