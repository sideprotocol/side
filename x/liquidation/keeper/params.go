package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MinLiquidationFactor returns the minimum liquidation factor
func (k Keeper) MinLiquidationFactor(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).MinLiquidationFactor
}

// LiquidationBonusFactor returns the liquidation bonus factor
func (k Keeper) LiquidationBonusFactor(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LiquidationBonusFactor
}

// ProtocolLiquidationFeeFactor returns the protocol liquidation fee factor
func (k Keeper) ProtocolLiquidationFeeFactor(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).ProtocolLiquidationFeeFactor
}

// ProtocolLiquidationFeeCollector returns the protocol liquidation fee collector
func (k Keeper) ProtocolLiquidationFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).ProtocolLiquidationFeeCollector
}
