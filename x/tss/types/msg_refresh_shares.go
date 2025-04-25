package types

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"

	errorsmod "cosmossdk.io/errors"
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

	return nil
}
