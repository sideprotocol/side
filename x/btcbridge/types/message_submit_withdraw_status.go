package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitWithdrawStatus = "submit_withdraw_status"

func NewMsgSubmitWithdrawStatus(
	sender string,
	txid string,
	status WithdrawStatus,
) *MsgSubmitWithdrawStatus {
	return &MsgSubmitWithdrawStatus{
		Sender: sender,
		Txid:   txid,
		Status: status,
	}
}

func (msg *MsgSubmitWithdrawStatus) Route() string {
	return RouterKey
}

func (msg *MsgSubmitWithdrawStatus) Type() string {
	return TypeMsgSubmitWithdrawStatus
}

func (msg *MsgSubmitWithdrawStatus) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitWithdrawStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitWithdrawStatus) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid Sender address (%s)", err)
	}

	if len(msg.Txid) == 0 {
		return sdkerrors.Wrap(ErrWithdrawRequestNotExist, "txid cannot be empty")
	}

	if msg.Status != WithdrawStatus_WITHDRAW_STATUS_BROADCASTED {
		return sdkerrors.Wrap(ErrInvalidStatus, "invalid status")
	}

	return nil
}
