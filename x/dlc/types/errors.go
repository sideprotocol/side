package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidSignature         = errorsmod.Register(ModuleName, 1100, "invalid signature")
	ErrInvalidNonce             = errorsmod.Register(ModuleName, 1101, "invalid nonce")
	ErrDCMDoesNotExist          = errorsmod.Register(ModuleName, 1102, "dcm does not exist")
	ErrDCMAlreadyExists         = errorsmod.Register(ModuleName, 1103, "dcm already exists")
	ErrOracleDoesNotExist       = errorsmod.Register(ModuleName, 1104, "oracle does not exist")
	ErrOracleAlreadyExists      = errorsmod.Register(ModuleName, 1105, "oracle already exists")
	ErrEventDoesNotExist        = errorsmod.Register(ModuleName, 1106, "event does not exist")
	ErrEventNotTriggered        = errorsmod.Register(ModuleName, 1107, "event not triggered")
	ErrAttestationAlreadyExists = errorsmod.Register(ModuleName, 1108, "attestation already exists")
	ErrInvalidParticipants      = errorsmod.Register(ModuleName, 1109, "invalid participants")
	ErrInvalidThreshold         = errorsmod.Register(ModuleName, 1110, "invalid threshold")
	ErrInvalidDKGIntent         = errorsmod.Register(ModuleName, 1111, "invalid dkg intent")

	ErrInsufficientOracleParticipants = errorsmod.Register(ModuleName, 2100, "insufficient oracle participants")

	ErrInvalidParams = errorsmod.Register(ModuleName, 3100, "invalid params")
)
