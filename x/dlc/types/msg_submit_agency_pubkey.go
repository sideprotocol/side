package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitAgencyPubKey{}

func NewMsgSubmitAgencyPubKey(
	sender string,
	id uint64,
	pubKey string,
	signature string,
) *MsgSubmitAgencyPubKey {
	return &MsgSubmitAgencyPubKey{
		Sender:    sender,
		Id:        id,
		PubKey:    pubKey,
		Signature: signature,
	}
}

// ValidateBasic performs basic MsgSubmitAgencyPubKey message validation.
func (m *MsgSubmitAgencyPubKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	pkBytes, err := hex.DecodeString(m.PubKey)
	if err != nil {
		return ErrInvalidPubKey
	}

	if _, err := btcec.ParsePubKey(pkBytes); err != nil {
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
