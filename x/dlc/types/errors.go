package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPubKey       = errorsmod.Register(ModuleName, 1101, "invalid public key")
	ErrInvalidSignature    = errorsmod.Register(ModuleName, 1102, "invalid signature")
	ErrInvalidNonce        = errorsmod.Register(ModuleName, 1103, "invalid nonce")
	ErrOracleDoesNotExist  = errorsmod.Register(ModuleName, 1104, "oracle does not exist")
	ErrEventDoesNotExist   = errorsmod.Register(ModuleName, 1105, "event does not exist")
	ErrInvalidParticipants = errorsmod.Register(ModuleName, 1106, "invalid participants")
	ErrInvalidThreshold    = errorsmod.Register(ModuleName, 1107, "invalid threshold")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")
)
