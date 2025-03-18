package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitNonce{}

func NewMsgSubmitNonce(
	sender string,
	eventType DlcEventType,
	nonce string,
	oraclePubKey string,
	signature string,
) *MsgSubmitNonce {
	return &MsgSubmitNonce{
		Sender:       sender,
		EventType:    eventType,
		Nonce:        nonce,
		OraclePubkey: oraclePubKey,
		Signature:    signature,
	}
}

// ValidateBasic performs basic MsgSubmitNonce message validation.
func (m *MsgSubmitNonce) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if m.EventType == DlcEventType_UNSPECIFIED {
		return ErrInvalidEventType
	}

	nonceBytes, err := hex.DecodeString(m.Nonce)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidNonce, "failed to decode nonce")
	}

	if _, err := schnorr.ParsePubKey(nonceBytes); err != nil {
		return ErrInvalidNonce
	}

	oraclePk, err := hex.DecodeString(m.OraclePubkey)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "failed to decode oracle pub key")
	}

	if _, err := schnorr.ParsePubKey(oraclePk); err != nil {
		return ErrInvalidPubKey
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidSignature, "failed to decode signature")
	}

	if _, err := schnorr.ParseSignature(sigBytes); err != nil {
		return ErrInvalidSignature
	}

	return nil
}
