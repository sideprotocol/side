package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClaimAll{}

func NewMsgClaimAll(staker string) *MsgClaimAll {
	return &MsgClaimAll{
		Staker: staker,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgClaimAll) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Staker); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	return nil
}
