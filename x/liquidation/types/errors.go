package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidSender            = errorsmod.Register(ModuleName, 1100, "invalid sender")
	ErrInvalidAmount            = errorsmod.Register(ModuleName, 1101, "invalid amount")
	ErrLiquidationDoesNotExist  = errorsmod.Register(ModuleName, 1102, "liquidation does not exist")
	ErrInvalidLiquidationStatus = errorsmod.Register(ModuleName, 1103, "invalid liquidation status")
	ErrInvalidPrice             = errorsmod.Register(ModuleName, 1104, "invalid price")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")

	ErrFailedToBuildTx   = errorsmod.Register(ModuleName, 3100, "failed to build transaction")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 3101, "invalid signatures")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 3102, "invalid signature")
)
