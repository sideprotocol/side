package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidAmount            = errorsmod.Register(ModuleName, 1100, "invalid amount")
	ErrLiquidationDoesNotExist  = errorsmod.Register(ModuleName, 1101, "liquidation does not exist")
	ErrInvalidLiquidationStatus = errorsmod.Register(ModuleName, 1102, "invalid liquidation status")
	ErrInvalidPrice             = errorsmod.Register(ModuleName, 1103, "invalid price")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")

	ErrFailedToBuildTx   = errorsmod.Register(ModuleName, 3100, "failed to build transaction")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 3101, "invalid signatures")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 3102, "invalid signature")
)
