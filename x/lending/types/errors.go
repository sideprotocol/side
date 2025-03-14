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
	ErrInactivePool          = errorsmod.Register(ModuleName, 2005, "inactive pool")
	ErrInsufficientLiquidity = errorsmod.Register(ModuleName, 2006, "insufficient liquidity")

	ErrInvalidBorrowerPubkey  = errorsmod.Register(ModuleName, 3001, "invalid pubkey of borrower")
	ErrInvalidMaturityTime    = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrInvalidDepositTx       = errorsmod.Register(ModuleName, 3003, "invalid deposit tx")
	ErrDuplicatedVault        = errorsmod.Register(ModuleName, 3004, "duplicated vault address")
	ErrInvalidAgency          = errorsmod.Register(ModuleName, 3005, "invalid agency")
	ErrInvalidPriceEvent      = errorsmod.Register(ModuleName, 3006, "invalid price event")
	ErrInvalidFunding         = errorsmod.Register(ModuleName, 3007, "invalid funding")
	ErrInvalidCET             = errorsmod.Register(ModuleName, 3008, "invalid cet")
	ErrInsufficientCollateral = errorsmod.Register(ModuleName, 3009, "insufficient collateral")
	ErrLoanDoesNotExist       = errorsmod.Register(ModuleName, 3010, "loan does not exist")
	ErrFailedToBuildTx        = errorsmod.Register(ModuleName, 3011, "failed to build tx")

	ErrInvalidDepositTxHash  = errorsmod.Register(ModuleName, 4001, "invalid deposit tx hash")
	ErrInvalidBlockHash      = errorsmod.Register(ModuleName, 4002, "invalid block hash")
	ErrInvalidProof          = errorsmod.Register(ModuleName, 4003, "invalid proof")
	ErrDepositTxDoesNotExist = errorsmod.Register(ModuleName, 4004, "deposit tx does not exist")

	ErrMismatchedBorrower        = errorsmod.Register(ModuleName, 5001, "mismatched borrower")
	ErrInvalidTx                 = errorsmod.Register(ModuleName, 5002, "invalid tx")
	ErrCancellationDoesNotExist  = errorsmod.Register(ModuleName, 5003, "cancellation does not exist")
	ErrDcaSignaturesAlreadyExist = errorsmod.Register(ModuleName, 5004, "dca signatures already exist")

	ErrInvalidAdaptorPoint              = errorsmod.Register(ModuleName, 6001, "invalid adaptor point")
	ErrInvalidRepayment                 = errorsmod.Register(ModuleName, 6002, "invalid repayment")
	ErrInvalidRepaymentTx               = errorsmod.Register(ModuleName, 6003, "invalid repayment tx")
	ErrInvalidRepaymentSecret           = errorsmod.Register(ModuleName, 6004, "invalid repayment secret")
	ErrRepaymentAdaptorSigsAlreadyExist = errorsmod.Register(ModuleName, 6005, "repayment adaptor signatures already exist")
	ErrRepaymentAdaptorSigsDoNotExist   = errorsmod.Register(ModuleName, 6006, "repayment adaptor signatures do not exist")
	ErrInvalidAdaptorSignatures         = errorsmod.Register(ModuleName, 6007, "invalid adaptor signatures")
	ErrInvalidAdaptorSignature          = errorsmod.Register(ModuleName, 6008, "invalid adaptor signature")
	ErrEmptyLoanId                      = errorsmod.Register(ModuleName, 6009, "empty loan id")

	ErrLoanNotLiquidated                 = errorsmod.Register(ModuleName, 7001, "loan not liquidated yet")
	ErrLiquidationSignaturesAlreadyExist = errorsmod.Register(ModuleName, 7002, "agency liquidation signatures already exist")
	ErrInvalidLiquidationSignatures      = errorsmod.Register(ModuleName, 7003, "invalid agency liquidation signatures")

	ErrInvalidLoanStatus = errorsmod.Register(ModuleName, 8001, "invalid loan status")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 8002, "invalid signatures")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 8003, "invalid signature")
)
