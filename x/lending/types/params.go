package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	DefaultFinalTimeoutDuration = time.Duration(30 * 24 * 3600) // 30 days
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		OriginationFeeCollector: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		ProtocolFeeCollector:    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		FinalTimeoutDuration:    DefaultFinalTimeoutDuration,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if _, err := sdk.AccAddressFromBech32(p.OriginationFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid origination fee collector: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(p.ProtocolFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol fee collector: %v", err)
	}

	if p.FinalTimeoutDuration <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "final timeout duration must be greater than 0")
	}

	return nil
}
