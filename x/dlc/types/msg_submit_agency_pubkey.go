package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitAgencyPubKey{}

func NewMsgSubmitAgencyPubKey(
	sender string,
	pubKey string,
	agencyId uint64,
	agencyPubKey string,
	signature string,
) *MsgSubmitAgencyPubKey {
	return &MsgSubmitAgencyPubKey{
		Sender:       sender,
		PubKey:       pubKey,
		AgencyId:     agencyId,
		AgencyPubkey: agencyPubKey,
		Signature:    signature,
	}
}

// ValidateBasic performs basic MsgSubmitAgencyPubKey message validation.
func (m *MsgSubmitAgencyPubKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	pubKey, err := base64.StdEncoding.DecodeString(m.PubKey)
	if err != nil {
		return ErrInvalidPubKey
	}

	if len(pubKey) != ed25519.PubKeySize {
		return ErrInvalidPubKey
	}

	agencyPubKey, err := hex.DecodeString(m.AgencyPubkey)
	if err != nil {
		return ErrInvalidPubKey
	}

	if _, err := schnorr.ParsePubKey(agencyPubKey); err != nil {
		return ErrInvalidPubKey
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return ErrInvalidSignature
	}

	return nil
}
