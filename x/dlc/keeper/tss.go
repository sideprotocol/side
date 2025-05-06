package keeper

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// DKGCompletedHandler is callback handler when the DKG request completed by TSS
func (k Keeper) DKGCompletedHandler(ctx sdk.Context, id uint64, ty string, intent int32, pubKeys []string) error {
	switch ty {
	case types.DKG_TYPE_DCM:
		return k.CreateDCM(ctx, pubKeys[0])

	case types.DKG_TYPE_NONCE:
		// the first pub key is oracle and the remaining are nonces

		if err := k.CreateOracle(ctx, pubKeys[0]); err != nil {
			return err
		}

		return k.HandleNonces(ctx, pubKeys[0], pubKeys[1:], intent)
	}

	return nil
}

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, id uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	return k.HandleAttestation(ctx, sender, types.FromScopedId(scopedId), signatures[0])
}

// GetOracleParticipants gets oracle participants
func (k Keeper) GetOracleParticipants(ctx sdk.Context) []string {
	baseParticipants := k.OracleParticipantBaseSet(ctx)
	participantNum := int(k.OracleParticipantNum(ctx))

	if participantNum == len(baseParticipants) {
		return baseParticipants
	}

	participants := []string{}

	// select oracle participants randomly by the expected participant number
	rand := rand.New(rand.NewSource(ctx.BlockTime().Unix()))
	selectedIndices := rand.Perm(len(baseParticipants))[0:participantNum]
	for _, index := range selectedIndices {
		participants = append(participants, baseParticipants[index])
	}

	return participants
}
