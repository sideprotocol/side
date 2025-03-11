package types

import (
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
func GetLiquidationPrice(collateralAmount sdkmath.Int, borrowedAmount sdkmath.Int, lltv sdkmath.Int) sdkmath.Int {
	// liquidation price = borrowed amount / (lltv/100) / collateral amount
	liquidationPrice := borrowedAmount.Mul(sdkmath.NewInt(100000000)).Mul(Percent).Quo(lltv).Quo(collateralAmount).Quo(sdkmath.NewInt(1000000))

	// price precision
	precision := sdkmath.NewInt(100)

	return liquidationPrice.Quo(precision).Mul(precision)
}
