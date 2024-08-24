package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitWithdrawSignatures = "submit_withdraw_signatures"

func NewMsgSubmitWithdrawSignatures(
	sender string,
	txid string,
	pbst string,
) *MsgSubmitWithdrawSignatures {
	return &MsgSubmitWithdrawSignatures{
		Sender: sender,
		Txid:   txid,
		Psbt:   pbst,
	}
}

func (msg *MsgSubmitWithdrawSignatures) Route() string {
	return RouterKey
}

func (msg *MsgSubmitWithdrawSignatures) Type() string {
	return TypeMsgSubmitWithdrawSignatures
}

func (msg *MsgSubmitWithdrawSignatures) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitWithdrawSignatures) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitWithdrawSignatures) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Txid) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignatures, "txid cannot be empty")
	}

	if len(msg.Psbt) == 0 {
		return sdkerrors.Wrap(ErrInvalidSignatures, "psbt cannot be empty")
	}

	return nil
}
