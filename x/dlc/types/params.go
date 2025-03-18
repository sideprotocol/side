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

	// default initial nonce count for lending events
	DefaultLendingEventInitialNonceCount = uint32(1000)

	// default nonce queue size for price events
	DefaultLendingEventNonceUsageThreshold = uint32(70) // 70%

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
		LendingEventInitialNonceCount:   DefaultLendingEventInitialNonceCount,
		LendingEventNonceUsageThreshold: DefaultLendingEventNonceUsageThreshold,
		DkgTimeoutPeriod:                DefaultDKGTimeoutPeriod,
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

	if p.LendingEventInitialNonceCount == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "lending event initial nonce count must be greater than 0")
	}

	if p.LendingEventNonceUsageThreshold == 0 || p.LendingEventNonceUsageThreshold >= 100 {
		return errorsmod.Wrap(ErrInvalidParams, "lending event nonce usage threshold must be between (0, 100)")
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
