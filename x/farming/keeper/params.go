package keeper

import (
	"slices"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/farming/types"
)

// PhaseDuration gets the phase duration
func (k Keeper) PhaseStartTime(ctx sdk.Context) time.Time {
	return k.GetParams(ctx).PhaseStartTime
}

// PhaseDuration gets the phase duration
func (k Keeper) PhaseDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).PhaseDuration
}

// LockDurations gets the lock durations
func (k Keeper) LockDurations(ctx sdk.Context) []time.Duration {
	return k.GetParams(ctx).LockDurations
}

// LockDurationExists returns true if the given lock duration exists, false otherwise
func (k Keeper) LockDurationExists(ctx sdk.Context, lockDuration time.Duration) bool {
	return slices.Contains(k.LockDurations(ctx), lockDuration)
}

// DistributionInterval gets the distribution interval
func (k Keeper) DistributionInterval(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).DistributionInterval
}

// RewardsPerInterval gets the rewards per interval
func (k Keeper) RewardsPerInterval(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardsPerInterval
}

// AllowlistedAssets gets all allowlisted assets
func (k Keeper) AllowlistedAssets(ctx sdk.Context) []types.Asset {
	return k.GetParams(ctx).AllowlistedAssets
}

// AllowlistedAsset gets the allowlisted asset by the given denom
func (k Keeper) AllowlistedAsset(ctx sdk.Context, denom string) (types.Asset, bool) {
	for _, asset := range k.AllowlistedAssets(ctx) {
		if asset.Denom == denom {
			return asset, true
		}
	}

	return types.Asset{}, false
}

// IsAllowlistedAsset returns true if the given asset is allowlisted, false otherwise
func (k Keeper) IsAllowlistedAsset(ctx sdk.Context, denom string) bool {
	for _, asset := range k.AllowlistedAssets(ctx) {
		if asset.Denom == denom {
			return true
		}
	}

	return false
}
