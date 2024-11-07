package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitWithdrawTransaction{}

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

func (msg *MsgSubmitWithdrawTransaction) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Blockhash) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "blockhash cannot be empty")
	}

	if len(msg.TxBytes) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "transaction cannot be empty")
	}

	if len(msg.Proof) == 0 {
		return errorsmod.Wrap(ErrInvalidBtcTransaction, "proof cannot be empty")
	}

	return nil
}
