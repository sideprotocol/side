package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	// default liquidation bonus
	DefaultLiquidationBonus = uint32(50) // 5%

	// default protocol liquidation fee
	DefaultProtocolLiquidationFee = uint32(100) // 10%
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		LiquidationBonus:                DefaultLiquidationBonus,
		ProtocolLiquidationFee:          DefaultProtocolLiquidationFee,
		ProtocolLiquidationFeeCollector: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.LiquidationBonus == 0 || p.LiquidationBonus >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid liquidation bonus")
	}

	if p.ProtocolLiquidationFee >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid protocol liquidation fee")
	}

	if _, err := sdk.AccAddressFromBech32(p.ProtocolLiquidationFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol liquidation fee collector: %v", err)
	}

	return nil
}
