package liquidation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
	"github.com/sideprotocol/side/x/liquidation/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleCompletedLiquidations(ctx, k)
}

// handleCompletedLiquidations handles the completed liquidations
func handleCompletedLiquidations(ctx sdk.Context, k keeper.Keeper) {
	// get completed liquidations
	liquidations := k.GetLiquidations(ctx, types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED)
	if len(liquidations) == 0 {
		return
	}

	// get fee rate
	feeRate := k.BtcBridgeKeeper().GetFeeRate(ctx)
	if err := k.BtcBridgeKeeper().CheckFeeRate(ctx, feeRate); err != nil {
		k.Logger(ctx).Info("Failed to get fee rate to handle liquidation", "err", err)

		return
	}

	for _, liquidation := range liquidations {
		// handle liquidated debt(repay the lending pool)
		liquidatedDebtAmount := liquidation.LiquidatedDebtAmount
		if err := k.LiquidatedDebtHandler()(ctx, liquidation.Id, liquidation.LoanId, types.ModuleName, liquidatedDebtAmount); err != nil {
			k.Logger(ctx).Info("Failed to call LiquidatedDebtHandler", "liquidation id", liquidation.Id, "debt amount", liquidatedDebtAmount, "err", err)

			continue
		}

		// build settlement tx
		settlementTx, txHash, sigHashes, changeAmount, err := types.BuildSettlementTransaction(liquidation, k.GetLiquidationRecords(ctx, liquidation.Id), k.ProtocolLiquidationFeeCollector(ctx), feeRate.Value)
		if err != nil {
			k.Logger(ctx).Info("Failed to build settlement transaction", "liquidation id", liquidation.Id, "fee rate", feeRate.Value, "err", err)

			continue
		}

		liquidation.UnliquidatedCollateralAmount = sdk.NewInt64Coin(liquidation.CollateralAmount.Denom, changeAmount)
		liquidation.SettlementTx = settlementTx
		liquidation.SettlementTxId = txHash.String()
		liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_SETTLING

		// update liquidation
		k.SetLiquidation(ctx, liquidation)

		// initiate signing request via TSS
		k.TSSKeeper().InitiateSigningRequest(ctx, types.ModuleName, types.ToScopedId(liquidation.Id), tsstypes.SigningType_SIGNING_TYPE_SCHNORR_WITH_TWEAK, int32(types.SigningIntent_SIGNING_INTENT_DEFAULT), liquidation.DCM, sigHashes, &tsstypes.SigningOptions{Tweak: ""})
	}
}
