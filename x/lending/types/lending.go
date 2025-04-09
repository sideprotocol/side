package types

import (
	"encoding/hex"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	"github.com/sideprotocol/side/crypto/adaptor"
)

const (
	// OneYear represents the seconds in one year
	OneYear = 365 * 24 * 3600
)

// GetExchangeRate calculates the sToken exchange rate according to the given params
// Formula:
// exchange rate = (totalAvailable + total borrowed) / totalSTokens
func GetExchangeRate(totalAvailable sdkmath.Int, totalBorrowed sdkmath.Int, totalSTokens sdkmath.Int) sdkmath.LegacyDec {
	if totalSTokens.IsZero() {
		return sdkmath.LegacyOneDec()
	}

	return sdkmath.LegacyNewDecFromInt(totalAvailable.Add(totalBorrowed)).Quo(totalSTokens.ToLegacyDec())
}

// GetInterest calculates the loan interest based on the given params
func GetInterest(totalInterest sdkmath.Int, term time.Duration, startTime int64, destTime int64) sdkmath.Int {
	elapsed := destTime - startTime

	return totalInterest.Mul(sdkmath.NewInt(elapsed)).Quo(sdkmath.NewInt(int64(term)))
}

// GetLiquidationPrice calculates the liquidation price according to the liquidation LTV
// Formula:
// liquidation price = (borrow amount + interest) / lltv / collateral amount
func GetLiquidationPrice(collateralAmount sdkmath.Int, borrowAmount sdkmath.Int, term int64, borrowAPR uint32, lltv uint32) sdkmath.Int {
	interest := borrowAmount.Mul(sdkmath.NewInt(int64(borrowAPR))).Mul(sdkmath.NewInt(term)).Quo(sdkmath.NewInt(OneYear)).Quo(Permille)
	liquidationPrice := borrowAmount.Add(interest).Mul(sdkmath.NewInt(100000000)).Mul(Percent).Quo(sdkmath.NewInt(int64(lltv))).Quo(collateralAmount).Quo(sdkmath.NewInt(1000000))

	// price precision
	precision := sdkmath.NewInt(100)

	return liquidationPrice.Quo(precision).Mul(precision)
}

// GetDefaultLiquidationDate gets the date at which the loan will be liquidated due to default
func GetDefaultLiquidationDate(maturityTime int64) int64 {
	if maturityTime%(24*int64(time.Hour)) == 0 {
		return maturityTime
	}

	return time.Unix(maturityTime, 0).Truncate(24 * time.Hour).Add(24 * time.Hour).Unix()
}

// AdaptorPointFromSecret gets the corresponding adaptor point from the given secret
func AdaptorPointFromSecret(secret []byte) string {
	return hex.EncodeToString(adaptor.SecretToPubKey(secret))
}

// HasSupplyCap returns true if the supply cap set in the given pool, false otherwise
func HasSupplyCap(pool *LendingPool) bool {
	return pool.Config.SupplyCap.IsPositive()
}

// HasBorrowCap returns true if the borrow cap set in the given pool, false otherwise
func HasBorrowCap(pool *LendingPool) bool {
	return pool.Config.BorrowCap.IsPositive()
}

// HasMinBorrowAmountLimit returns true if the min borrow amount set in the given pool, false otherwise
func HasMinBorrowAmountLimit(pool *LendingPool) bool {
	return pool.Config.MinBorrowAmount.IsPositive()
}

// HasMaxBorrowAmountLimit returns true if the max borrow amount set in the given pool, false otherwise
func HasMaxBorrowAmountLimit(pool *LendingPool) bool {
	return pool.Config.MaxBorrowAmount.IsPositive()
}

// CheckSupplyCap checks if the supply cap will be exceeded for the given deposit amount
func CheckSupplyCap(pool *LendingPool, depositAmount sdkmath.Int) error {
	if HasSupplyCap(pool) && pool.Supply.Amount.Add(depositAmount).GT(pool.Config.SupplyCap) {
		return ErrSupplyCapExceeded
	}

	return nil
}

// CheckBorrowCap checks if the borrow cap will be exceeded for the given borrow amount
func CheckBorrowCap(pool *LendingPool, borrowAmount sdkmath.Int) error {
	if HasBorrowCap(pool) && pool.TotalBorrowed.Add(borrowAmount).GT(pool.Config.BorrowCap) {
		return ErrBorrowCapExceeded
	}

	return nil
}

// CheckBorrowAmountLimit checks if the borrow amount satisfies limits for the given pool
func CheckBorrowAmountLimit(pool *LendingPool, borrowAmount sdkmath.Int) error {
	if HasMinBorrowAmountLimit(pool) && borrowAmount.LT(pool.Config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount can not be less than min borrow amount")
	}

	if HasMaxBorrowAmountLimit(pool) && borrowAmount.GT(pool.Config.MaxBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount can not be greater than max borrow amount")
	}

	return nil
}

// ValidatePoolConfig validates the given pool config
func ValidatePoolConfig(config PoolConfig) error {
	if config.BorrowAPR == 0 || config.BorrowAPR >= 1000 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow apr must be between (0, 1000)")
	}

	if config.SupplyCap.IsNil() || config.SupplyCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "supply cap can not be nil or negative")
	}

	if config.BorrowCap.IsNil() || config.BorrowCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow cap can not be nil or negative")
	}

	if config.MinBorrowAmount.IsNil() || config.MinBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "min borrow amount can not be nil or negative")
	}

	if config.MaxBorrowAmount.IsNil() || config.MaxBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount can not be nil or negative")
	}

	if config.MinBorrowAmount.IsPositive() && config.MaxBorrowAmount.IsPositive() && config.MaxBorrowAmount.LT(config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount can not be less than min borrow amount")
	}

	if config.OriginationFee.IsNil() || config.OriginationFee.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee can not be nil or negative")
	}

	if config.OriginationFee.GTE(config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee must be less than min borrow amount")
	}

	if config.LiquidationThreshold == 0 || config.LiquidationThreshold >= 100 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid liquidation threshold")
	}

	if config.MaxLtv == 0 || config.MaxLtv >= 100 || config.MaxLtv >= config.LiquidationThreshold {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid max ltv")
	}

	return nil
}
