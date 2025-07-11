package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUnstake{}

func NewMsgUnstake(staker string, id uint64) *MsgUnstake {
	return &MsgUnstake{
		Staker: staker,
		Id:     id,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgUnstake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Staker); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	return nil
}
