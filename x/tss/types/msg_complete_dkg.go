package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCompleteDKG{}

func NewMsgCompleteDKG(
	sender string,
	id uint64,
	pubKeys []string,
	consensusPubKey string,
	signature string,
) *MsgCompleteDKG {
	return &MsgCompleteDKG{
		Sender:          sender,
		Id:              id,
		PubKeys:         pubKeys,
		ConsensusPubkey: consensusPubKey,
		Signature:       signature,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgCompleteDKG) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.PubKeys) == 0 {
		return errorsmod.Wrap(ErrInvalidPubKeys, "pub keys can not be empty")
	}

	for _, pubKey := range m.PubKeys {
		pkBytes, err := hex.DecodeString(pubKey)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode the pub key")
		}

		if _, err := schnorr.ParsePubKey(pkBytes); err != nil {
			return ErrInvalidPubKey
		}
	}

	consensusPubKey, err := base64.StdEncoding.DecodeString(m.ConsensusPubkey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode the consensus pub key")
	}

	if len(consensusPubKey) != ed25519.PubKeySize {
		return errorsmod.Wrap(ErrInvalidPubKey, "incorrect consensus pub key size")
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidSignature, "failed to decode the signature")
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return errorsmod.Wrap(ErrInvalidSignature, "incorrect signature size")
	}

	return nil
}
