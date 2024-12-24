package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitAttestation{}

func NewMsgSubmitAttestation(
	sender string,
	eventId uint64,
	signature string,
) *MsgSubmitAttestation {
	return &MsgSubmitAttestation{
		Sender:    sender,
		EventId:   eventId,
		Signature: signature,
	}
}

// ValidateBasic performs basic MsgSubmitAttestation message validation.
func (m *MsgSubmitAttestation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
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
