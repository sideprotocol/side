package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreatePool{}

func NewMsgCreatePool(creator string, poolId string, lendingAsset string) *MsgCreatePool {
	return &MsgCreatePool{
		Creator:      creator,
		PoolId:       poolId,
		LendingAsset: lendingAsset,
	}
}

// ValidateBasic performs basic MsgCreatePool message validation.
func (m *MsgCreatePool) ValidateBasic() error {
	if len(m.PoolId) < 2 {
		return ErrEmptyPoolId
	}
	if len(m.Creator) == 0 {
		return ErrEmptySender
	}
	if len(m.LendingAsset) == 0 {
		return ErrInvalidLengthParams
	}

	return nil
}
