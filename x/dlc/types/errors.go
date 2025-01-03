package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPubKey             = errorsmod.Register(ModuleName, 1101, "invalid public key")
	ErrInvalidSignature          = errorsmod.Register(ModuleName, 1102, "invalid signature")
	ErrInvalidNonce              = errorsmod.Register(ModuleName, 1103, "invalid nonce")
	ErrOracleDoesNotExist        = errorsmod.Register(ModuleName, 1104, "oracle does not exist")
	ErrAgencyDoesNotExist        = errorsmod.Register(ModuleName, 1105, "agency does not exist")
	ErrEventDoesNotExist         = errorsmod.Register(ModuleName, 1106, "event does not exist")
	ErrInvalidParticipants       = errorsmod.Register(ModuleName, 1107, "invalid participants")
	ErrUnauthorizedParticipant   = errorsmod.Register(ModuleName, 1108, "unauthorized participant")
	ErrInvalidThreshold          = errorsmod.Register(ModuleName, 1109, "invalid threshold")
	ErrPendingOraclePubKeyExists = errorsmod.Register(ModuleName, 1110, "pending oracle public key already exists")
	ErrInvalidOracleStatus       = errorsmod.Register(ModuleName, 1111, "invalid oracle status")
	ErrPendingAgencyPubKeyExists = errorsmod.Register(ModuleName, 1112, "pending agency public key already exists")
	ErrInvalidAgencyStatus       = errorsmod.Register(ModuleName, 1113, "invalid agency status")
	ErrDKGTimedOut               = errorsmod.Register(ModuleName, 1114, "dkg timed out")

	ErrInvalidParams = errorsmod.Register(ModuleName, 2100, "invalid params")
)
