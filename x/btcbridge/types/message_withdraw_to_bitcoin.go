package types

import (
	"strconv"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgWithdrawToBitcoin = "withdraw_to_bitcoin"

func NewMsgWithdrawToBitcoin(
	sender string,
	amount string,
	feeRate string,
) *MsgWithdrawToBitcoin {
	return &MsgWithdrawToBitcoin{
		Sender:  sender,
		Amount:  amount,
		FeeRate: feeRate,
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
		return sdkerrors.Wrapf(err, "invalid Sender address (%s)", err)
	}

	_, err = sdk.ParseCoinNormalized(msg.Amount)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidAmount, "invalid amount %s", msg.Amount)
	}

	feeRate, err := strconv.ParseInt(msg.FeeRate, 10, 64)
	if err != nil {
		return err
	}

	if feeRate <= 0 {
		return sdkerrors.Wrap(ErrInvalidFeeRate, "fee rate must be greater than zero")
	}

	return nil
}
