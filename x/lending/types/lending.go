package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
)

const (
	// minimum pool id length
	MinPoolIdLength = 2

	// loan secret length
	LoanSecretLength = 32

	// loan secret hash length
	LoanSecretHashLength = 32
)

// GetLiquidationPrice gets the liquidation price according to the liquidation LTV
func GetLiquidationPrice(collateralAmout sdkmath.Int, borrowedAmount sdkmath.Int, lltv sdkmath.Int) sdkmath.Int {
	liquidationValue := new(big.Int).Div(borrowedAmount.BigInt(), lltv.BigInt())
	liquidationPrice := new(big.Int).Div(liquidationValue, collateralAmout.BigInt())

	return sdkmath.NewIntFromBigInt(liquidationPrice)
}
