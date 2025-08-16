package keeper

import (
	"slices"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/types"
)

// Enabled returns true if farming is enabled, false otherwise
func (k Keeper) FarmingEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).Enabled
}

// EpochDuration gets the epoch duration
func (k Keeper) EpochDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).EpochDuration
}

// RewardsPerEpoch gets the reward per epoch
func (k Keeper) RewardPerEpoch(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardPerEpoch
}

// LockDurations gets the lock durations
func (k Keeper) LockDurations(ctx sdk.Context) []time.Duration {
	return k.GetParams(ctx).LockDurations
}

// LockDurationExists returns true if the given lock duration exists, false otherwise
func (k Keeper) LockDurationExists(ctx sdk.Context, lockDuration time.Duration) bool {
	return slices.Contains(k.LockDurations(ctx), lockDuration)
}

// EligibleAssets gets all eligible assets
func (k Keeper) EligibleAssets(ctx sdk.Context) []types.Asset {
	return k.GetParams(ctx).EligibleAssets
}

// IsEligibleAsset returns true if the given asset is eligible, false otherwise
func (k Keeper) IsEligibleAsset(ctx sdk.Context, denom string) bool {
	for _, asset := range k.EligibleAssets(ctx) {
		if asset.Denom == denom {
			return true
		}
	}

	return false
}

// Asset gets the asset by the given denom
func (k Keeper) Asset(ctx sdk.Context, denom string) types.Asset {
	for _, asset := range k.EligibleAssets(ctx) {
		if asset.Denom == denom {
			return asset
		}
	}

	return types.Asset{}
}
