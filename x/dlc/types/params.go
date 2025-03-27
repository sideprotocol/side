package types

import (
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PricePairSeparator defines the separator of the price pair
const PricePairSeparator = "-"

var (
	// default nonce queue size for price events
	DefaultPriceEventNonceQueueSize = uint32(50)

	// default price interval
	DefaultPriceInterval = int32(100)

	// default nonce queue size for date events
	DefaultDateEventNonceQueueSize = uint32(180)

	// default date interval
	DefaultDateInterval = 24 * time.Hour // 1 day

	// default nonce queue size for lending events
	DefaultLendingEventNonceQueueSize = uint32(1000)

	// default DKG timeout period
	DefaultDKGTimeoutPeriod = time.Duration(86400) * time.Second // 1 day
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		PriceEventNonceQueueSize: DefaultPriceEventNonceQueueSize,
		PriceIntervals: []PriceInterval{
			{
				PricePair: "BTC-USD",
				Interval:  int32(DefaultPriceInterval),
			},
		},
		DateEventNonceQueueSize:    DefaultDateEventNonceQueueSize,
		DateInterval:               DefaultDateInterval,
		LendingEventNonceQueueSize: DefaultLendingEventNonceQueueSize,
		DkgTimeoutPeriod:           DefaultDKGTimeoutPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates params
func (p Params) Validate() error {
	if p.PriceEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "price event nonce queue size must be greater than 0")
	}

	for _, pi := range p.PriceIntervals {
		if err := validatePriceInterval(pi); err != nil {
			return err
		}
	}

	if p.DateEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "date event nonce queue size must be greater than 0")
	}

	if p.DateInterval <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "date interval must be greater than 0")
	}

	if p.LendingEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "lending event nonce queue size must be greater than 0")
	}

	if err := validateDKGTimeoutPeriod(p.DkgTimeoutPeriod); err != nil {
		return err
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

// validateDKGTimeoutPeriod validates the given DKG timeout period
func validateDKGTimeoutPeriod(timeoutPeriod time.Duration) error {
	if timeoutPeriod == 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid dkg timeout period")
	}

	return nil
}
