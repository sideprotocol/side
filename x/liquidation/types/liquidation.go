package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// network fee reserve for liquidation settlement
	LiquidationNetworkFeeReserve = int64(10000)
)

// LiquidatedDebtHandler defines the handler to perform liquidated debt handling
type LiquidatedDebtHandler func(ctx sdk.Context, liquidationId uint64, loanId string, moduleAccount string, debtAmount sdk.Coin) error

// GetPricePair gets the price pair of the given liquidation
func GetPricePair(liquidation *Liquidation) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(liquidation.CollateralAsset.PriceSymbol), strings.ToUpper(liquidation.DebtAsset.PriceSymbol))
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
