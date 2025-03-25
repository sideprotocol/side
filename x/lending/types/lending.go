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
	OneYear = 365 * 24 * time.Hour
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

// GetCurrentInterest calculates the current interest based on the given params
func GetCurrentInterest(totalInterest sdkmath.Int, term time.Duration, startTime time.Time, currentTime time.Time) sdkmath.Int {
	elapsed := currentTime.Sub(startTime)

	return totalInterest.Mul(sdkmath.NewInt(int64(elapsed))).Quo(sdkmath.NewInt(int64(term)))
}

// GetLiquidationPrice calculates the liquidation price according to the liquidation LTV
// Formula:
// liquidation price = (borrow amount + interest) / lltv / collateral amount
func GetLiquidationPrice(collateralAmount sdkmath.Int, borrowAmount sdkmath.Int, borrowAPR uint32, lltv uint32) sdkmath.Int {
	interest := borrowAmount.Mul(sdkmath.NewInt(int64(borrowAPR))).Quo(Permille)
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

	if config.DebtCeiling.IsNil() || config.DebtCeiling.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "debt ceiling can not be nil or negative")
	}

	if config.OriginationFee.IsNil() || config.OriginationFee.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee can not be nil or negative")
	}

	if config.LiquidationThreshold == 0 || config.LiquidationThreshold >= 100 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid liquidation threshold")
	}

	if config.MaxLtv == 0 || config.MaxLtv >= 100 || config.MaxLtv > config.LiquidationThreshold {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid max ltv")
	}

	return nil
}
