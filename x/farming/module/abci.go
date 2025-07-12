package farming

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/keeper"
	"github.com/sideprotocol/side/x/farming/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleMatureStakings(ctx, k)

	updateEpoch(ctx, k)
}

// updateEpoch updates the epoch
func updateEpoch(ctx sdk.Context, k keeper.Keeper) {
	if k.FarmingEnabled(ctx) {
		currentEpoch := k.GetCurrentEpoch(ctx)
		if !ctx.BlockTime().Before(currentEpoch.EndTime) {
			// call handler on epoch ended
			k.OnEpochEnded(ctx)

			// end the current epoch
			currentEpoch.Status = types.EpochStatus_EPOCH_STATUS_ENDED
			k.SetEpoch(ctx, currentEpoch)

			// start the new epoch
			k.NewEpoch(ctx)

			// call handler on epoch started
			k.OnEpochStarted(ctx)
		}
	}
}

// handleMatureStakings performs handling for the mature stakings
func handleMatureStakings(ctx sdk.Context, k keeper.Keeper) {
	// get all stakings
	stakings := k.GetAllStakings(ctx)

	for _, staking := range stakings {
		if staking.Status != types.StakingStatus_STAKING_STATUS_STAKED {
			continue
		}

		if ctx.BlockTime().Before(staking.StartTime.Add(staking.LockDuration)) {
			continue
		}

		// update status
		staking.Status = types.StakingStatus_STAKING_STATUS_UNLOCKED
		k.SetStaking(ctx, staking)
	}
}
