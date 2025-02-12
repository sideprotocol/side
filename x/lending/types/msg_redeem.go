package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRedeem{}

func NewMsgRedeem(borrower string, loanId string, loanSecret string) *MsgRedeem {
	return &MsgRedeem{
		Borrower:   borrower,
		LoanId:     loanId,
		LoanSecret: loanSecret,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRedeem) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Borrower); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.LoanId) == 0 {
		return ErrEmptyLoanId
	}

	if secretBytes, err := hex.DecodeString(m.LoanSecret); err != nil || len(secretBytes) != LoanSecretLength {
		return ErrInvalidLoanSecret
	}

	return nil
}
