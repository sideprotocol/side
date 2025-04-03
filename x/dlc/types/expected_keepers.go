package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// OracleKeeper defines the expected oracle keeper interface
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error)
}

// TSSKeeper defines the expected TSS keeper interface
type TSSKeeper interface {
	InitiateDKG(ctx sdk.Context, module string, ty string, intent int32, participants []string, threshold uint32, batchSize uint32) *tsstypes.DKGRequest
	InitiateSigningRequest(ctx sdk.Context, module string, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, sigHashes []string, options *tsstypes.SigningOptions) *tsstypes.SigningRequest

	RegisterDKGRequestCompletedHandler(module string, handler tsstypes.DKGRequestCompletedHandler)
	RegisterSigningRequestCompletedHandler(module string, handler tsstypes.SigningRequestCompletedHandler)
}
