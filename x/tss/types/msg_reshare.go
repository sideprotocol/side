package types

import (
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgReshare{}

// ValidateBasic performs basic MsgReshare message validation.
func (m *MsgReshare) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if len(m.RemovedParticipants) == 0 || len(m.NewParticipants) == 0 {
		return errorsmod.Wrap(ErrInvalidParticipants, "removed or new participants cannot be empty")
	}

	participants := append(m.RemovedParticipants, m.NewParticipants...)
	participantMap := make(map[string]bool)

	for _, p := range participants {
		if pubKey, err := base64.StdEncoding.DecodeString(p); err != nil || len(pubKey) != ed25519.PubKeySize {
			return errorsmod.Wrap(ErrInvalidParticipants, "invalid participant consensus pub key")
		}

		if participantMap[p] {
			return errorsmod.Wrap(ErrInvalidParticipants, "duplicate participant")
		}

		participantMap[p] = true
	}

	if m.TimeoutDuration < 0 {
		return ErrInvalidTimeoutDuration
	}

	return nil
}
