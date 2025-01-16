package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClose{}

func NewMsgClose(relayer string, loanId string, signature string) *MsgClose {
	return &MsgClose{
		Relayer:   relayer,
		LoanId:    loanId,
		Signature: signature,
	}
}

// ValidateBasic performs basic message validation.
func (m *MsgClose) ValidateBasic() error {
	if len(m.Relayer) == 0 {
		return ErrEmptySender
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if len(m.Signature) == 0 {
		return ErrInvalidSignature
	}

	return nil
}
