package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateTrustedNonBtcRelayers = "update_trusted_non_btc_relayers"

func NewMsgUpdateTrustedNonBtcRelayers(
	sender string,
	relayers []string,
) *MsgUpdateTrustedNonBtcRelayers {
	return &MsgUpdateTrustedNonBtcRelayers{
		Sender:   sender,
		Relayers: relayers,
	}
}

func (msg *MsgUpdateTrustedNonBtcRelayers) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTrustedNonBtcRelayers) Type() string {
	return TypeMsgUpdateTrustedNonBtcRelayers
}

func (msg *MsgUpdateTrustedNonBtcRelayers) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgUpdateTrustedNonBtcRelayers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTrustedNonBtcRelayers) ValidateBasic() error {
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
