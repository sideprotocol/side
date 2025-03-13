package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgCreatePool{}

// ValidateBasic performs basic message validation.
func (m *MsgCreatePool) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid sender address")
	}

	if len(m.Id) < MinPoolIdLength {
		return errorsmod.Wrap(ErrInvalidPoolId, fmt.Sprintf("minimum length of the pool id is %d", MinPoolIdLength))
	}

	if err := sdk.ValidateDenom(m.LendingAsset); err != nil {
		return ErrInvalidLendingAsset
	}

	return ValidatePoolConfig(m.Config)
}
