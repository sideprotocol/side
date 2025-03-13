package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdatePoolConfig{}

// ValidateBasic performs basic message validation.
func (m *MsgUpdatePoolConfig) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if len(m.PoolId) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "empty pool id")
	}

	return ValidatePoolConfig(m.Config)
}
