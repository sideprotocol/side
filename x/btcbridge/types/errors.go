package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrInvalidParams = errorsmod.Register(ModuleName, 1000, "invalid params")

	ErrBlockNotFound             = errorsmod.Register(ModuleName, 2101, "block not found")
	ErrTransactionNotIncluded    = errorsmod.Register(ModuleName, 2102, "transaction not included in block")
	ErrNotConfirmed              = errorsmod.Register(ModuleName, 2103, "transaction not confirmed")
	ErrExceedMaxAcceptanceDepth  = errorsmod.Register(ModuleName, 2104, "exceed max acceptance block depth")
	ErrUnsupportedScriptType     = errorsmod.Register(ModuleName, 2105, "unsupported script type")
	ErrInvalidBtcTransaction     = errorsmod.Register(ModuleName, 2106, "invalid bitcoin transaction")
	ErrTransactionAlreadyMinted  = errorsmod.Register(ModuleName, 2107, "transaction already minted")
	ErrInvalidDepositTransaction = errorsmod.Register(ModuleName, 2108, "invalid deposit transaction")
	ErrInvalidDepositAmount      = errorsmod.Register(ModuleName, 2109, "invalid deposit amount")
	ErrDepositNotEnabled         = errorsmod.Register(ModuleName, 2110, "deposit not enabled")
	ErrUntrustedNonBtcRelayer    = errorsmod.Register(ModuleName, 2111, "untrusted non btc relayer")
	ErrUntrustedFeeProvider      = errorsmod.Register(ModuleName, 2112, "untrusted fee provider")

	ErrInvalidWithdrawAmount        = errorsmod.Register(ModuleName, 3100, "invalid withdrawal amount")
	ErrInvalidBtcAddress            = errorsmod.Register(ModuleName, 3101, "invalid btc address")
	ErrAssetNotSupported            = errorsmod.Register(ModuleName, 3102, "asset not supported")
	ErrInvalidFeeRate               = errorsmod.Register(ModuleName, 3103, "invalid fee rate")
	ErrDustOutput                   = errorsmod.Register(ModuleName, 3104, "too small output amount")
	ErrInsufficientUTXOs            = errorsmod.Register(ModuleName, 3105, "insufficient utxos")
	ErrMaxTransactionWeightExceeded = errorsmod.Register(ModuleName, 3106, "maximum transaction weight exceeded")
	ErrMaxUTXONumExceeded           = errorsmod.Register(ModuleName, 3107, "maximum utxo number exceeded")
	ErrFailToSerializePsbt          = errorsmod.Register(ModuleName, 3108, "failed to serialize psbt")
	ErrInvalidTxHash                = errorsmod.Register(ModuleName, 3109, "invalid tx hash")
	ErrInvalidSignatures            = errorsmod.Register(ModuleName, 3110, "invalid signatures")
	ErrInvalidSignature             = errorsmod.Register(ModuleName, 3111, "invalid signature")
	ErrSigningRequestDoesNotExist   = errorsmod.Register(ModuleName, 3112, "signing request does not exist")
	ErrSigningRequestConfirmed      = errorsmod.Register(ModuleName, 3113, "signing request has been confirmed")
	ErrWithdrawNotEnabled           = errorsmod.Register(ModuleName, 3114, "withdrawal not enabled")
	ErrRateLimitReached             = errorsmod.Register(ModuleName, 3115, "rate limit reached")

	ErrUTXODoesNotExist = errorsmod.Register(ModuleName, 4100, "utxo does not exist")
	ErrUTXOLocked       = errorsmod.Register(ModuleName, 4101, "utxo locked")
	ErrUTXOUnlocked     = errorsmod.Register(ModuleName, 4102, "utxo unlocked")

	ErrInvalidRunes  = errorsmod.Register(ModuleName, 5100, "invalid runes")
	ErrInvalidRuneId = errorsmod.Register(ModuleName, 5101, "invalid rune id")

	ErrInvalidRelayers     = errorsmod.Register(ModuleName, 6100, "invalid relayers")
	ErrInvalidFeeProviders = errorsmod.Register(ModuleName, 6101, "invalid fee providers")

	ErrInvalidDKGParams                 = errorsmod.Register(ModuleName, 7100, "invalid dkg params")
	ErrDKGRequestDoesNotExist           = errorsmod.Register(ModuleName, 7101, "dkg request does not exist")
	ErrDKGCompletionRequestExists       = errorsmod.Register(ModuleName, 7102, "dkg completion request already exists")
	ErrInvalidDKGCompletionRequest      = errorsmod.Register(ModuleName, 7103, "invalid dkg completion request")
	ErrUnauthorizedDKGCompletionRequest = errorsmod.Register(ModuleName, 7104, "unauthorized dkg completion request")
	ErrInvalidVaultVersion              = errorsmod.Register(ModuleName, 7105, "invalid vault version")
	ErrInvalidVault                     = errorsmod.Register(ModuleName, 7106, "invalid vault")
	ErrVaultDoesNotExist                = errorsmod.Register(ModuleName, 7107, "vault does not exist")
	ErrInvalidPsbt                      = errorsmod.Register(ModuleName, 7108, "invalid psbt")

	ErrInvalidConsolidation = errorsmod.Register(ModuleName, 8100, "invalid consolidation")

	ErrInvalidDKGs                       = errorsmod.Register(ModuleName, 9000, "invalid dkgs")
	ErrInvalidParticipants               = errorsmod.Register(ModuleName, 9001, "invalid participants")
	ErrInvalidTimeoutDuration            = errorsmod.Register(ModuleName, 9002, "invalid timeout duration")
	ErrInvalidPubKey                     = errorsmod.Register(ModuleName, 9003, "invalid pub key")
	ErrInvalidDKGStatus                  = errorsmod.Register(ModuleName, 9004, "invalid dkg status")
	ErrRefreshingRequestDoesNotExist     = errorsmod.Register(ModuleName, 9005, "refreshing request does not exist")
	ErrInvalidRefreshingStatus           = errorsmod.Register(ModuleName, 9006, "invalid refreshing status")
	ErrRefreshingRequestExpired          = errorsmod.Register(ModuleName, 9007, "refreshing request expired")
	ErrUnauthorizedParticipant           = errorsmod.Register(ModuleName, 9008, "unauthorized participant")
	ErrRefreshingCompletionAlreadyExists = errorsmod.Register(ModuleName, 9009, "refreshing completion already exists")

	ErrInvalidIBCTransferScript = errorsmod.Register(ModuleName, 10000, "invalid deposit script for IBC transfer")
)
