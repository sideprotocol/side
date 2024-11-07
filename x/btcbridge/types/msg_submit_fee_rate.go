package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitFeeRate{}

func NewMsgSubmitFeeRate(
	sender string,
	feeRate int64,
) *MsgSubmitFeeRate {
	return &MsgSubmitFeeRate{
		Sender:  sender,
		FeeRate: feeRate,
	}
}

func (msg *MsgSubmitFeeRate) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if msg.FeeRate <= 0 {
		return ErrInvalidFeeRate
	}

	return nil
}
