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

	ErrInvalidDKGParams           = errorsmod.Register(ModuleName, 2000, "invalid dkg params")
	ErrDKGRequestDoesNotExist     = errorsmod.Register(ModuleName, 2001, "dkg request does not exist")
	ErrInvalidDKGStatus           = errorsmod.Register(ModuleName, 2002, "invalid dkg status")
	ErrDKGRequestExpired          = errorsmod.Register(ModuleName, 2003, "dkg request expired")
	ErrUnauthorizedParticipant    = errorsmod.Register(ModuleName, 2004, "unauthorized participant")
	ErrDKGCompletionAlreadyExists = errorsmod.Register(ModuleName, 2005, "dkg completion already exists")
	ErrInvalidDKGCompletion       = errorsmod.Register(ModuleName, 2006, "invalid dkg completion")

	ErrSigningRequestDoesNotExist = errorsmod.Register(ModuleName, 3000, "signing request does not exist")
	ErrInvalidSigningStatus       = errorsmod.Register(ModuleName, 3001, "invalid signing status")

	ErrInvalidDKGs                      = errorsmod.Register(ModuleName, 4000, "invalid dkgs")
	ErrInvalidParticipants              = errorsmod.Register(ModuleName, 4001, "invalid participants")
	ErrInvalidTimeoutDuration           = errorsmod.Register(ModuleName, 4002, "invalid timeout duration")
	ErrResharingRequestDoesNotExist     = errorsmod.Register(ModuleName, 4003, "resharing request does not exist")
	ErrInvalidResharingStatus           = errorsmod.Register(ModuleName, 4004, "invalid resharing status")
	ErrResharingRequestExpired          = errorsmod.Register(ModuleName, 4005, "resharing request expired")
	ErrResharingCompletionAlreadyExists = errorsmod.Register(ModuleName, 4006, "resharing completion already exists")

	ErrInvalidParams = errorsmod.Register(ModuleName, 5000, "invalid params")
)
