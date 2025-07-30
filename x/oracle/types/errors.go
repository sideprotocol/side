package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrInvalidBitcoinRPC     = errorsmod.Register(ModuleName, 1001, "invalid bitcoin rpc endpoint")
	ErrInvalidBitcoinRPCUser = errorsmod.Register(ModuleName, 1002, "invalid bitcoin rpc user")
	ErrInvalidBitcoinRPCPass = errorsmod.Register(ModuleName, 1003, "invalid bitcoin rpc password")

	ErrInvalidBlockHeader  = errorsmod.Register(ModuleName, 1100, "invalid block header")
	ErrInvalidBlockHeaders = errorsmod.Register(ModuleName, 1101, "invalid block headers")

	ErrInsufficientVotingPower = errorsmod.Register(ModuleName, 1200, "insufficient voting power")
)
