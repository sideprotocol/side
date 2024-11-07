package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateTrustedOracles{}

func NewMsgUpdateTrustedOracles(
	sender string,
	oracles []string,
) *MsgUpdateTrustedOracles {
	return &MsgUpdateTrustedOracles{
		Sender:  sender,
		Oracles: oracles,
	}
}

func (msg *MsgUpdateTrustedOracles) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Oracles) == 0 {
		return errorsmod.Wrapf(ErrInvalidOracles, "oracles can not be empty")
	}

	for _, oracle := range msg.Oracles {
		_, err := sdk.AccAddressFromBech32(oracle)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid oracle address (%s)", err)
		}
	}

	return nil
}
