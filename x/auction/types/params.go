package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
)

var (
	// default price drop period
	DefaultPriceDropPeriod = time.Duration(10) * time.Minute // 10min

	// default initial discount
	DefaultInitialDiscount = uint32(3) // 3%

	// default minimum amount for bid
	DefaultMinBidAmount = uint64(100000) // 100000sat
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		PriceDropPeriod: DefaultPriceDropPeriod,
		InitialDiscount: DefaultInitialDiscount,
		MinBidAmount:    DefaultMinBidAmount,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.PriceDropPeriod <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "price drop period must be greater than 0")
	}

	if p.InitialDiscount == 0 || p.InitialDiscount >= 100 {
		return errorsmod.Wrap(ErrInvalidParams, "initial discount must be between (0, 100)")
	}

	if p.MinBidAmount == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "minimum bid amount must be greater than 0")
	}

	return nil
}
