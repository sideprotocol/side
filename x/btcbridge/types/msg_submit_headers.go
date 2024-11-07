package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitBlockHeaders{}

func NewMsgSubmitBlockHeaders(
	sender string,
	headers []*BlockHeader,
) *MsgSubmitBlockHeaders {
	return &MsgSubmitBlockHeaders{
		Sender:       sender,
		BlockHeaders: headers,
	}
}

func (msg *MsgSubmitBlockHeaders) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.BlockHeaders) == 0 {
		return errorsmod.Wrap(ErrInvalidBlockHeader, "block headers cannot be empty")
	}

	return nil
}
