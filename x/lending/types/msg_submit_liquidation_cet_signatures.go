package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitLiquidationCetSignatures{}

func NewMsgSubmitLiquidationCetSignatures(sender string, loanId string, signatures []string) *MsgSubmitLiquidationCetSignatures {
	return &MsgSubmitLiquidationCetSignatures{
		Sender:     sender,
		LoanId:     loanId,
		Signatures: signatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitLiquidationCetSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if len(m.Signatures) == 0 {
		return errorsmod.Wrap(ErrInvalidSignature, "signatures can not be empty")
	}

	for _, sig := range m.Signatures {
		sigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return ErrInvalidSignature
		}

		if _, err := schnorr.ParseSignature(sigBytes); err != nil {
			return ErrInvalidSignature
		}
	}

	return nil
}
