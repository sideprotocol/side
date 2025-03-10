package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidBid           = errorsmod.Register(ModuleName, 1100, "invalid bid")
	ErrBidDoesNotExist      = errorsmod.Register(ModuleName, 1101, "bid does not exist")
	ErrInvalidBidStatus     = errorsmod.Register(ModuleName, 1102, "invalid bid status")
	ErrUnauthorized         = errorsmod.Register(ModuleName, 1103, "unauthorized operation")
	ErrAuctionDoesNotExist  = errorsmod.Register(ModuleName, 1104, "auction does not exist")
	ErrInvalidAuctionStatus = errorsmod.Register(ModuleName, 1105, "invalid auction status")
	ErrAuctionEnded         = errorsmod.Register(ModuleName, 1106, "auction already ended")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")

	ErrFailedToBuildTx   = errorsmod.Register(ModuleName, 3100, "failed to build transaction")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 3101, "invalid signatures")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 3102, "invalid signature")
)
