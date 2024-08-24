package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitBlockHeader = "submit_block_header"

func NewMsgSubmitBlockHeaders(
	sender string,
	headers []*BlockHeader,
) *MsgSubmitBlockHeaders {
	return &MsgSubmitBlockHeaders{
		Sender:       sender,
		BlockHeaders: headers,
	}
}

func (msg *MsgSubmitBlockHeaders) Route() string {
	return RouterKey
}

func (msg *MsgSubmitBlockHeaders) Type() string {
	return TypeMsgSubmitBlockHeader
}

func (msg *MsgSubmitBlockHeaders) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitBlockHeaders) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitBlockHeaders) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.BlockHeaders) == 0 {
		return sdkerrors.Wrap(ErrInvalidHeader, "block headers cannot be empty")
	}

	return nil
}
