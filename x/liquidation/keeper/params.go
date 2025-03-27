package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
