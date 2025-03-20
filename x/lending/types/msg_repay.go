package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRepay{}

func NewMsgRepay(borrower string, loanId string) *MsgRepay {
	return &MsgRepay{
		Borrower: borrower,
		LoanId:   loanId,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRepay) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	return nil
}
