package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgWithdrawToBitcoin{}

func NewMsgWithdrawToBitcoin(
	sender string,
	amount string,
) *MsgWithdrawToBitcoin {
	return &MsgWithdrawToBitcoin{
		Sender: sender,
		Amount: amount,
	}
}

func (msg *MsgWithdrawToBitcoin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if !IsValidBtcAddress(msg.Sender) {
		return ErrInvalidBtcAddress
	}

	_, err = sdk.ParseCoinNormalized(msg.Amount)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid withdrawal amount")
	}

	return nil
}
