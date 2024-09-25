package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgUpdateTrustedOracles = "update_trusted_oracles"

func NewMsgUpdateTrustedOracles(
	sender string,
	oracles []string,
) *MsgUpdateTrustedOracles {
	return &MsgUpdateTrustedOracles{
		Sender:  sender,
		Oracles: oracles,
	}
}

func (msg *MsgUpdateTrustedOracles) Route() string {
	return RouterKey
}

func (msg *MsgUpdateTrustedOracles) Type() string {
	return TypeMsgUpdateTrustedOracles
}

func (msg *MsgUpdateTrustedOracles) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgUpdateTrustedOracles) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateTrustedOracles) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Oracles) == 0 {
		return sdkerrors.Wrapf(ErrInvalidOracles, "oracles can not be empty")
	}

	for _, oracle := range msg.Oracles {
		_, err := sdk.AccAddressFromBech32(oracle)
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid oracle address (%s)", err)
		}
	}

	return nil
}
