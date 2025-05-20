package types

import (
	"encoding/base64"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCompleteRefreshing{}

func NewMsgCompleteRefreshing(
	sender string,
	id uint64,
	consensusPubKey string,
	signature string,
) *MsgCompleteRefreshing {
	return &MsgCompleteRefreshing{
		Sender:          sender,
		Id:              id,
		ConsensusPubkey: consensusPubKey,
		Signature:       signature,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgCompleteRefreshing) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
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
