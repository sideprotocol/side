package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	// default minimum liquidation factor
	DefaultMinLiquidationFactor = uint32(20) // 2%

	// default liquidation bonus factor
	DefaultLiquidationBonusFactor = uint32(50) // 5%

	// default protocol liquidation fee factor
	DefaultProtocolLiquidationFeeFactor = uint32(100) // 10%
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		MinLiquidationFactor:            DefaultMinLiquidationFactor,
		LiquidationBonusFactor:          DefaultLiquidationBonusFactor,
		ProtocolLiquidationFeeFactor:    DefaultProtocolLiquidationFeeFactor,
		ProtocolLiquidationFeeCollector: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.MinLiquidationFactor == 0 || p.MinLiquidationFactor >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid minimum liquidation factor")
	}

	if p.LiquidationBonusFactor == 0 || p.LiquidationBonusFactor >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid liquidation bonus factor")
	}

	if p.ProtocolLiquidationFeeFactor >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid protocol liquidation fee factor")
	}

	if _, err := sdk.AccAddressFromBech32(p.ProtocolLiquidationFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol liquidation fee collector: %v", err)
	}

	return nil
}
