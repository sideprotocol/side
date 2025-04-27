package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRefreshShares{}

// ValidateBasic performs basic MsgRefreshShares message validation.
func (m *MsgRefreshShares) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	pubKey, err := hex.DecodeString(m.PubKey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode the pub key")
	}

	if _, err := schnorr.ParsePubKey(pubKey); err != nil {
		return ErrInvalidPubKey
	}

	if len(m.Participants) == 0 {
		return errorsmod.Wrap(ErrInvalidParticipants, "participants can not be empty")
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

	return nil
}
