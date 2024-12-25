package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateAgency{}

func NewMsgCreateAgency(
	authority string,
	participants []string,
	threshold uint32,
) *MsgCreateAgency {
	return &MsgCreateAgency{
		Authority:    authority,
		Participants: participants,
		Threshold:    threshold,
	}
}

// ValidateBasic performs basic MsgCreateAgency message validation.
func (m *MsgCreateAgency) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if m.Threshold == 0 {
		return errorsmod.Wrap(ErrInvalidThreshold, "threshold must be greater than 0")
	}

	if len(m.Participants) == 0 || len(m.Participants) < int(m.Threshold) {
		return errorsmod.Wrap(ErrInvalidParticipants, "incorrect participant length")
	}

	for _, p := range m.Participants {
		if _, err := sdk.ConsAddressFromHex(p); err != nil {
			return errorsmod.Wrap(ErrInvalidParticipants, "invalid consensus address")
		}
	}

	return nil
}