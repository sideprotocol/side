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
				LiquidationCet:     k.GetDLCMeta(ctx, loan.VaultAddress).LiquidationCet,
			})
			loan.AuctionId = auction.Id

			// trigger price event
			k.DLCKeeper().TriggerEvent(ctx, loan.EventId)

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
		// check if the signed liquidation cet has been generated in the dlc meta
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.SignedLiquidationCetHex) != 0 {
			continue
		}

		// check if the adapted signature has been set in the dlc meta
		if len(dlcMeta.LiquidationAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.EventId)
			if attestation == nil {
				continue
			}

			for _, adaptorSignature := range dlcMeta.LiquidationAdaptorSignatures {
				// decrypt the adaptor signature
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptorSecret, _ := hex.DecodeString(attestation.Signature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.LiquidationAdaptedSignatures = append(
					dlcMeta.LiquidationAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed liquidation cet if both adapted signatures(obviously exist) and agency signatures already exist
		if len(dlcMeta.LiquidationAgencySignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedLiquidationCet(dlcMeta.LiquidationCet, loan.BorrowerPubKey, dlcMeta.LiquidationAdaptedSignatures, loan.Agency, dlcMeta.LiquidationAgencySignatures)
			if err != nil {
				k.Logger(ctx).Info("failed to build signed liquidation cet", "err", err)
			} else {
				dlcMeta.SignedLiquidationCetHex = hex.EncodeToString(signedTx)

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
