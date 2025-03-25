package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitDCMPubKey{}

func NewMsgSubmitDCMPubKey(
	sender string,
	pubKey string,
	dcmId uint64,
	dcmPubKey string,
	signature string,
) *MsgSubmitDCMPubKey {
	return &MsgSubmitDCMPubKey{
		Sender:    sender,
		PubKey:    pubKey,
		DCMId:     dcmId,
		DCMPubKey: dcmPubKey,
		Signature: signature,
	}
}

// ValidateBasic performs basic MsgSubmitDCMPubKey message validation.
func (m *MsgSubmitDCMPubKey) ValidateBasic() error {
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

	dcmPubKey, err := hex.DecodeString(m.DCMPubKey)
	if err != nil {
		return ErrInvalidPubKey
	}

	if _, err := schnorr.ParsePubKey(dcmPubKey); err != nil {
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
