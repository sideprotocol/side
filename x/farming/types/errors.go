package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPhaseParams   = errorsmod.Register(ModuleName, 1000, "invalid amount")
	ErrInvalidAmount        = errorsmod.Register(ModuleName, 1001, "invalid amount")
	ErrInvalidLockDuration  = errorsmod.Register(ModuleName, 1002, "invalid lock duration")
	ErrPhaseDoesNotExist    = errorsmod.Register(ModuleName, 1003, "phase does not exist")
	ErrPhaseNotStarted      = errorsmod.Register(ModuleName, 1004, "phase not started")
	ErrPhaseEnded           = errorsmod.Register(ModuleName, 1005, "phase ended")
	ErrAssetNotAllowed      = errorsmod.Register(ModuleName, 1006, "asset not allowed")
	ErrUnauthorized         = errorsmod.Register(ModuleName, 1007, "unauthorized")
	ErrStakingDoesNotExist  = errorsmod.Register(ModuleName, 1008, "staking does not exist")
	ErrInvalidStakingStatus = errorsmod.Register(ModuleName, 1009, "invalid staking status")
	ErrLockDurationNotEnded = errorsmod.Register(ModuleName, 1010, "lock duration not ended")
	ErrNoPendingReward      = errorsmod.Register(ModuleName, 1011, "no pending reward")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2000, "invalid params")
)
