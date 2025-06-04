package liquidation

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
	"github.com/sideprotocol/side/x/liquidation/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handlePendingLiquidations(ctx, k)
	handleCompletedLiquidations(ctx, k)
}

// handlePendingLiquidations handles the pending liquidations
func handlePendingLiquidations(ctx sdk.Context, k keeper.Keeper) {
	// get pending liquidations
	liquidations := k.GetLiquidations(ctx, types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING)
	if len(liquidations) == 0 {
		return
	}

	for _, liquidation := range liquidations {
		remainingCollateralAmount := liquidation.UnliquidatedCollateralAmount
		remainingDebtAmount := liquidation.DebtAmount.Sub(liquidation.LiquidatedDebtAmount)

		// check if there is no collateral or debt remaining
		if remainingCollateralAmount.Amount.IsZero() || remainingDebtAmount.Amount.IsZero() {
			k.Logger(ctx).Info("no collateral or debt remaining", "remaining collateral amount", remainingCollateralAmount, "remaining debt amount", remainingDebtAmount)

			liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED
			k.SetLiquidation(ctx, liquidation)

			continue
		}

		// check if the remaining collateral amount is dust
		if remainingCollateralAmount.Amount.Int64() < types.DefaultDustOutValue {
			k.Logger(ctx).Info("remaining collateral amount is dust value", "remaining collateral amount", remainingCollateralAmount)

			liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED
			k.SetLiquidation(ctx, liquidation)

			continue
		}

		minLiquidationDebtAmount := liquidation.DebtAmount.Amount.Mul(sdkmath.NewInt(int64(k.MinLiquidationFactor(ctx)))).Quo(sdkmath.NewInt(1000))
		if remainingDebtAmount.Amount.GTE(minLiquidationDebtAmount) {
			continue
		}

		currentPrice, err := k.GetPrice(ctx, types.GetPricePair(liquidation))
		if err != nil {
			continue
		}

		collateralDecimals := int(liquidation.CollateralAsset.Decimals)
		debtDecimals := int(liquidation.DebtAsset.Decimals)

		// check if the collateral amount corresponding to the remaining debt amount is dust
		collateralAmount := remainingDebtAmount.Amount.Mul(sdkmath.NewIntWithDecimal(1, collateralDecimals)).Quo(sdkmath.NewIntWithDecimal(1, debtDecimals)).ToLegacyDec().Quo(currentPrice).TruncateInt()
		if collateralAmount.Int64() < types.DefaultDustOutValue {
			k.Logger(ctx).Info("collateral amount corresponding to the remaining debt amount is dust value", "remaining debt amount", remainingDebtAmount, "collateral amount", collateralAmount, "price", currentPrice)

			liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED
			k.SetLiquidation(ctx, liquidation)
		}
	}
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
		if err := k.LiquidatedDebtHandler()(ctx, liquidation.Id, liquidation.LoanId, types.ModuleName, liquidation.LiquidatedDebtAmount); err != nil {
			k.Logger(ctx).Info("Failed to call LiquidatedDebtHandler", "liquidation id", liquidation.Id, "debt amount", liquidation.LiquidatedDebtAmount, "err", err)

			continue
		}

		// build settlement tx
		settlementTx, txHash, sigHashes, changeAmount, err := types.BuildSettlementTransaction(liquidation, k.GetLiquidationRecords(ctx, liquidation.Id), k.ProtocolLiquidationFeeCollector(ctx), feeRate.Value)
		if err != nil {
			k.Logger(ctx).Info("Failed to build settlement transaction", "liquidation id", liquidation.Id, "fee rate", feeRate.Value, "err", err)

			continue
		}

		liquidation.UnliquidatedCollateralAmount = sdk.NewInt64Coin(liquidation.CollateralAsset.Denom, changeAmount)
		liquidation.SettlementTx = settlementTx
		liquidation.SettlementTxId = txHash.String()
		liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_SETTLING

		// update liquidation
		k.SetLiquidation(ctx, liquidation)

		// initiate signing request via TSS
		k.TSSKeeper().InitiateSigningRequest(ctx, types.ModuleName, types.ToScopedId(liquidation.Id), tsstypes.SigningType_SIGNING_TYPE_SCHNORR_WITH_TWEAK, int32(types.SigningIntent_SIGNING_INTENT_DEFAULT), liquidation.DCM, sigHashes, &tsstypes.SigningOptions{Tweak: ""})
	}
}
