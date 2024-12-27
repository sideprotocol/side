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
	ErrEmptyPoolId        = errorsmod.Register(ModuleName, 2200, "invalid pool id")
	ErrNotAuthorized      = errorsmod.Register(ModuleName, 2201, "not authorized")
	ErrDuplicatedPoolId   = errorsmod.Register(ModuleName, 2202, "duplicated pool id")
	ErrPootNotExists      = errorsmod.Register(ModuleName, 2203, "pool not exists")
	ErrInactivePool       = errorsmod.Register(ModuleName, 2203, "inactive pool")

	ErrEmptyBorrowerPubkey = errorsmod.Register(ModuleName, 3001, "invalid pubkey of borrower")
	ErrInvalidMaturityTime = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrInvalidFinalTimeout = errorsmod.Register(ModuleName, 3003, "final time great than maturity time")
	ErrInvalidLoanSecret   = errorsmod.Register(ModuleName, 3003, "invalid loan secret")

	ErrEmptyDepositTx = errorsmod.Register(ModuleName, 4001, "invalid deposit tx")
	ErrEmptyPoof      = errorsmod.Register(ModuleName, 4002, "invalid proof")

	ErrEmptyLoanSecret = errorsmod.Register(ModuleName, 5001, "invalid loan secret")

	ErrEmptyAdaptorPoint = errorsmod.Register(ModuleName, 6001, "invalid adaptor point")
	ErrInvalidRepayment  = errorsmod.Register(ModuleName, 6002, "invalid repayment")
)
