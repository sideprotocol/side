package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/incentive module sentinel errors
var (
	ErrDepositIncentiveNotEnabled  = errorsmod.Register(ModuleName, 1001, "incentive not enabled for deposit")
	ErrWithdrawIncentiveNotEnabled = errorsmod.Register(ModuleName, 1002, "incentive not enabled for withdrawal")
	ErrInvalidParams               = errorsmod.Register(ModuleName, 1003, "invalid params")
)
