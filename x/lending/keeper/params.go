package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FinalTimeoutDuration gets the final timeout duration
func (k Keeper) FinalTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).FinalTimeoutDuration
}
