package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// DKGCompletedHandler is callback handler when the DKG request completed by TSS
func (k Keeper) DKGCompletedHandler(ctx sdk.Context, dkgRequestId uint64, ty string, intent int32, pubKeys []string) error {
	switch ty {
	case types.DKG_TYPE_DCM:
		k.CreateDCM(ctx, pubKeys[0])

	case types.DKG_TYPE_NONCE:
		// the first pub key is oracle and the remaining are nonces
		k.CreateOracle(ctx, pubKeys[0])
		k.HandleNonces(ctx, pubKeys[0], pubKeys[1:], intent)
	}

	return nil
}

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, signingRequestId uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	return k.HandleAttestation(ctx, sender, types.FromScopedId(scopedId), signatures[0])
}
