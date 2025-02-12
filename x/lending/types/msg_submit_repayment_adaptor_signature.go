package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitRepaymentAdaptorSignature{}

func NewMsgSubmitRepaymentAdaptorSignature(relayer string, loanId string, adaptorSignature string) *MsgSubmitRepaymentAdaptorSignature {
	return &MsgSubmitRepaymentAdaptorSignature{
		Relayer:          relayer,
		LoanId:           loanId,
		AdaptorSignature: adaptorSignature,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgSubmitRepaymentAdaptorSignature) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Relayer); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	adaptorSigBytes, err := hex.DecodeString(m.AdaptorSignature)
	if err != nil {
		return ErrInvalidAdaptorSignature
	}

	if _, err := schnorr.ParseSignature(adaptorSigBytes); err != nil {
		return ErrInvalidAdaptorSignature
	}

	return nil
}
