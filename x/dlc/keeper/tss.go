package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// DKGCompletedHandler is callback handler when the DKG request completed by TSS
func (k Keeper) DKGCompletedHandler(ctx sdk.Context, dkgRequestId uint64, ty string, intent int32, pubKeys []string) error {
	switch types.DKGIntent(intent) {
	case types.DKGIntent_DKG_INTENT_DCM:
		k.CreateDCM(ctx, pubKeys[0])

	default:
		return nil
	}

	return nil
}

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, signingRequestId uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	return k.HandleAttestation(ctx, sender, types.FromScopedId(scopedId), signatures[0])
}
