package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClaim{}

func NewMsgClaim(staker string, id uint64) *MsgClaim {
	return &MsgClaim{
		Staker: staker,
		Id:     id,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgClaim) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Staker); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	return nil
}
