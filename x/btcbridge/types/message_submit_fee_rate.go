package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const TypeMsgSubmitFeeRate = "submit_fee_rate"

func NewMsgSubmitFeeRate(
	sender string,
	feeRate int64,
) *MsgSubmitFeeRate {
	return &MsgSubmitFeeRate{
		Sender:  sender,
		FeeRate: feeRate,
	}
}

func (msg *MsgSubmitFeeRate) Route() string {
	return RouterKey
}

func (msg *MsgSubmitFeeRate) Type() string {
	return TypeMsgSubmitFeeRate
}

func (msg *MsgSubmitFeeRate) GetSigners() []sdk.AccAddress {
	Sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{Sender}
}

func (msg *MsgSubmitFeeRate) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitFeeRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid sender address (%s)", err)
	}

	if msg.FeeRate <= 0 {
		return ErrInvalidFeeRate
	}

	return nil
}
