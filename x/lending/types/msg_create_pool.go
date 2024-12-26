package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreatePool{}

func NewMsgCreatePool(creator string, denom string) *MsgCreatePool {
	return &MsgCreatePool{
		Creator:      creator,
		LendingDenom: denom,
	}
}

// ValidateBasic performs basic MsgAddLiquidity message validation.
func (m *MsgCreatePool) ValidateBasic() error {
	if len(m.Creator) == 0 {
		return ErrEmptySender
	}

	if len(m.LendingDenom) == 0 {
		return ErrInvalidLengthParams
	}

	return nil
}
