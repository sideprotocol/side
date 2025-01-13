package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateTrustedFeeProviders{}

func NewMsgUpdateTrustedFeeProviders(
	sender string,
	feeProviders []string,
) *MsgUpdateTrustedFeeProviders {
	return &MsgUpdateTrustedFeeProviders{
		Sender:       sender,
		FeeProviders: feeProviders,
	}
}

func (msg *MsgUpdateTrustedFeeProviders) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return errorsmod.Wrapf(err, "invalid sender address (%s)", err)
	}

	if len(msg.FeeProviders) == 0 {
		return errorsmod.Wrapf(ErrInvalidFeeProviders, "fee providers can not be empty")
	}

	for _, provider := range msg.FeeProviders {
		_, err := sdk.AccAddressFromBech32(provider)
		if err != nil {
			return errorsmod.Wrapf(err, "invalid fee provider address (%s)", err)
		}
	}

	return nil
}
