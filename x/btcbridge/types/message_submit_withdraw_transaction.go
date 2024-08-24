package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitWithdrawTransaction = "submit_withdraw_transaction"

func NewMsgSubmitWithdrawTransaction(
	sender string,
	blockhash string,
	transaction string,
	proof []string,
) *MsgSubmitWithdrawTransaction {
	return &MsgSubmitWithdrawTransaction{
		Sender:    sender,
		Blockhash: blockhash,
		TxBytes:   transaction,
		Proof:     proof,
	}
}

func (msg *MsgSubmitWithdrawTransaction) Route() string {
	return RouterKey
}

func (msg *MsgSubmitWithdrawTransaction) Type() string {
	return TypeMsgSubmitDepositTransaction
}

func (msg *MsgSubmitWithdrawTransaction) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitWithdrawTransaction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitWithdrawTransaction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Blockhash) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "blockhash cannot be empty")
	}

	if len(msg.TxBytes) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "transaction cannot be empty")
	}

	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "proof cannot be empty")
	}

	return nil
}
