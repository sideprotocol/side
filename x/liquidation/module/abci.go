package liquidation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/keeper"
	"github.com/sideprotocol/side/x/liquidation/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// handle pending liquidations
	handlePendingLiquidations(ctx, k)

	// handle completed liquidations
	return handleCompletedLiquidations(ctx, k)
}

// handlePendingLiquidations handles the pending liquidations
func handlePendingLiquidations(ctx sdk.Context, k keeper.Keeper) {
	// get pending liquidations
	liquidations := k.GetLiquidationsByStatus(ctx, types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING)
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

		minLiquidationDebtAmount := liquidation.DebtAmount.Amount.ToLegacyDec().Mul(k.MinLiquidationFactor(ctx)).TruncateInt()
		if remainingDebtAmount.Amount.GTE(minLiquidationDebtAmount) {
			continue
		}

		currentPrice, err := k.GetPrice(ctx, types.GetPricePair(liquidation))
		if err != nil {
			continue
		}

		collateralDecimals := int(liquidation.CollateralAsset.Decimals)
		debtDecimals := int(liquidation.DebtAsset.Decimals)
		collateralIsBaseAsset := liquidation.CollateralAsset.IsBasePriceAsset

		// check if the collateral amount corresponding to the remaining debt amount is dust
		collateralAmount := types.GetCollateralAmount(remainingDebtAmount.Amount, debtDecimals, collateralDecimals, currentPrice, collateralIsBaseAsset)
		if collateralAmount.Int64() < types.DefaultDustOutValue {
			k.Logger(ctx).Info("collateral amount corresponding to the remaining debt amount is dust value", "remaining debt amount", remainingDebtAmount, "collateral amount", collateralAmount, "price", currentPrice)

			liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED
			k.SetLiquidation(ctx, liquidation)
		}
	}
}

// handleCompletedLiquidations handles the completed liquidations
func handleCompletedLiquidations(ctx sdk.Context, k keeper.Keeper) error {
	// get completed liquidations
	liquidations := k.GetLiquidationsByStatus(ctx, types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED)
	if len(liquidations) == 0 {
		return nil
	}

	// get fee rate
	// NOTE: the fee rate validity is not necessary here
	// if it is 0 or too high, we use the reserved network fee duration liquidation, which is sufficient to relay the tx by design
	feeRate := k.BtcBridgeKeeper().GetFeeRate(ctx)
	if err := k.BtcBridgeKeeper().CheckFeeRate(ctx, feeRate); err != nil {
		k.Logger(ctx).Warn("Failed to get valid fee rate to handle liquidation", "err", err)
	}

	for _, liquidation := range liquidations {
		// build settlement tx
		settlementTx, txHash, sigHashes, changeAmount, err := types.BuildSettlementTransaction(liquidation, k.GetLiquidationRecords(ctx, liquidation.Id), k.ProtocolLiquidationFeeCollector(ctx), feeRate.Value, types.LiquidationNetworkFeeReserve)
		if err != nil {
			k.Logger(ctx).Error("Failed to build settlement transaction", "liquidation id", liquidation.Id, "fee rate", feeRate.Value, "err", err)
			continue
		}

		// handle liquidated debt (repay the lending pool)
		if err := k.LiquidatedDebtHandler()(ctx, liquidation.Id, liquidation.LoanId, types.ModuleName, liquidation.LiquidatedDebtAmount); err != nil {
			// unexpected error
			k.Logger(ctx).Error("Failed to call LiquidatedDebtHandler", "liquidation id", liquidation.Id, "debt amount", liquidation.LiquidatedDebtAmount, "err", err)
			return err
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

	return nil
}
