package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LiquidationBonus returns the liquidation bonus
func (k Keeper) LiquidationBonus(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).LiquidationBonus
}

// ProtocolLiquidationFee returns the protocol liquidation fee
func (k Keeper) ProtocolLiquidationFee(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).ProtocolLiquidationFee
}

// ProtocolLiquidationFeeCollector returns the protocol liquidation fee collector
func (k Keeper) ProtocolLiquidationFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).ProtocolLiquidationFeeCollector
}
