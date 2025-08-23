package types

import (
	"encoding/base64"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRefresh{}

// ValidateBasic performs basic MsgRefresh message validation.
func (m *MsgRefresh) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if len(m.DkgIds) == 0 {
		return errorsmod.Wrap(ErrInvalidDKGs, "dkgs cannot be empty")
	}

	if len(m.RemovedParticipants) == 0 {
		return errorsmod.Wrap(ErrInvalidParticipants, "removed participants cannot be empty")
	}

	participants := make(map[string]bool)

	for _, p := range m.RemovedParticipants {
		if pubKey, err := base64.StdEncoding.DecodeString(p); err != nil || len(pubKey) != ed25519.PubKeySize {
			return errorsmod.Wrap(ErrInvalidParticipants, "invalid participant consensus pub key")
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
