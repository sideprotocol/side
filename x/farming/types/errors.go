package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidAmount        = errorsmod.Register(ModuleName, 1000, "invalid amount")
	ErrInvalidLockDuration  = errorsmod.Register(ModuleName, 1001, "invalid lock duration")
	ErrPhaseNotStarted      = errorsmod.Register(ModuleName, 1002, "phase not started")
	ErrPhaseEnded           = errorsmod.Register(ModuleName, 1003, "phase already ended")
	ErrStakingDoesNotExist  = errorsmod.Register(ModuleName, 1004, "staking does not exist")
	ErrInvalidStakingStatus = errorsmod.Register(ModuleName, 1005, "invalid staking status")
	ErrLockDurationNotEnded = errorsmod.Register(ModuleName, 1006, "lock duration not ended")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2000, "invalid params")
)
