package types

import (
	"encoding/base64"
	"slices"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ sdk.Msg = &MsgInitiateDKG{}

// ValidateBasic performs basic MsgInitiateDKG message validation.
func (m *MsgInitiateDKG) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if len(m.Participants) == 0 || m.Threshold == 0 || m.Threshold > uint32(len(m.Participants)) {
		return ErrInvalidDKGParams
	}

	participants := make(map[string]bool)

	for _, p := range m.Participants {
		if len(p.Moniker) > stakingtypes.MaxMonikerLength {
			return ErrInvalidDKGParams
		}

		if _, err := sdk.ValAddressFromBech32(p.OperatorAddress); err != nil {
			return errorsmod.Wrap(err, "invalid operator address")
		}

		if pubKey, err := base64.StdEncoding.DecodeString(p.ConsensusPubkey); err != nil || len(pubKey) != ed25519.PubKeySize {
			return errorsmod.Wrap(err, "invalid consensus public key")
		}

		if participants[p.ConsensusPubkey] {
			return errorsmod.Wrap(ErrInvalidDKGParams, "duplicate participant")
		}

		participants[p.ConsensusPubkey] = true
	}

	if !slices.Equal(m.VaultTypes, SupportedAssetTypes()) {
		return errorsmod.Wrap(ErrInvalidDKGParams, "incorrect vault types")
	}

	if m.EnableTransfer {
		if m.TargetUtxoNum == 0 {
			return errorsmod.Wrap(ErrInvalidDKGParams, "target number of utxos must be greater than 0")
		}
	}

	return nil
}
