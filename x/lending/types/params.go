package types

import "cosmossdk.io/math"

var ()

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{}
}

// Validate validates the set of params
func (p Params) Validate() error {

	if p.BorrowRatePermille.LTE(p.SupplyRatePermille) {
		return ErrInvalidParams
	}
	if p.BorrowRatePermille.GT(Permille) {
		return ErrInvalidParams
	}
	if p.LiquidationThresholdPercent.LTE(math.NewInt(0)) {
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
