package keeper

import (
	"slices"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FinalTimeoutDuration gets the final timeout duration
func (k Keeper) FinalTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).FinalTimeoutDuration
}

// IsAuthorizedPoolCreator returns true if the given creator is authorized, false otherwise
func (k Keeper) IsAuthorizedPoolCreator(ctx sdk.Context, creator string) bool {
	authorizedPoolCreators := k.GetParams(ctx).PoolCreators

	if len(authorizedPoolCreators) == 0 {
		return true
	}

	return slices.Contains(authorizedPoolCreators, creator)
}
