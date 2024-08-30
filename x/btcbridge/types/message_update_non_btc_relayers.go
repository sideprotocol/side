package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateNonBtcRelayers = "update_non_btc_relayers"

func NewMsgUpdateNonBtcRelayers(
	sender string,
	relayers []string,
) *MsgUpdateNonBtcRelayers {
	return &MsgUpdateNonBtcRelayers{
		Sender:   sender,
		Relayers: relayers,
	}
}

func (msg *MsgUpdateNonBtcRelayers) Route() string {
	return RouterKey
}

func (msg *MsgUpdateNonBtcRelayers) Type() string {
	return TypeMsgUpdateNonBtcRelayers
}

func (msg *MsgUpdateNonBtcRelayers) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgUpdateNonBtcRelayers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateNonBtcRelayers) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Relayers) == 0 {
		return sdkerrors.Wrapf(ErrInvalidRelayers, "relayers can not be empty")
	}

	for _, relayer := range msg.Relayers {
		_, err := sdk.AccAddressFromBech32(relayer)
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid relayer address (%s)", err)
		}
	}

	return nil
}
