package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrEmptySender   = errorsmod.Register(ModuleName, 1000, "invalid tx sender")
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1100, "invalid amount")
	ErrInvalidParams = errorsmod.Register(ModuleName, 1101, "invalid params")
	ErrInvalidSender = errorsmod.Register(ModuleName, 1002, "invalid tx sender")

	ErrInvalidLiquidation = errorsmod.Register(ModuleName, 2100, "invalid liquidation")
	ErrEmptyPoolId        = errorsmod.Register(ModuleName, 2200, "invalid pool id")
	ErrNotAuthorized      = errorsmod.Register(ModuleName, 2201, "not authorized")
	ErrDuplicatedPoolId   = errorsmod.Register(ModuleName, 2202, "duplicated pool id")
	ErrPootNotExists      = errorsmod.Register(ModuleName, 2203, "pool not exists")
	ErrInactivePool       = errorsmod.Register(ModuleName, 2203, "inactive pool")

	ErrEmptyBorrowerPubkey    = errorsmod.Register(ModuleName, 3001, "invalid pubkey of borrower")
	ErrInvalidMaturityTime    = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrInvalidFinalTimeout    = errorsmod.Register(ModuleName, 3003, "final time great than maturity time")
	ErrInvalidLoanSecret      = errorsmod.Register(ModuleName, 3003, "invalid loan secret")
	ErrDuplicatedVault        = errorsmod.Register(ModuleName, 3004, "duplicated vault address")
	ErrInvalidPriceEvent      = errorsmod.Register(ModuleName, 3005, "invalid price event")
	ErrInvalidFunding         = errorsmod.Register(ModuleName, 3006, "invalid funding")
	ErrInvalidCET             = errorsmod.Register(ModuleName, 3007, "invalid cet")
	ErrInsufficientCollateral = errorsmod.Register(ModuleName, 3008, "insufficient collateral")
	ErrLoanNotExists          = errorsmod.Register(ModuleName, 3009, "loan not exists")

	ErrEmptyDepositTx     = errorsmod.Register(ModuleName, 4001, "invalid deposit tx")
	ErrInvalidProof       = errorsmod.Register(ModuleName, 4002, "invalid proof")
	ErrDepositTxNotExists = errorsmod.Register(ModuleName, 4002, "deposit not exists")

	ErrMismatchedBorrower = errorsmod.Register(ModuleName, 5001, "mismatched borrower")
	ErrEmptyLoanSecret    = errorsmod.Register(ModuleName, 5002, "invalid loan secret")
	ErrMismatchLoanSecret = errorsmod.Register(ModuleName, 5003, "mismatch loan secret")

	ErrEmptyAdaptorPoint      = errorsmod.Register(ModuleName, 6001, "invalid adaptor point")
	ErrInvalidRepayment       = errorsmod.Register(ModuleName, 6002, "invalid repayment")
	ErrInvalidRepaymentTx     = errorsmod.Register(ModuleName, 6003, "invalid repayment tx")
	ErrInvalidRepaymentSecret = errorsmod.Register(ModuleName, 6003, "invalid repayment secret")
)
