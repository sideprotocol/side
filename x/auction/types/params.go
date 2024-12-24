package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
)

var (
	// default price drop period
	DefaultPriceDropPeriod = time.Duration(10) * time.Minute

	// default initial discount
	DefaultInitialDiscount = uint32(90)

	// default fee rate base point
	DefaultFeeRate = uint32(30) // fee rate base point; 3/1000

	// default minimum amount for bid
	DefaultMinBidAmount = uint64(100000) // 100000sat
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		PriceDropPeriod: DefaultPriceDropPeriod,
		InitialDiscount: DefaultInitialDiscount,
		FeeRate:         DefaultFeeRate,
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

	if p.InitialDiscount == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "initial discount must be greater than 0")
	}

	if p.FeeRate == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "fee rate must be greater than 0")
	}

	if p.MinBidAmount == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "minimum bid amount must be greater than 0")
	}

	return nil
}
