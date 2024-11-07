package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitDepositTransaction{}

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

func (msg *MsgSubmitDepositTransaction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Blockhash) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "blockhash cannot be empty")
	}

	if len(msg.PrevTxBytes) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "previous transaction cannot be empty")
	}

	if len(msg.TxBytes) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "transaction cannot be empty")
	}

	if len(msg.Proof) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "proof cannot be empty")
	}

	return nil
}
