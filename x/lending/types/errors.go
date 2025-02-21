package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidAmount = errorsmod.Register(ModuleName, 1000, "invalid amount")
	ErrInvalidParams = errorsmod.Register(ModuleName, 1001, "invalid params")
	ErrInvalidSender = errorsmod.Register(ModuleName, 1002, "invalid tx sender")

	ErrInvalidLiquidity    = errorsmod.Register(ModuleName, 2100, "invalid liquidity")
	ErrInvalidPoolId       = errorsmod.Register(ModuleName, 2200, "invalid pool id")
	ErrInvalidLendingAsset = errorsmod.Register(ModuleName, 2201, "invalid lending asset")
	ErrNotAuthorized       = errorsmod.Register(ModuleName, 2202, "not authorized")
	ErrDuplicatedPoolId    = errorsmod.Register(ModuleName, 2203, "duplicated pool id")
	ErrPoolDoesNotExist    = errorsmod.Register(ModuleName, 2204, "pool does not exist")
	ErrInactivePool        = errorsmod.Register(ModuleName, 2205, "inactive pool")

	ErrInvalidBorrowerPubkey  = errorsmod.Register(ModuleName, 3001, "invalid pubkey of borrower")
	ErrInvalidMaturityTime    = errorsmod.Register(ModuleName, 3002, "maturity time great than 0")
	ErrInvalidFinalTimeout    = errorsmod.Register(ModuleName, 3003, "final time great than maturity time")
	ErrInvalidLoanSecret      = errorsmod.Register(ModuleName, 3004, "invalid loan secret")
	ErrInvalidDepositTx       = errorsmod.Register(ModuleName, 3005, "invalid deposit tx")
	ErrDuplicatedVault        = errorsmod.Register(ModuleName, 3006, "duplicated vault address")
	ErrInvalidAgency          = errorsmod.Register(ModuleName, 3007, "invalid agency")
	ErrInvalidPriceEvent      = errorsmod.Register(ModuleName, 3008, "invalid price event")
	ErrInvalidFunding         = errorsmod.Register(ModuleName, 3009, "invalid funding")
	ErrInvalidCET             = errorsmod.Register(ModuleName, 3010, "invalid cet")
	ErrInsufficientCollateral = errorsmod.Register(ModuleName, 3011, "insufficient collateral")
	ErrLoanDoesNotExist       = errorsmod.Register(ModuleName, 3012, "loan does not exist")
	ErrFailedToBuildTx        = errorsmod.Register(ModuleName, 3013, "failed to build tx")

	ErrInvalidDepositTxHash  = errorsmod.Register(ModuleName, 4001, "invalid deposit tx hash")
	ErrInvalidBlockHash      = errorsmod.Register(ModuleName, 4002, "invalid block hash")
	ErrInvalidProof          = errorsmod.Register(ModuleName, 4003, "invalid proof")
	ErrDepositTxDoesNotExist = errorsmod.Register(ModuleName, 4004, "deposit tx does not exist")

	ErrMismatchedBorrower    = errorsmod.Register(ModuleName, 5001, "mismatched borrower")
	ErrInvalidLoanSecretHash = errorsmod.Register(ModuleName, 5002, "invalid loan secret hash")
	ErrMismatchedLoanSecret  = errorsmod.Register(ModuleName, 5003, "mismatched loan secret")

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
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 8002, "invalid signature")
)
