package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FinalTimeoutDuration gets the final timeout duration in seconds
func (k Keeper) FinalTimeoutDuration(ctx sdk.Context) int64 {
	return int64(k.GetParams(ctx).FinalTimeoutDuration / time.Second)
}

// RequestFeeCollector gets the request fee collector
func (k Keeper) RequestFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).RequestFeeCollector
}

// OriginationFeeCollector gets the origination fee collector
func (k Keeper) OriginationFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).OriginationFeeCollector
}

// ProtocolFeeCollector gets the protocol fee collector
func (k Keeper) ProtocolFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).ProtocolFeeCollector
}
