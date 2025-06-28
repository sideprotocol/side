package keeper

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// DKGCompletionReceivedHandler is callback handler when the DKG completion received by TSS
func (k Keeper) DKGCompletionReceivedHandler(ctx sdk.Context, id uint64, ty string, intent int32, participant string) error {
	switch ty {
	case types.DKG_TYPE_NONCE:
		k.SetOracleParticipantLiveness(ctx, &types.OracleParticipantLiveness{
			ConsensusPubkey: participant,
			IsAlive:         true,
			LastDkgId:       id,
			LastBlockHeight: ctx.BlockHeight(),
		})
	}

	return nil
}

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

// DKGTimeoutHandler is callback handler when the DKG timed out in TSS
func (k Keeper) DKGTimeoutHandler(ctx sdk.Context, id uint64, ty string, intent int32, absentParticipants []string) error {
	switch ty {
	case types.DKG_TYPE_NONCE:
		for _, participant := range absentParticipants {
			k.SetOracleParticipantLiveness(ctx, &types.OracleParticipantLiveness{
				ConsensusPubkey: participant,
				IsAlive:         false,
			})
		}
	}

	return nil
}

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, id uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	return k.HandleAttestation(ctx, sender, types.FromScopedId(scopedId), signatures[0])
}

// GetOracleParticipants gets oracle participants
func (k Keeper) GetOracleParticipants(ctx sdk.Context) []string {
	// get alive participants
	aliveParticipants := k.GetAliveOracleParticipants(ctx)

	// check the participant num
	participantNum := int(k.OracleParticipantNum(ctx))
	if len(aliveParticipants) < participantNum {
		return nil
	}

	if len(aliveParticipants) == participantNum {
		return aliveParticipants
	}

	participants := []string{}

	// select oracle participants randomly from the alive list based on the expected participant number
	rand := rand.New(rand.NewSource(ctx.BlockTime().Unix()))
	selectedIndices := rand.Perm(len(aliveParticipants))[0:participantNum]
	for _, index := range selectedIndices {
		participants = append(participants, aliveParticipants[index])
	}

	return participants
}

// GetAliveOracleParticipants gets alive oracle participants
func (k Keeper) GetAliveOracleParticipants(ctx sdk.Context) []string {
	// get base participants
	baseParticipants := k.OracleParticipantBaseSet(ctx)

	// get alive participants
	aliveParticipants := []string{}
	for _, participant := range baseParticipants {
		if k.IsOracleParticipantAlive(ctx, participant) {
			aliveParticipants = append(aliveParticipants, participant)
		}
	}

	return aliveParticipants
}
