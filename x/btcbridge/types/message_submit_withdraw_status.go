package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitWithdrawStatus = "submit_withdraw_status"

func NewMsgSubmitWithdrawStatusRequest(
	sender string,
	sequence uint64,
	txid string,
	status WithdrawStatus,
) *MsgSubmitWithdrawStatusRequest {
	return &MsgSubmitWithdrawStatusRequest{
		Sender:   sender,
		Sequence: sequence,
		Txid:     txid,
		Status:   status,
	}
}

func (msg *MsgSubmitWithdrawStatusRequest) Route() string {
	return RouterKey
}

func (msg *MsgSubmitWithdrawStatusRequest) Type() string {
	return TypeMsgSubmitWithdrawStatus
}

func (msg *MsgSubmitWithdrawStatusRequest) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitWithdrawStatusRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitWithdrawStatusRequest) ValidateBasic() error {
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
