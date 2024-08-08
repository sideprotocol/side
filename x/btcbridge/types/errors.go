package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrSenderAddressNotAuthorized = errorsmod.Register(ModuleName, 1000, "sender address not authorized")
	ErrInvalidHeader              = errorsmod.Register(ModuleName, 1100, "invalid block header")
	ErrReorgFailed                = errorsmod.Register(ModuleName, 1101, "failed to reorg chain")
	ErrForkedBlockHeader          = errorsmod.Register(ModuleName, 1102, "Invalid forked block header")

	ErrInvalidSenders = errorsmod.Register(ModuleName, 2100, "invalid allowed senders")

	ErrInvalidBtcTransaction      = errorsmod.Register(ModuleName, 3100, "invalid bitcoin transaction")
	ErrBlockNotFound              = errorsmod.Register(ModuleName, 3101, "block not found")
	ErrTransactionNotIncluded     = errorsmod.Register(ModuleName, 3102, "transaction not included in block")
	ErrNotConfirmed               = errorsmod.Register(ModuleName, 3200, "transaction not confirmed")
	ErrExceedMaxAcceptanceDepth   = errorsmod.Register(ModuleName, 3201, "exceed max acceptance block depth")
	ErrUnsupportedScriptType      = errorsmod.Register(ModuleName, 3202, "unsupported script type")
	ErrTransactionAlreadyMinted   = errorsmod.Register(ModuleName, 3203, "transaction already minted")
	ErrInvalidDepositTransaction  = errorsmod.Register(ModuleName, 3204, "invalid deposit transaction")
	ErrInvalidWithdrawTransaction = errorsmod.Register(ModuleName, 3205, "invalid withdrawal transaction")

	ErrInvalidSignatures        = errorsmod.Register(ModuleName, 4200, "invalid signatures")
	ErrInsufficientBalance      = errorsmod.Register(ModuleName, 4201, "insufficient balance")
	ErrWithdrawRequestNotExist  = errorsmod.Register(ModuleName, 4202, "withdrawal request does not exist")
	ErrWithdrawRequestConfirmed = errorsmod.Register(ModuleName, 4203, "withdrawal request has been confirmed")
	ErrInvalidStatus            = errorsmod.Register(ModuleName, 4204, "invalid status")

	ErrUTXODoesNotExist = errorsmod.Register(ModuleName, 5100, "utxo does not exist")
	ErrUTXOLocked       = errorsmod.Register(ModuleName, 5101, "utxo locked")
	ErrUTXOUnlocked     = errorsmod.Register(ModuleName, 5102, "utxo unlocked")

	ErrInvalidAmount       = errorsmod.Register(ModuleName, 6100, "invalid amount")
	ErrInvalidFeeRate      = errorsmod.Register(ModuleName, 6101, "invalid fee rate")
	ErrAssetNotSupported   = errorsmod.Register(ModuleName, 6102, "asset not supported")
	ErrDustOutput          = errorsmod.Register(ModuleName, 6103, "too small output amount")
	ErrInsufficientUTXOs   = errorsmod.Register(ModuleName, 6104, "insufficient utxos")
	ErrFailToSerializePsbt = errorsmod.Register(ModuleName, 6105, "failed to serialize psbt")
	ErrInvalidRunes        = errorsmod.Register(ModuleName, 6106, "invalid runes")
	ErrInvalidRuneId       = errorsmod.Register(ModuleName, 6107, "invalid rune id")

	ErrInvalidParams = errorsmod.Register(ModuleName, 7100, "invalid module params")

	ErrInvalidDepositAmount  = errorsmod.Register(ModuleName, 8100, "invalid deposit amount")
	ErrInvalidWithdrawAmount = errorsmod.Register(ModuleName, 8101, "invalid withdrawal amount")
	ErrInvalidSequence       = errorsmod.Register(ModuleName, 8102, "invalid sequence")

	ErrInvalidDKGParams                 = errorsmod.Register(ModuleName, 9100, "invalid dkg params")
	ErrDKGRequestDoesNotExist           = errorsmod.Register(ModuleName, 9101, "dkg request does not exist")
	ErrDKGCompletionRequestExists       = errorsmod.Register(ModuleName, 9102, "dkg completion request already exists")
	ErrInvalidDKGCompletionRequest      = errorsmod.Register(ModuleName, 9103, "invalid dkg completion request")
	ErrUnauthorizedDKGCompletionRequest = errorsmod.Register(ModuleName, 9104, "unauthorized dkg completion request")
)
