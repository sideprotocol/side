package keeper

import (
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
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

// NonceGenerationTimeoutDuration gets the nonce generation timeout duration
func (k Keeper) NonceGenerationTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).NonceGenerationTimeoutDuration
}

// AllowedOracleParticipants gets the allowed oracle participants
func (k Keeper) AllowedOracleParticipants(ctx sdk.Context) []string {
	return k.GetParams(ctx).AllowedOracleParticipants
}

// OracleParticipantNum gets the oracle participant number
func (k Keeper) OracleParticipantNum(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantNum
}

// OracleParticipantThreshold gets the oracle participant threshold
func (k Keeper) OracleParticipantThreshold(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).OracleParticipantThreshold
}

// ValidateOracleParticipantAllowlist validates the allowed oracle participants
func (k Keeper) ValidateOracleParticipantAllowlist(ctx sdk.Context, allowedOracleParticipants []string) error {
	baseParticipants := k.tssKeeper.AllowedDKGParticipants(ctx)

	if len(allowedOracleParticipants) != 0 && len(baseParticipants) != 0 {
		for _, p := range allowedOracleParticipants {
			if !slices.Contains(baseParticipants, p) {
				return errorsmod.Wrap(types.ErrInvalidParams, "oracle participant not authorized")
			}
		}
	}

	return nil
}
