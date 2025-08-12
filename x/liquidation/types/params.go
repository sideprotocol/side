package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	// default minimum liquidation factor
	DefaultMinLiquidationFactor = sdkmath.LegacyMustNewDecFromStr("0.02") // 2%

	// maximum liquidation bonus factor
	MaxLiquidationBonusFactor = sdkmath.LegacyMustNewDecFromStr("0.1") // 10%

	// default liquidation bonus factor
	DefaultLiquidationBonusFactor = sdkmath.LegacyMustNewDecFromStr("0.05") // 5%

	// default protocol liquidation fee factor
	DefaultProtocolLiquidationFeeFactor = sdkmath.LegacyMustNewDecFromStr("0.1") // 10%
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
	if !p.MinLiquidationFactor.IsPositive() || p.MinLiquidationFactor.GT(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "invalid minimum liquidation factor")
	}

	if !p.LiquidationBonusFactor.IsPositive() || p.LiquidationBonusFactor.GT(MaxLiquidationBonusFactor) {
		return errorsmod.Wrap(ErrInvalidParams, "invalid liquidation bonus factor")
	}

	if p.ProtocolLiquidationFeeFactor.IsNegative() || p.ProtocolLiquidationFeeFactor.GTE(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "invalid protocol liquidation fee factor")
	}

	if _, err := sdk.AccAddressFromBech32(p.ProtocolLiquidationFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol liquidation fee collector: %v", err)
	}

	return nil
}
