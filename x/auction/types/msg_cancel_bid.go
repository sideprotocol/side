package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCancelBid{}

func NewMsgCancelBid(
	sender string,
	id uint64,
) *MsgCancelBid {
	return &MsgCancelBid{
		Sender: sender,
		Id:     id,
	}
}

// ValidateBasic performs basic MsgCancelBid message validation.
func (m *MsgCancelBid) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	return nil
}
