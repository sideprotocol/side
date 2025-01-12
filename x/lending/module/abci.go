package lending

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	auctiontypes "github.com/sideprotocol/side/x/auction/types"
	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleLoans(ctx, k)
}

// handleLoans handles loans
func handleLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all valid loans
	loans := k.GetLoans(ctx, types.LoanStatus_Disburse)

	for _, loan := range loans {
		liquidationPrice := types.GetLiquidationPrice(loan.CollateralAmount, loan.BorrowAmount.Amount, k.GetParams(ctx).LiquidationThresholdPercent)

		price, err := k.OracleKeeper().GetPrice(ctx, fmt.Sprintf("BTC-%s", loan.BorrowAmount.Denom))
		if err != nil {
			k.Logger(ctx).Info("failed to get oracle price", "err", err)
			continue
		}

		// liquidated
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

			dlcEvent := k.DLCKeeper().GetEvent(ctx, loan.EventId)

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLiquidate,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyEventPubKey, dlcEvent.Pubkey),
					sdk.NewAttribute(types.AttributeKeyEventNonce, dlcEvent.Nonce),
					sdk.NewAttribute(types.AttributeKeyEventPrice, liquidationPrice.String()),
				),
			)
		}
	}
}
