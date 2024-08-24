package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitDepositTransaction = "submit_deposit_transaction"

func NewMsgSubmitDepositTransaction(
	sender string,
	blockhash string,
	prevTx string,
	tx string,
	proof []string,
) *MsgSubmitDepositTransaction {
	return &MsgSubmitDepositTransaction{
		Sender:      sender,
		Blockhash:   blockhash,
		PrevTxBytes: prevTx,
		TxBytes:     tx,
		Proof:       proof,
	}
}

func (msg *MsgSubmitDepositTransaction) Route() string {
	return RouterKey
}

func (msg *MsgSubmitDepositTransaction) Type() string {
	return TypeMsgSubmitDepositTransaction
}

func (msg *MsgSubmitDepositTransaction) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitDepositTransaction) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitDepositTransaction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Blockhash) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "blockhash cannot be empty")
	}

	if len(msg.PrevTxBytes) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "previous transaction cannot be empty")
	}

	if len(msg.TxBytes) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "transaction cannot be empty")
	}

	if len(msg.Proof) == 0 {
		return sdkerrors.Wrap(ErrInvalidBtcTransaction, "proof cannot be empty")
	}

	return nil
}
