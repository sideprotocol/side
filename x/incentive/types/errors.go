package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/incentive module sentinel errors
var (
	ErrIncentiveNotEnabled = errorsmod.Register(ModuleName, 1001, "incentive not enabled")
	ErrInvalidParams       = errorsmod.Register(ModuleName, 1002, "invalid params")
)
