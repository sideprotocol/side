package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AllowedDKGParticipants gets the allowed DKG participants
func (k Keeper) AllowedDKGParticipants(ctx sdk.Context) []string {
	participants := []string{}

	for _, p := range k.GetParams(ctx).AllowedDkgParticipants {
		participants = append(participants, p.ConsensusPubkey)
	}

	return participants
}

// DKGTimeoutDuration gets the DKG timeout duration
func (k Keeper) DKGTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutDuration
}
