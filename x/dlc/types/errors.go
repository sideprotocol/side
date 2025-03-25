package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPubKey      = errorsmod.Register(ModuleName, 1101, "invalid public key")
	ErrInvalidSignature   = errorsmod.Register(ModuleName, 1102, "invalid signature")
	ErrInvalidNonce       = errorsmod.Register(ModuleName, 1103, "invalid nonce")
	ErrOracleDoesNotExist = errorsmod.Register(ModuleName, 1104, "oracle does not exist")
	ErrDCMDoesNotExist = errorsmod.Register(ModuleName, 1105, "dcm does not exist")
	ErrInvalidEventType   = errorsmod.Register(ModuleName, 1106, "invalid event type")

	ErrEventDoesNotExist         = errorsmod.Register(ModuleName, 1107, "event does not exist")
	ErrEventNotTriggered         = errorsmod.Register(ModuleName, 1108, "event not triggered")
	ErrAttestationAlreadyExists  = errorsmod.Register(ModuleName, 1109, "attestation already exists")
	ErrInvalidParticipants       = errorsmod.Register(ModuleName, 1110, "invalid participants")
	ErrUnauthorizedParticipant   = errorsmod.Register(ModuleName, 1111, "unauthorized participant")
	ErrInvalidThreshold          = errorsmod.Register(ModuleName, 1112, "invalid threshold")
	ErrPendingOraclePubKeyExists = errorsmod.Register(ModuleName, 1113, "pending oracle public key already exists")
	ErrInvalidOracleStatus       = errorsmod.Register(ModuleName, 1114, "invalid oracle status")
	ErrPendingDCMPubKeyExists = errorsmod.Register(ModuleName, 1115, "pending dcm public key already exists")
	ErrInvalidDCMStatus       = errorsmod.Register(ModuleName, 1116, "invalid dcm status")
	ErrDKGTimedOut               = errorsmod.Register(ModuleName, 1117, "dkg timed out")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")
)
