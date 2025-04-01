package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/sideprotocol/side/crypto/adaptor"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitSignatures{}

func NewMsgSubmitSignatures(
	sender string,
	id uint64,
	signatures []string,
) *MsgSubmitSignatures {
	return &MsgSubmitSignatures{
		Sender:     sender,
		Id:         id,
		Signatures: signatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.Signatures) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "signatures can not be empty")
	}

	for _, signature := range m.Signatures {
		sigBytes, err := hex.DecodeString(signature)
		if err != nil {
			return errorsmod.Wrap(ErrInvalidSignature, "failed to decode the signature")
		}

		switch len(sigBytes) {
		case SchnorrSignatureSize:
			if _, err := schnorr.ParseSignature(sigBytes); err != nil {
				return errorsmod.Wrap(ErrInvalidSignature, "invalid schnorr signature")
			}

		case SchnorrAdaptorSignatureSize:
			if _, err := adaptor.ParseSignature(sigBytes); err != nil {
				return errorsmod.Wrap(ErrInvalidSignature, "invalid schnorr adaptor signature")
			}

		default:
			return errorsmod.Wrap(ErrInvalidSignature, "invalid signature size")
		}
	}

	return nil
}
