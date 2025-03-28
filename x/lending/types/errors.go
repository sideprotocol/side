package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1000, "invalid amount")
	ErrInvalidParams = errorsmod.Register(ModuleName, 1001, "invalid params")
	ErrInvalidSender = errorsmod.Register(ModuleName, 1002, "invalid tx sender")

	ErrInvalidPoolId         = errorsmod.Register(ModuleName, 2000, "invalid pool id")
	ErrInvalidLendingAsset   = errorsmod.Register(ModuleName, 2001, "invalid lending asset")
	ErrInvalidPoolConfig     = errorsmod.Register(ModuleName, 2002, "invalid pool config")
	ErrPoolAlreadyExists     = errorsmod.Register(ModuleName, 2003, "pool already exists")
	ErrPoolDoesNotExist      = errorsmod.Register(ModuleName, 2004, "pool does not exist")
	ErrPoolPaused            = errorsmod.Register(ModuleName, 2005, "pool paused")
	ErrPoolNotActive         = errorsmod.Register(ModuleName, 2006, "pool not active")
	ErrInsufficientLiquidity = errorsmod.Register(ModuleName, 2007, "insufficient liquidity")

	ErrInvalidPubKey           = errorsmod.Register(ModuleName, 3001, "invalid pubkey")
	ErrInvalidMaturityTime     = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrDuplicatedVault         = errorsmod.Register(ModuleName, 3003, "duplicated vault address")
	ErrInvalidDCM              = errorsmod.Register(ModuleName, 3004, "invalid dcm")
	ErrEmptyLoanId             = errorsmod.Register(ModuleName, 3005, "empty loan id")
	ErrLoanDoesNotExist        = errorsmod.Register(ModuleName, 3006, "loan does not exist")
	ErrInvalidDepositTx        = errorsmod.Register(ModuleName, 3007, "invalid deposit tx")
	ErrInvalidCET              = errorsmod.Register(ModuleName, 3008, "invalid cet")
	ErrInvalidEvent            = errorsmod.Register(ModuleName, 3009, "invalid event")
	ErrInsufficientCollateral  = errorsmod.Register(ModuleName, 3010, "insufficient collateral")
	ErrLiquidationPriceReached = errorsmod.Register(ModuleName, 3011, "liquidation price reached")
	ErrMaturityTimeReached     = errorsmod.Register(ModuleName, 3012, "maturity time reached")
	ErrFailedToBuildTx         = errorsmod.Register(ModuleName, 3013, "failed to build tx")

	ErrInvalidVault          = errorsmod.Register(ModuleName, 4001, "invalid vault")
	ErrInvalidBlockHash      = errorsmod.Register(ModuleName, 4002, "invalid block hash")
	ErrInvalidProof          = errorsmod.Register(ModuleName, 4003, "invalid proof")
	ErrDepositTxDoesNotExist = errorsmod.Register(ModuleName, 4004, "deposit tx does not exist")

	ErrMismatchedBorrower        = errorsmod.Register(ModuleName, 5001, "mismatched borrower")
	ErrInvalidTx                 = errorsmod.Register(ModuleName, 5002, "invalid tx")
	ErrCancellationDoesNotExist  = errorsmod.Register(ModuleName, 5003, "cancellation does not exist")
	ErrDCMSignaturesAlreadyExist = errorsmod.Register(ModuleName, 5004, "dcm signatures already exist")

	ErrRepaymentAdaptorSigsAlreadyExist = errorsmod.Register(ModuleName, 6001, "repayment adaptor signatures already exist")
	ErrRepaymentAdaptorSigsDoNotExist   = errorsmod.Register(ModuleName, 6002, "repayment adaptor signatures do not exist")
	ErrInvalidAdaptorSignatures         = errorsmod.Register(ModuleName, 6003, "invalid adaptor signatures")
	ErrInvalidAdaptorSignature          = errorsmod.Register(ModuleName, 6004, "invalid adaptor signature")

	ErrLoanNotLiquidated                 = errorsmod.Register(ModuleName, 7001, "loan not liquidated yet")
	ErrLiquidationSignaturesAlreadyExist = errorsmod.Register(ModuleName, 7002, "dcm liquidation signatures already exist")

	ErrInvalidLoanStatus = errorsmod.Register(ModuleName, 8001, "invalid loan status")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 8002, "invalid signatures")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 8003, "invalid signature")
)
