package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DKGTimeoutDuration gets the DKG timeout duration
func (k Keeper) DKGTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutDuration
}
