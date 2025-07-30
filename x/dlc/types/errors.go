package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidNonce             = errorsmod.Register(ModuleName, 1100, "invalid nonce")
	ErrDCMDoesNotExist          = errorsmod.Register(ModuleName, 1101, "dcm does not exist")
	ErrDCMAlreadyExists         = errorsmod.Register(ModuleName, 1102, "dcm already exists")
	ErrOracleDoesNotExist       = errorsmod.Register(ModuleName, 1103, "oracle does not exist")
	ErrOracleAlreadyExists      = errorsmod.Register(ModuleName, 1104, "oracle already exists")
	ErrEventDoesNotExist        = errorsmod.Register(ModuleName, 1105, "event does not exist")
	ErrEventNotTriggered        = errorsmod.Register(ModuleName, 1106, "event not triggered")
	ErrAttestationAlreadyExists = errorsmod.Register(ModuleName, 1107, "attestation already exists")
	ErrInvalidParticipants      = errorsmod.Register(ModuleName, 1108, "invalid participants")
	ErrInvalidThreshold         = errorsmod.Register(ModuleName, 1109, "invalid threshold")
	ErrInvalidTimeoutDuration   = errorsmod.Register(ModuleName, 1110, "invalid timeout duration")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")
)
