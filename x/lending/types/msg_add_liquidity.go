package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgAddLiquidity{}

func NewMsgAddLiquidity(poolId string, lender string, amount sdk.Coin) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		PoolId: poolId,
		Lender: lender,
		Amount: &amount,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgAddLiquidity) ValidateBasic() error {
	if m.Amount.Amount.LTE(math.NewInt(0)) {
		return ErrInvalidLiquidation
	}

	if len(m.PoolId) == 0 {
		return ErrEmptyPoolId
	}

	return nil
}
