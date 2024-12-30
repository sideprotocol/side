package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRepay{}

func NewMsgRepay(borrower string, loanId string, adaptorPoint string, txid string, signature string) *MsgRepay {
	return &MsgRepay{
		Borrower:         borrower,
		AdaptorPoint:     adaptorPoint,
		LoanId:           loanId,
		ClaimTxId:        txid,
		AdaptorSignature: signature,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRepay) ValidateBasic() error {
	if len(m.Borrower) == 0 {
		return ErrEmptySender
	}

	if len(m.AdaptorPoint) == 0 {
		return ErrEmptyAdaptorPoint
	}

	if len(m.LoanId) == 0 {
		return ErrInvalidRepayment
	}

	if len(m.ClaimTxId) == 0 {
		return ErrInvalidRepaymentTx
	}

	if len(m.AdaptorSignature) == 0 {
		return ErrInvalidRepaymentTx
	}

	return nil
}
