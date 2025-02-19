package lending

import (
	"encoding/hex"
	"fmt"

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
	loans := k.GetLoans(ctx, types.LoanStatus_Disburse)

	for _, loan := range loans {
		// check if the loan has defaulted
		if ctx.BlockTime().Unix() >= loan.MaturityTime {
			loan.Status = types.LoanStatus_Default
			k.SetLoan(ctx, *loan)

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeDefault,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
				),
			)

			continue
		}

		liquidationPrice := types.GetLiquidationPrice(loan.CollateralAmount, loan.BorrowAmount.Amount, k.GetParams(ctx).LiquidationThresholdPercent)

		price, err := k.GetPrice(ctx, fmt.Sprintf("BTC-%s", loan.BorrowAmount.Denom))
		if err != nil {
			k.Logger(ctx).Info("failed to get oracle price", "err", err)
			continue
		}

		// check if the loan is to be liquidated
		if price.LTE(liquidationPrice) {
			loan.Status = types.LoanStatus_Liquidate
			k.SetLoan(ctx, *loan)

			// create auction
			auction := &auctiontypes.Auction{
				DepositedAsset:  sdk.NewCoin("sat", loan.CollateralAmount),
				Borrower:        loan.Borrower,
				LiquidatedPrice: liquidationPrice.Int64(),
				LiquidatedTime:  ctx.BlockTime(),
			}
			k.AuctionKeeper().CreateAuction(ctx, auction)

			// trigger price event
			k.DLCKeeper().TriggerEvent(ctx, loan.EventId)
		}
	}
}

// handleLiquidatedLoans handles liquidated loans
func handleLiquidatedLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all liquidated loans
	loans := k.GetLoans(ctx, types.LoanStatus_Liquidate)

	for _, loan := range loans {
		// check if the event attestation has been submitted
		attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.EventId)
		if attestation == nil {
			continue
		}

		// check if the adapted signature has been set in the dlc meta
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.LiquidationAdaptedSignature) != 0 {
			continue
		}

		// decrypt the liquidation adaptor signature
		adaptorSignature, _ := hex.DecodeString(dlcMeta.LiquidationAdaptorSignature)
		adaptorSecret, _ := hex.DecodeString(attestation.Signature)
		adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

		// set the adapted signature
		dlcMeta.LiquidationAdaptedSignature = hex.EncodeToString(adaptedSignature)
		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}
