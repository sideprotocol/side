package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/liquidation/types"
)

// HandleLiquidation performs the liquidation handling
func (k Keeper) HandleLiquidation(ctx sdk.Context, liquidator string, liquidationId uint64, debtAmount sdk.Coin) (*types.LiquidationRecord, error) {
	if !k.HasLiquidation(ctx, liquidationId) {
		return nil, types.ErrLiquidationDoesNotExist
	}

	liquidation := k.GetLiquidation(ctx, liquidationId)
	if liquidation.Status != types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING {
		return nil, errorsmod.Wrap(types.ErrInvalidLiquidationStatus, "non liquidating status")
	}

	if debtAmount.Denom != liquidation.DebtAmount.Denom {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "mismatched debt amount denom")
	}

	if debtAmount.Amount.LT(liquidation.DebtAmount.Amount.Mul(sdkmath.NewInt(int64(k.MinLiquidationFactor(ctx)))).Quo(sdkmath.NewInt(1000))) {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "liquidation debt amount must be greater or equal than minimum liquidation factor")
	}

	currentPrice, err := k.GetPrice(ctx, "BTCUSD")
	if err != nil {
		return nil, types.ErrInvalidPrice
	}

	// check remaining debt amount
	remainingDebtAmount := liquidation.DebtAmount.Sub(liquidation.LiquidatedDebtAmount)
	if remainingDebtAmount.IsLT(debtAmount) {
		debtAmount = remainingDebtAmount
	}

	// calculate collateral amount
	collateralAmount := debtAmount.Amount.Mul(sdkmath.NewIntWithDecimal(1, 8)).Quo(sdkmath.NewIntWithDecimal(1, 6)).ToLegacyDec().Quo(currentPrice).TruncateInt()

	// check remaining collateral amount
	remainingCollateralAmount := liquidation.CollateralAmount.Sub(liquidation.LiquidatedCollateralAmount).SubAmount(sdkmath.NewInt(10000))
	if remainingCollateralAmount.Amount.LT(collateralAmount) {
		collateralAmount = remainingCollateralAmount.Amount
		debtAmount.Amount = collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, 6)).ToLegacyDec().Mul(currentPrice).QuoInt(sdkmath.NewIntWithDecimal(1, 8)).TruncateInt()
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(liquidator), types.ModuleName, sdk.NewCoins(debtAmount)); err != nil {
		return nil, err
	}

	remainingCollateralAmount = remainingCollateralAmount.SubAmount(collateralAmount)

	// calculate bonus
	bonusAmountInDebt := debtAmount.Amount.Mul(sdkmath.NewInt(int64(k.LiquidationBonusFactor(ctx)))).Quo(sdkmath.NewInt(1000))
	bonusAmount := bonusAmountInDebt.Mul(sdkmath.NewIntWithDecimal(1, 8)).Quo(sdkmath.NewIntWithDecimal(1, 6)).ToLegacyDec().Quo(currentPrice).TruncateInt()

	// check if there is left collateral for bonus
	if bonusAmount.GT(remainingCollateralAmount.Amount) {
		bonusAmount = remainingCollateralAmount.Amount
	}

	protocolLiquidationFee := bonusAmount.Mul(sdkmath.NewInt(int64(k.ProtocolLiquidationFeeFactor(ctx)))).Quo(sdkmath.NewInt(1000))

	record := &types.LiquidationRecord{
		Id:               k.IncrementLiquidationRecordId(ctx),
		LiquidationId:    liquidationId,
		Liquidator:       liquidator,
		DebtAmount:       debtAmount,
		CollateralAmount: sdk.NewCoin(liquidation.CollateralAmount.Denom, collateralAmount.Add(bonusAmount).Sub(protocolLiquidationFee)),
		Time:             ctx.BlockTime(),
	}

	liquidation.LiquidatedCollateralAmount = liquidation.LiquidatedCollateralAmount.AddAmount(collateralAmount).AddAmount(bonusAmount)
	liquidation.LiquidatedDebtAmount = liquidation.LiquidatedDebtAmount.Add(debtAmount)
	liquidation.LiquidationBonusAmount = liquidation.LiquidationBonusAmount.AddAmount(bonusAmount)
	liquidation.ProtocolLiquidationFee = liquidation.ProtocolLiquidationFee.AddAmount(protocolLiquidationFee)
	liquidation.UnliquidatedCollateralAmount = liquidation.CollateralAmount.Sub(liquidation.LiquidatedCollateralAmount)

	remainingCollateralAmount = liquidation.CollateralAmount.Sub(liquidation.LiquidatedCollateralAmount).SubAmount(sdkmath.NewInt(10000))
	if remainingCollateralAmount.Amount.IsZero() || liquidation.LiquidatedDebtAmount.Amount.Equal(liquidation.DebtAmount.Amount) {
		liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATED
	}

	k.SetLiquidationRecord(ctx, record)
	k.SetLiquidation(ctx, liquidation)

	return record, nil
}

// GetLiquidationId gets the current liquidation id
func (k Keeper) GetLiquidationId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.LiquidationIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementLiquidationId increments the liquidation id and returns the new id
func (k Keeper) IncrementLiquidationId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetLiquidationId(ctx) + 1
	store.Set(types.LiquidationIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasLiquidation returns true if the given liquidation exists, false otherwise
func (k Keeper) HasLiquidation(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.LiquidationKey(id))
}

// GetLiquidation gets the liquidation by the given id
func (k Keeper) GetLiquidation(ctx sdk.Context, id uint64) *types.Liquidation {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.LiquidationKey(id))
	var liquidation types.Liquidation
	k.cdc.MustUnmarshal(bz, &liquidation)

	return &liquidation
}

// SetLiquidation sets the given liquidation
func (k Keeper) SetLiquidation(ctx sdk.Context, liquidation *types.Liquidation) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(liquidation)
	store.Set(types.LiquidationKey(liquidation.Id), bz)
}

// CreateLiquidation creates and returns the newly created liquidation
func (k Keeper) CreateLiquidation(ctx sdk.Context, liquidation *types.Liquidation) *types.Liquidation {
	// set the id
	liquidation.Id = k.IncrementLiquidationId(ctx)

	// set the status to liquidating
	liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING

	k.SetLiquidation(ctx, liquidation)

	return liquidation
}

// GetAllLiquidations gets all liquidations
func (k Keeper) GetAllLiquidations(ctx sdk.Context) []*types.Liquidation {
	liquidations := make([]*types.Liquidation, 0)

	k.IterateLiquidations(ctx, func(liquidation *types.Liquidation) (stop bool) {
		liquidations = append(liquidations, liquidation)
		return false
	})

	return liquidations
}

// GetLiquidations gets liquidations by the given status
func (k Keeper) GetLiquidations(ctx sdk.Context, status types.LiquidationStatus) []*types.Liquidation {
	liquidations := make([]*types.Liquidation, 0)

	k.IterateLiquidations(ctx, func(liquidation *types.Liquidation) (stop bool) {
		if liquidation.Status == status {
			liquidations = append(liquidations, liquidation)
		}

		return false
	})

	return liquidations
}

// IterateLiquidations iterates through all liquidations
func (k Keeper) IterateLiquidations(ctx sdk.Context, cb func(liquidation *types.Liquidation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidationKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var liquidation types.Liquidation
		k.cdc.MustUnmarshal(iterator.Value(), &liquidation)

		if cb(&liquidation) {
			break
		}
	}
}
