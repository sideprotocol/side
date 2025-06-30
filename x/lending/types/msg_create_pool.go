package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreatePool{}

// ValidateBasic performs basic message validation.
func (m *MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if err := sdk.ValidateDenom(YTokenDenom(m.Id)); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolId, "%v", err)
	}

	return ValidatePoolConfig(m.Config)
}
