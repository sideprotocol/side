package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitSignatures{}

func NewMsgSubmitSignatures(
	sender string,
	txid string,
	pbst string,
) *MsgSubmitSignatures {
	return &MsgSubmitSignatures{
		Sender: sender,
		Txid:   txid,
		Psbt:   pbst,
	}
}

func (msg *MsgSubmitSignatures) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Txid) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "txid cannot be empty")
	}

	if len(msg.Psbt) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "psbt cannot be empty")
	}

	return nil
}
