package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidBid          = errorsmod.Register(ModuleName, 1100, "invalid bid")
	ErrBidDoesNotExist     = errorsmod.Register(ModuleName, 1101, "bid does not exist")
	ErrInvalidBidStatus    = errorsmod.Register(ModuleName, 1102, "invalid bid status")
	ErrUnauthorized        = errorsmod.Register(ModuleName, 1103, "unauthorized operation")
	ErrAuctionDoesNotExist = errorsmod.Register(ModuleName, 1104, "auction does not exist")
	ErrAuctionClosed       = errorsmod.Register(ModuleName, 1105, "auction already closed")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")
)
