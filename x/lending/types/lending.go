package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

const (
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

// ValidatePoolConfig validates the given pool config
func ValidatePoolConfig(config PoolConfig) error {
	if config.BorrowRate <= config.SupplyRate {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow rate must be greater than supply rate")
	}

	if config.SupplyCap.IsNil() || config.SupplyCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "supply cap can not be nil or negative")
	}

	if config.BorrowCap.IsNil() || config.BorrowCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow cap can not be nil or negative")
	}

	if config.DebtCeiling.IsNil() || config.DebtCeiling.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "debt ceiling can not be nil or negative")
	}

	if config.OriginationFee.IsNil() || config.OriginationFee.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee can not be nil or negative")
	}

	if config.LiquidationThreshold == 0 || config.LiquidationThreshold >= 100 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid liquidation threshold")
	}

	if config.Ltv == 0 || config.Ltv >= 100 || config.Ltv > config.LiquidationThreshold {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid ltv")
	}

	return nil
}
