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

	ErrInsufficientUTXOs            = errorsmod.Register(ModuleName, 3100, "insufficient utxos")
	ErrMaxTransactionWeightExceeded = errorsmod.Register(ModuleName, 3101, "maximum transaction weight exceeded")
	ErrFailedToBuildTx              = errorsmod.Register(ModuleName, 3102, "failed to build transaction")
)
