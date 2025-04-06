package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DKGTimeoutPeriod gets the DKG timeout period
func (k Keeper) DKGTimeoutPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DkgTimeoutPeriod
}
