package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgApply{}

func NewMsgRequestVaultAddress(borrower string, borrowerPubkey string, hashLoanSecret string, maturityTime uint64, finalTimeout uint64) *MsgApply {
	return &MsgApply{
		Borrower:         borrower,
		BorrowerPubkey:   borrowerPubkey,
		HashOfLoanSecret: hashLoanSecret,
		MaturityTime:     maturityTime,
		FinalTimeout:     finalTimeout,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgApply) ValidateBasic() error {
	if m.MaturityTime <= 0 {
		return ErrInvalidMaturityTime
	}

	if m.MaturityTime <= m.FinalTimeout {
		return ErrInvalidFinalTimeout
	}

	if len(m.Borrower) == 0 {
		return ErrEmptySender
	}

	if len(m.HashOfLoanSecret) == 0 {
		return ErrInvalidLoanSecret
	}

	if len(m.BorrowerPubkey) == 0 {
		return ErrEmptyBorrowerPubkey
	}

	return nil
}
