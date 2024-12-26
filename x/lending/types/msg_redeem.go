package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRedeem{}

func NewMsgRedeem(borrower string, loanSecret string) *MsgRedeem {
	return &MsgRedeem{
		Borrower:   borrower,
		LoanSecret: loanSecret,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRedeem) ValidateBasic() error {
	if len(m.Borrower) == 0 {
		return ErrEmptySender
	}

	if len(m.LoanSecret) == 0 {
		return ErrEmptyLoanSecret
	}

	return nil
}
