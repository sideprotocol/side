package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/incentive module sentinel errors
var (
	ErrInvalidParams = errorsmod.Register(ModuleName, 1001, "invalid params")
)
