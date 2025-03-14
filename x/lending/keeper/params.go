package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FinalTimeoutDuration gets the final timeout duration
func (k Keeper) FinalTimeoutDuration(ctx sdk.Context) int64 {
	return int64(k.GetParams(ctx).FinalTimeoutDuration)
}
