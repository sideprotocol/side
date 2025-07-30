package types

import (
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreateDCM{}

// ValidateBasic performs basic MsgCreateDCM message validation.
func (m *MsgCreateDCM) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if m.Threshold == 0 {
		return errorsmod.Wrap(ErrInvalidThreshold, "threshold must be greater than 0")
	}

	if len(m.Participants) == 0 || len(m.Participants) < int(m.Threshold) {
		return errorsmod.Wrap(ErrInvalidParticipants, "incorrect participant length")
	}

	participants := make(map[string]bool)

	for _, p := range m.Participants {
		if pubKey, err := base64.StdEncoding.DecodeString(p); err != nil || len(pubKey) != ed25519.PubKeySize {
			return errorsmod.Wrap(err, "invalid participant public key")
		}

		if participants[p] {
			return errorsmod.Wrap(ErrInvalidParticipants, "duplicate participant")
		}

		participants[p] = true
	}

	if m.TimeoutDuration < 0 {
		return ErrInvalidTimeoutDuration
	}

	return nil
}
