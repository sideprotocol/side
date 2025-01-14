package types

import (
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
	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if len(m.AdaptorSignature) == 0 {
		return ErrInvalidAdaptorSignature
	}

	return nil
}
