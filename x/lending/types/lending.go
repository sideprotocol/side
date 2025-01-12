package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

// GetLiquidationPrice gets the liquidation price according to the liquidation LTV
func GetLiquidationPrice(collateralAmout sdkmath.Int, borrowedAmount sdkmath.Int, lltv sdkmath.Int) sdkmath.Int {
	liquidationValue := new(big.Int).Div(borrowedAmount.BigInt(), lltv.BigInt())
	liquidationPrice := new(big.Int).Div(liquidationValue, collateralAmout.BigInt())

	return sdkmath.NewIntFromBigInt(liquidationPrice)
}
