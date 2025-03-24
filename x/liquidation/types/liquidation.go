package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LiquidatedDebtHandler defines the handler to perform liquidated debt handling
type LiquidatedDebtHandler func(ctx sdk.Context, loanId string, moduleAccount string, debtAmount sdk.Coin) error
