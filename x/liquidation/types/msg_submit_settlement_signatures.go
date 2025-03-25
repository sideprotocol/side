package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitSettlementSignatures{}

func NewMsgSubmitSettlementSignatures(
	sender string,
	liquidationId uint64,
	signatures []string,
) *MsgSubmitSettlementSignatures {
	return &MsgSubmitSettlementSignatures{
		Sender:        sender,
		LiquidationId: liquidationId,
		Signatures:    signatures,
	}
}

// ValidateBasic performs basic MsgSubmitSettlementSignatures message validation.
func (m *MsgSubmitSettlementSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.Signatures) == 0 {
		return errorsmod.Wrap(ErrInvalidSignatures, "signatures can not be empty")
	}

	for _, sig := range m.Signatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidSignature, "failed to decode signature")
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return errorsmod.Wrapf(ErrInvalidSignature, "%v", err)
		}
	}

	return nil
}
