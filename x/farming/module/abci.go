package farming

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/farming/keeper"
	"github.com/sideprotocol/side/x/farming/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	distributeRewards(ctx, k)
}

// distributeRewards performs rewards distribution
func distributeRewards(ctx sdk.Context, k keeper.Keeper) {
	// get all stakings
	stakings := k.GetAllStakings(ctx)

	for _, staking := range stakings {
		if staking.Status == types.StakingStatus_STAKING_STATUS_UNSTAKED {
			continue
		}

		if !ctx.BlockTime().Before(staking.StartTime.Add(staking.LockDuration)) {
			continue
		}

		// get the pending reward of the last distribution interval
		pendingReward := k.GetPendingReward(ctx, staking.Id)

		// accumulate reward
		staking.PendingReward = staking.PendingReward.Add(pendingReward)
		k.SetStaking(ctx, staking)
	}
}
