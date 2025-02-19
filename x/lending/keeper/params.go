package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsAuthorizedPoolCreator returns true if the given creator is authorized, false otherwise
func (k Keeper) IsAuthorizedPoolCreator(ctx sdk.Context, creator string) bool {
	authorizedPoolCreators := k.GetParams(ctx).PoolCreators

	if len(authorizedPoolCreators) == 0 {
		return true
	}

	return slices.Contains(authorizedPoolCreators, creator)
}
