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
	nonce string,
	signature string,
) *MsgSubmitNonce {
	return &MsgSubmitNonce{
		Sender:    sender,
		Nonce:     nonce,
		Signature: signature,
	}
}

// ValidateBasic performs basic MsgSubmitNonce message validation.
func (m *MsgSubmitNonce) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	nonceBytes, err := hex.DecodeString(m.Nonce)
	if err != nil {
		return ErrInvalidNonce
	}

	if _, err := schnorr.ParsePubKey(nonceBytes); err != nil {
		return ErrInvalidNonce
	}

	sigBytes, err := hex.DecodeString(m.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	if _, err := schnorr.ParseSignature(sigBytes); err != nil {
		return ErrInvalidSignature
	}

	return nil
}
