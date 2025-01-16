package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgRemoveLiquidity{}

func NewMsgRemoveLiquidity(lender string, amount sdk.Coin) *MsgRemoveLiquidity {
	return &MsgRemoveLiquidity{
		Lender: lender,
		Shares: &amount,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgRemoveLiquidity) ValidateBasic() error {
	if m.Shares.Amount.LTE(math.NewInt(0)) {
		return ErrInvalidLiquidation
	}

	if len(m.Lender) == 0 {
		return ErrEmptySender
	}

	return nil
}
