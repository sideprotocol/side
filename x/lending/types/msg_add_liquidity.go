package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgAddLiquidity{}

func NewMsgAddLiquidity(lender string, poolId string, amount sdk.Coin) *MsgAddLiquidity {
	return &MsgAddLiquidity{
		Lender: lender,
		PoolId: poolId,
		Amount: &amount,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgAddLiquidity) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Lender); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.PoolId) < MinPoolIdLength {
		return errorsmod.Wrap(ErrInvalidPoolId, fmt.Sprintf("minimum length of the pool id is %d", MinPoolIdLength))
	}

	if m.Amount.Amount.LTE(math.NewInt(0)) {
		return ErrInvalidLiquidity
	}

	return nil
}
