package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPubKeys          = errorsmod.Register(ModuleName, 1000, "invalid pub keys")
	ErrInvalidPubKey           = errorsmod.Register(ModuleName, 1001, "invalid pub key")
	ErrInvalidSignatures       = errorsmod.Register(ModuleName, 1002, "invalid signatures")
	ErrInvalidSignature        = errorsmod.Register(ModuleName, 1003, "invalid signature")
	ErrUnauthorizedParticipant = errorsmod.Register(ModuleName, 1004, "unauthorized participant")
	ErrDKGTimedOut             = errorsmod.Register(ModuleName, 1005, "dkg timed out")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2000, "invalid params")
)
