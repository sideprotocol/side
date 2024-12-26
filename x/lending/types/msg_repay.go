package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRepay{}

func NewMsgRepay(borrower string, amount sdk.Coin, adaptorPoint string) *MsgRepay {
	return &MsgRepay{
		Borrower:     borrower,
		Amount:       &amount,
		AdaptorPoint: adaptorPoint,
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

	if m.Amount.Amount.LTE(math.NewInt(0)) {
		return ErrInvalidRepayment
	}

	return nil
}
