package liquidation

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
	"github.com/sideprotocol/side/x/liquidation/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleCompletedLiquidations(ctx, k)
}

// handleCompletedLiquidations handles the completed liquidations
func handleCompletedLiquidations(ctx sdk.Context, k keeper.Keeper) {
	// get completed liquidations
	liquidations := k.GetLiquidations(ctx, types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED)

	for _, liquidation := range liquidations {
		// handle liquidated debt(repay the lending pool)
		liquidatedDebtAmount := liquidation.LiquidatedDebtAmount
		if err := k.LiquidatedDebtHandler()(ctx, liquidation.Id, liquidation.LoanId, types.ModuleName, liquidatedDebtAmount); err != nil {
			k.Logger(ctx).Info("Failed to call LiquidatedDebtHandler", "liquidation id", liquidation.Id, "debt amount", liquidatedDebtAmount, "err", err)

			continue
		}

		// build settlement tx
		settlementTx, txHash, sigHashes, changeAmount, err := types.BuildSettlementTransaction(liquidation, k.GetLiquidationRecords(ctx, liquidation.Id), k.ProtocolLiquidationFeeCollector(ctx), 5)
		if err != nil {
			k.Logger(ctx).Info("Failed to build settlement transaction", "liquidation id", liquidation.Id, "err", err)

			continue
		}

		liquidation.UnliquidatedCollateralAmount = sdk.NewInt64Coin(liquidation.CollateralAmount.Denom, changeAmount)
		liquidation.SettlementTx = settlementTx
		liquidation.SettlementTxId = txHash.String()
		liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_SETTLING

		// update liquidation
		k.SetLiquidation(ctx, liquidation)

		// emit event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSignSettlementTransaction,
				sdk.NewAttribute(types.AttributeKeyLiquidationId, fmt.Sprintf("%d", liquidation.Id)),
				sdk.NewAttribute(types.AttributeKeyDCMPubKey, liquidation.DCM),
				sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(sigHashes, types.AttributeValueSeparator)),
			),
		)
	}
}
