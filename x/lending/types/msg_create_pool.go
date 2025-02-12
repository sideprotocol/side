package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
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
	if _, err := sdk.AccAddressFromBech32(m.Creator); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.PoolId) < MinPoolIdLength {
		return errorsmod.Wrap(ErrInvalidPoolId, fmt.Sprintf("minimum length of the pool id is %d", MinPoolIdLength))
	}

	if err := sdk.ValidateDenom(m.LendingAsset); err != nil {
		return ErrInvalidLendingAsset
	}

	return nil
}
