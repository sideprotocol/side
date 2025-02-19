package types

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	DefaultSupplyRatePermille = sdkmath.NewInt(5)

	DefaultBorrowRatePermille = sdkmath.NewInt(7)

	DefaultLiquidationThresholdPercent = sdkmath.NewInt(70)

	DefaultMinInitialLtvPercent = sdkmath.NewInt(80)
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		SupplyRatePermille:          DefaultSupplyRatePermille,
		BorrowRatePermille:          DefaultBorrowRatePermille,
		LiquidationThresholdPercent: DefaultLiquidationThresholdPercent,
		MinInitialLtvPercent:        DefaultMinInitialLtvPercent,
		FeeRecipient:                authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.BorrowRatePermille.LTE(p.SupplyRatePermille) {
		return ErrInvalidParams
	}

	if p.BorrowRatePermille.GT(Permille) {
		return ErrInvalidParams
	}

	if p.LiquidationThresholdPercent.LTE(sdkmath.NewInt(0)) {
		return ErrInvalidParams
	}

	if p.MinInitialLtvPercent.LTE(p.LiquidationThresholdPercent) {
		return ErrInvalidParams
	}

	if p.MinInitialLtvPercent.GT(Percent) {
		return ErrInvalidParams
	}

	return nil
}
