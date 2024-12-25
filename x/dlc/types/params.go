package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PricePairSeparator defines the separator of the price pair
const PricePairSeparator = "-"

var (
	// default nonce queue size
	DefaultNonceQueueSize = uint32(50)

	// default price interval for btc-usd
	DefaultPriceIntervalForBTCUSD = int32(100)
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		NonceQueueSize: DefaultNonceQueueSize,
		PriceIntervals: []PriceInterval{
			{
				PricePair: "btc-usd",
				Interval:  int32(DefaultPriceIntervalForBTCUSD),
			},
		},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates params
func (p Params) Validate() error {
	if p.NonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "nonce queue size must be greater than 0")
	}

	for _, pi := range p.PriceIntervals {
		if err := validatePriceInterval(pi); err != nil {
			return err
		}
	}

	return nil
}

// validatePriceInterval validates the given price interval
func validatePriceInterval(priceInterval PriceInterval) error {
	if err := validatePricePair(priceInterval.PricePair); err != nil {
		return err
	}

	if priceInterval.Interval == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid price interval")
	}

	return nil
}

// validatePricePair validates the given price pair
func validatePricePair(pair string) error {
	denoms := strings.Split(pair, PricePairSeparator)
	if len(denoms) != 2 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid price pair")
	}

	for _, denom := range denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return err
		}
	}

	return nil
}
