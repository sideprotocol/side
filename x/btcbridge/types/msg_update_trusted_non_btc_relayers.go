package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateTrustedNonBtcRelayers{}

func NewMsgUpdateTrustedNonBtcRelayers(
	sender string,
	relayers []string,
) *MsgUpdateTrustedNonBtcRelayers {
	return &MsgUpdateTrustedNonBtcRelayers{
		Sender:   sender,
		Relayers: relayers,
	}
}

func (msg *MsgUpdateTrustedNonBtcRelayers) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.Relayers) == 0 {
		return errorsmod.Wrapf(ErrInvalidRelayers, "relayers can not be empty")
	}

	for _, relayer := range msg.Relayers {
		_, err := sdk.AccAddressFromBech32(relayer)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid relayer address (%s)", err)
		}
	}

	return nil
}
