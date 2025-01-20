package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AfterDeposit performs the extended logic after deposit
func (k Keeper) AfterDeposit(ctx sdk.Context, addr string) error {
	// distribute deposit reward
	if k.incentiveKeeper.DepositIncentiveEnabled(ctx) {
		_ = k.incentiveKeeper.DistributeDepositReward(ctx, addr)
	}

	return nil
}

// AfterWithdraw performs the extended logic after withdrawal
func (k Keeper) AfterWithdraw(ctx sdk.Context, txHash string) error {
	// distribute rewards for all withdrawals
	if k.incentiveKeeper.WithdrawIncentiveEnabled(ctx) {
		withdrawRequests := k.GetWithdrawRequestsByTxHash(ctx, txHash)
		for _, req := range withdrawRequests {
			_ = k.incentiveKeeper.DistributeWithdrawReward(ctx, req.Address)
		}
	}

	return nil
}
