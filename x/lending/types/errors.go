package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrEmptySender   = errorsmod.Register(ModuleName, 1000, "invalid tx sender")
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1100, "invalid amount")
	ErrInvalidParams = errorsmod.Register(ModuleName, 1101, "invalid params")

	ErrInvalidLiquidation = errorsmod.Register(ModuleName, 2100, "invalid liquidation")
	ErrEmptyPoolId        = errorsmod.Register(ModuleName, 2200, "pool id should not be empty")

	ErrEmptyBorrowerPubkey = errorsmod.Register(ModuleName, 3001, "invalid pubkey of borrower")
	ErrInvalidMaturityTime = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrInvalidFinalTimeout = errorsmod.Register(ModuleName, 3003, "final time great than maturity time")
	ErrInvalidLoanSecret   = errorsmod.Register(ModuleName, 3003, "invalid loan secret")
)
