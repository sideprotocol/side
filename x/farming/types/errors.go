package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidAmount        = errorsmod.Register(ModuleName, 1001, "invalid amount")
	ErrInvalidLockDuration  = errorsmod.Register(ModuleName, 1002, "invalid lock duration")
	ErrFarmingNotEnabled    = errorsmod.Register(ModuleName, 1003, "farming not enabled")
	ErrAssetNotEligible     = errorsmod.Register(ModuleName, 1006, "asset not eligible")
	ErrUnauthorized         = errorsmod.Register(ModuleName, 1007, "unauthorized")
	ErrStakingDoesNotExist  = errorsmod.Register(ModuleName, 1008, "staking does not exist")
	ErrInvalidStakingStatus = errorsmod.Register(ModuleName, 1009, "invalid staking status")
	ErrLockDurationNotEnded = errorsmod.Register(ModuleName, 1010, "lock duration not ended")
	ErrNoPendingRewards     = errorsmod.Register(ModuleName, 1011, "no pending rewards")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2000, "invalid params")
)
