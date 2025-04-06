package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
)

var (
	// default DKG timeout period
	DefaultDKGTimeoutPeriod = time.Duration(86400) * time.Second // 1 day
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		DkgTimeoutPeriod: DefaultDKGTimeoutPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates params
func (p Params) Validate() error {
	if err := validateDKGTimeoutPeriod(p.DkgTimeoutPeriod); err != nil {
		return err
	}

	return nil
}

// validateDKGTimeoutPeriod validates the given DKG timeout period
func validateDKGTimeoutPeriod(timeoutPeriod time.Duration) error {
	if timeoutPeriod <= 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid dkg timeout period")
	}

	return nil
}
