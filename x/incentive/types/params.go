package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// default reward per deposit tx via btc bridge
	DefaultRewardPerDeposit = sdk.NewInt64Coin("uside", 100000000) // 100SIDE

	// default reward per withdrawal tx via btc bridge
	DefaultRewardPerWithdraw = sdk.NewInt64Coin("uside", 100000000) // 100SIDE
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		Enabled:           true,
		RewardPerDeposit:  DefaultRewardPerDeposit,
		RewardPerWithdraw: DefaultRewardPerWithdraw,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if !p.RewardPerDeposit.IsValid() || !p.RewardPerDeposit.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid deposit reward")
	}

	if !p.RewardPerDeposit.IsValid() || !p.RewardPerDeposit.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid withdrawal reward")
	}

	return nil
}
