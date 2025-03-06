package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcbridge module sentinel errors
var (
	ErrInvalidBlockHeader  = errorsmod.Register(ModuleName, 1100, "invalid block header")
	ErrInvalidBlockHeaders = errorsmod.Register(ModuleName, 1101, "invalid block headers")
	ErrInvalidReorgDepth   = errorsmod.Register(ModuleName, 1102, "invalid reorg depth")
)
