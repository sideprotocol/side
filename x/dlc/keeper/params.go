package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EventNonceQueueSize gets the nonce queue size
func (k Keeper) NonceQueueSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).NonceQueueSize
}

// NonceGenerationBatchSize gets the nonce generation batch size
func (k Keeper) NonceGenerationBatchSize(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).NonceGenerationBatchSize
}

// NonceGenerationInterval gets the nonce generation interval
func (k Keeper) NonceGenerationInterval(ctx sdk.Context) int64 {
	return k.GetParams(ctx).NonceGenerationInterval
}

// OracleParticipantBaseSet gets the oracle participant base set
func (k Keeper) OracleParticipantBaseSet(ctx sdk.Context) []string {
	if len(k.GetParams(ctx).AllowedOracleParticipants) != 0 {
		return k.GetParams(ctx).AllowedOracleParticipants
	}

	return k.tssKeeper.AllowedDKGParticipants(ctx)
}

// OracleParticipantNum gets the oracle participant number
func (k Keeper) OracleParticipantNum(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantNum
}

// OracleParticipantThreshold gets the oracle participant threshold
func (k Keeper) OracleParticipantThreshold(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantThreshold
}

// OracleParticipantLivenessCheckInterval gets the oracle participant liveness check interval
func (k Keeper) OracleParticipantLivenessCheckInterval(ctx sdk.Context) int64 {
	return k.GetParams(ctx).OracleParticipantLivenessCheckInterval
}
