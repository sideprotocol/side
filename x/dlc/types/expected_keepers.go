package types

import (
	context "context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// OracleKeeper defines the expected oracle keeper interface
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error)
}

// StakingKeeper defines the expected staking keeper used to retrieve validator (noalias)
type StakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)

	IterateBondedValidatorsByPower(context.Context, func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error
}

// TSSKeeper defines the expected TSS keeper interface
type TSSKeeper interface {
	AllowedDKGParticipants(ctx sdk.Context) []string

	InitiateDKG(ctx sdk.Context, module string, ty string, intent int32, participants []string, threshold uint32, batchSize uint32) *tsstypes.DKGRequest
	InitiateSigningRequest(ctx sdk.Context, module string, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, sigHashes []string, options *tsstypes.SigningOptions) *tsstypes.SigningRequest

	RegisterDKGCompletionReceivedHandler(module string, handler tsstypes.DKGCompletionReceivedHandler)
	RegisterDKGRequestCompletedHandler(module string, handler tsstypes.DKGRequestCompletedHandler)
	RegisterDKGRequestTimeoutHandler(module string, handler tsstypes.DKGRequestTimeoutHandler)
	RegisterSigningRequestCompletedHandler(module string, handler tsstypes.SigningRequestCompletedHandler)
}
