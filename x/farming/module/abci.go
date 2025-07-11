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

		if !ctx.BlockTime().Before(staking.EndTime) {
			continue
		}

		pendingReward := k.GetPendingReward(ctx, staking.Id)

		if err := k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(""), sdk.NewCoins(pendingReward)); err != nil {
			k.Logger(ctx).Error("Failed to distribute reward", "staker", "", "reward", pendingReward, "err", err)
		}
	}
}
