package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrInvalidHeader     = errorsmod.Register(ModuleName, 1100, "invalid block header")
	ErrReorgFailed       = errorsmod.Register(ModuleName, 1101, "failed to reorg chain")
	ErrForkedBlockHeader = errorsmod.Register(ModuleName, 1102, "Invalid forked block header")

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

	ErrInvalidAmount                = errorsmod.Register(ModuleName, 3100, "invalid amount")
	ErrInvalidFeeRate               = errorsmod.Register(ModuleName, 3101, "invalid fee rate")
	ErrAssetNotSupported            = errorsmod.Register(ModuleName, 3102, "asset not supported")
	ErrInvalidWithdrawAmount        = errorsmod.Register(ModuleName, 3103, "invalid withdrawal amount")
	ErrDustOutput                   = errorsmod.Register(ModuleName, 3104, "too small output amount")
	ErrInsufficientUTXOs            = errorsmod.Register(ModuleName, 3105, "insufficient utxos")
	ErrMaxTransactionWeightExceeded = errorsmod.Register(ModuleName, 3106, "maximum transaction weight exceeded")
	ErrFailToSerializePsbt          = errorsmod.Register(ModuleName, 3107, "failed to serialize psbt")
	ErrInvalidSignatures            = errorsmod.Register(ModuleName, 3108, "invalid signatures")
	ErrWithdrawRequestNotExist      = errorsmod.Register(ModuleName, 3109, "withdrawal request does not exist")
	ErrWithdrawRequestConfirmed     = errorsmod.Register(ModuleName, 3110, "withdrawal request has been confirmed")
	ErrWithdrawNotEnabled           = errorsmod.Register(ModuleName, 3111, "withdrawal not enabled")

	ErrUTXODoesNotExist = errorsmod.Register(ModuleName, 4100, "utxo does not exist")
	ErrUTXOLocked       = errorsmod.Register(ModuleName, 4101, "utxo locked")
	ErrUTXOUnlocked     = errorsmod.Register(ModuleName, 4102, "utxo unlocked")

	ErrInvalidRunes  = errorsmod.Register(ModuleName, 5100, "invalid runes")
	ErrInvalidRuneId = errorsmod.Register(ModuleName, 5101, "invalid rune id")

	ErrInvalidParams = errorsmod.Register(ModuleName, 6100, "invalid module params")

	ErrInvalidDKGParams                 = errorsmod.Register(ModuleName, 7100, "invalid dkg params")
	ErrDKGRequestDoesNotExist           = errorsmod.Register(ModuleName, 7101, "dkg request does not exist")
	ErrDKGCompletionRequestExists       = errorsmod.Register(ModuleName, 7102, "dkg completion request already exists")
	ErrInvalidDKGCompletionRequest      = errorsmod.Register(ModuleName, 7103, "invalid dkg completion request")
	ErrUnauthorizedDKGCompletionRequest = errorsmod.Register(ModuleName, 7104, "unauthorized dkg completion request")
)
