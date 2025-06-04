package types

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// network fee reserve for liquidation settlement
	LiquidationNetworkFeeReserve = int64(10000)

	// default dust output value
	DefaultDustOutValue = int64(546)
)

// LiquidatedDebtHandler defines the handler to perform liquidated debt handling
type LiquidatedDebtHandler func(ctx sdk.Context, liquidationId uint64, loanId string, moduleAccount string, debtAmount sdk.Coin) error

// GetPricePair gets the price pair of the given liquidation
func GetPricePair(liquidation *Liquidation) string {
	if liquidation.CollateralAsset.IsBasePriceAsset {
		return fmt.Sprintf("%s%s", liquidation.CollateralAsset.PriceSymbol, liquidation.DebtAsset.PriceSymbol)
	}

	return fmt.Sprintf("%s%s", liquidation.DebtAsset.PriceSymbol, liquidation.CollateralAsset.PriceSymbol)
}

// GetCollateralAmount calculates the corresponding collateral amount according to the given debt amount and price
func GetCollateralAmount(debtAmount sdkmath.Int, debtAssetDecimals int, collateralAssetDecimals int, price sdkmath.LegacyDec, collateralIsBaseAsset bool) sdkmath.Int {
	if collateralIsBaseAsset {
		return debtAmount.Mul(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).Quo(sdkmath.NewIntWithDecimal(1, debtAssetDecimals)).ToLegacyDec().Quo(price).TruncateInt()
	}

	return debtAmount.Mul(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).ToLegacyDec().Mul(price).QuoInt(sdkmath.NewIntWithDecimal(1, debtAssetDecimals)).TruncateInt()
}

// GetDebtAmount calculates the corresponding debt amount according to the given collateral amount and price
func GetDebtAmount(collateralAmount sdkmath.Int, collateralAssetDecimals int, debtAssetDecimals int, price sdkmath.LegacyDec, collateralIsBaseAsset bool) sdkmath.Int {
	if collateralIsBaseAsset {
		return collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, debtAssetDecimals)).ToLegacyDec().Mul(price).QuoInt(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).TruncateInt()
	}

	return collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, debtAssetDecimals)).Quo(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).ToLegacyDec().Quo(price).TruncateInt()
}

// ToScopedId converts the given local id to the scoped id
func ToScopedId(id uint64) string {
	return fmt.Sprintf("%d", id)
}

// FromScopedId converts the scoped id to the local id
// Assume that the scoped id is valid
func FromScopedId(scopedId string) uint64 {
	id, _ := strconv.ParseUint(scopedId, 10, 64)
	return id
}
