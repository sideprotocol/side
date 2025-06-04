package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrInvalidBitcoinRPC     = errorsmod.Register(ModuleName, 101, "invalid bitcoin rpc endpoint")
	ErrInvalidBitcoinRPCUser = errorsmod.Register(ModuleName, 102, "invalid bitcoin rpc user")
	ErrInvalidBitcoinRPCPass = errorsmod.Register(ModuleName, 103, "invalid bitcoin rpc password")

	ErrInvalidBlockHeader  = errorsmod.Register(ModuleName, 1100, "invalid block header")
	ErrInvalidBlockHeaders = errorsmod.Register(ModuleName, 1101, "invalid block headers")
	ErrInvalidReorgDepth   = errorsmod.Register(ModuleName, 1102, "invalid reorg depth")

	ErrInsufficientVotingPower = errorsmod.Register(ModuleName, 1200, "insufficient voting power")
)
