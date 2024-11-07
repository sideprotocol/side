package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdrawToBitcoin = "withdraw_to_bitcoin"

func NewMsgWithdrawToBitcoin(
	sender string,
	amount string,
) *MsgWithdrawToBitcoin {
	return &MsgWithdrawToBitcoin{
		Sender: sender,
		Amount: amount,
	}
}

func (msg *MsgWithdrawToBitcoin) Route() string {
	return RouterKey
}

func (msg *MsgWithdrawToBitcoin) Type() string {
	return TypeMsgWithdrawToBitcoin
}

func (msg *MsgWithdrawToBitcoin) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgWithdrawToBitcoin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
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
