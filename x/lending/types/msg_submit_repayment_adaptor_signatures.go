package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitRepaymentAdaptorSignatures{}

func NewMsgSubmitRepaymentAdaptorSignatures(sender string, loanId string, adaptorSignatures []string) *MsgSubmitRepaymentAdaptorSignatures {
	return &MsgSubmitRepaymentAdaptorSignatures{
		Sender:            sender,
		LoanId:            loanId,
		AdaptorSignatures: adaptorSignatures,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitRepaymentAdaptorSignatures) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if len(m.AdaptorSignatures) == 0 {
		return errorsmod.Wrap(ErrInvalidAdaptorSignatures, "empty adaptor signatures")
	}

	for _, sig := range m.AdaptorSignatures {
		adaptorSigBytes, err := hex.DecodeString(sig)
		if err != nil {
			return ErrInvalidAdaptorSignature
		}

		if _, err := schnorr.ParseSignature(adaptorSigBytes); err != nil {
			return ErrInvalidAdaptorSignature
		}
	}

	return nil
}
