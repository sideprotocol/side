package types

import (
	"time"
)

var (
	DefaultFinalTimeoutDuration = 30 * 24 * time.Hour // 30 days
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}
