package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidPubKey     = errorsmod.Register(ModuleName, 1000, "invalid pub key")
	ErrInvalidPubKeys    = errorsmod.Register(ModuleName, 1001, "invalid pub keys")
	ErrInvalidSignature  = errorsmod.Register(ModuleName, 1002, "invalid signature")
	ErrInvalidSignatures = errorsmod.Register(ModuleName, 1003, "invalid signatures")

	ErrDKGRequestDoesNotExist     = errorsmod.Register(ModuleName, 2000, "dkg request does not exist")
	ErrInvalidDKGStatus           = errorsmod.Register(ModuleName, 2001, "invalid dkg status")
	ErrDKGRequestExpired          = errorsmod.Register(ModuleName, 2002, "dkg request expired")
	ErrUnauthorizedParticipant    = errorsmod.Register(ModuleName, 2003, "unauthorized participant")
	ErrDKGCompletionAlreadyExists = errorsmod.Register(ModuleName, 2004, "dkg completion already exists")
	ErrInvalidDKGCompletion       = errorsmod.Register(ModuleName, 2005, "invalid dkg completion")
	ErrInvalidThreshold           = errorsmod.Register(ModuleName, 2006, "invalid threshold")

	ErrSigningRequestDoesNotExist = errorsmod.Register(ModuleName, 3000, "signing request does not exist")
	ErrInvalidSigningStatus       = errorsmod.Register(ModuleName, 3001, "invalid signing status")

	ErrInvalidDKGs                       = errorsmod.Register(ModuleName, 4000, "invalid dkgs")
	ErrInvalidParticipants               = errorsmod.Register(ModuleName, 4001, "invalid participants")
	ErrInvalidThresholds                 = errorsmod.Register(ModuleName, 4002, "invalid thresholds")
	ErrInvalidTimeoutDuration            = errorsmod.Register(ModuleName, 4003, "invalid timeout duration")
	ErrRefreshingRequestDoesNotExist     = errorsmod.Register(ModuleName, 4004, "refreshing request does not exist")
	ErrInvalidRefreshingStatus           = errorsmod.Register(ModuleName, 4005, "invalid refreshing status")
	ErrRefreshingRequestExpired          = errorsmod.Register(ModuleName, 4006, "refreshing request expired")
	ErrRefreshingCompletionAlreadyExists = errorsmod.Register(ModuleName, 4007, "refreshing completion already exists")

	ErrInvalidParams = errorsmod.Register(ModuleName, 5000, "invalid params")
)
