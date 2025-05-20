package types

import (
	"encoding/base64"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	// maximum moniker size
	MaxMonikerLength = stakingtypes.MaxMonikerLength

	// minimum DKG participant number
	MinDKGParticipantNum = 3

	// default DKG timeout duration
	DefaultDKGTimeoutDuration = time.Duration(86400) * time.Second // 1 day
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		AllowedDkgParticipants: []DKGParticipant{},
		DkgTimeoutDuration:     DefaultDKGTimeoutDuration,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates params
func (p Params) Validate() error {
	if err := validateDKGParticipants(p.AllowedDkgParticipants); err != nil {
		return err
	}

	if err := validateDKGTimeoutDuration(p.DkgTimeoutDuration); err != nil {
		return err
	}

	return nil
}

// validateDKGParticipants validates the given DKG participants
// Note: the participant is the ed25519 consensus pub key
func validateDKGParticipants(participants []DKGParticipant) error {
	if len(participants) == 0 {
		return nil
	}

	if len(participants) < MinDKGParticipantNum {
		return errorsmod.Wrapf(ErrInvalidParams, "number of participants cannot be less than min participant number %d", MinDKGParticipantNum)
	}

	participantMap := make(map[string]bool)

	for _, p := range participants {
		if len(p.Moniker) > MaxMonikerLength {
			return errorsmod.Wrapf(ErrInvalidParams, "moniker size cannot be greater than %d", MaxMonikerLength)
		}

		consensusPubKey, err := base64.StdEncoding.DecodeString(p.ConsensusPubkey)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidParams, "failed to decode the participant consensus pub key")
		}

		if len(consensusPubKey) != ed25519.PubKeySize {
			return errorsmod.Wrap(ErrInvalidParams, "incorrect participant consensus pub key size")
		}

		if participantMap[p.ConsensusPubkey] {
			return errorsmod.Wrap(ErrInvalidParams, "duplicate participant")
		}

		participantMap[p.ConsensusPubkey] = true
	}

	return nil
}

// validateDKGTimeoutDuration validates the given DKG timeout duration
func validateDKGTimeoutDuration(timeoutDuration time.Duration) error {
	if timeoutDuration <= 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid dkg timeout duration")
	}

	return nil
}
