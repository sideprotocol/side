package types

import (
	"encoding/base64"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitOraclePubKey{}

func NewMsgSubmitOraclePubKey(
	sender string,
	pubKey string,
	oracleId uint64,
	oraclePubKey string,
	signature string,
) *MsgSubmitOraclePubKey {
	return &MsgSubmitOraclePubKey{
		Sender:       sender,
		PubKey:       pubKey,
		OracleId:     oracleId,
		OraclePubkey: oraclePubKey,
		Signature:    signature,
	}
}

// ValidateBasic performs basic MsgSubmitOraclePubKey message validation.
func (m *MsgSubmitOraclePubKey) ValidateBasic() error {
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

	oraclePubKey, err := hex.DecodeString(m.OraclePubkey)
	if err != nil {
		return ErrInvalidPubKey
	}

	if _, err := schnorr.ParsePubKey(oraclePubKey); err != nil {
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
