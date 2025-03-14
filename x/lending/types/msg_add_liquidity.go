package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgAddLiquidity{}

func NewMsgAddLiquidity(lender string, poolId string, amount sdk.Coin) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Lender: lender,
		PoolId: poolId,
		Amount: amount,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgAddLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.PoolId) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolId, "empty pool id")
	}

	if !m.Amount.IsValid() || !m.Amount.IsPositive() {
		return errorsmod.Wrap(ErrInvalidAmount, "amount must be positive")
	}

	return nil
}
