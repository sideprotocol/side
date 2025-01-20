package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sideprotocol/side/x/incentive/types"
)

// IncentiveEnabled returns true if the incentive mechanism is enabled, false otherwise
func (k Keeper) IncentiveEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).Enabled && !k.bankKeeper.SpendableCoins(ctx, authtypes.NewModuleAddress(types.ModuleName)).IsZero()
}

// DepositIncentiveEnabled returns true if the incentive is enabled for deposit, false otherwise
func (k Keeper) DepositIncentiveEnabled(ctx sdk.Context) bool {
	return k.IncentiveEnabled(ctx) && k.RewardPerDeposit(ctx).IsPositive() && k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), k.DepositRewardDenom(ctx)).IsPositive()
}

// WithdrawIncentiveEnabled returns true if the incentive is enabled for withdrawal, false otherwise
func (k Keeper) WithdrawIncentiveEnabled(ctx sdk.Context) bool {
	return k.IncentiveEnabled(ctx) && k.RewardPerWithdraw(ctx).IsPositive() && k.bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(types.ModuleName), k.WithdrawRewardDenom(ctx)).IsPositive()
}

// RewardPerDeposit returns the reward amount for each deposit
func (k Keeper) RewardPerDeposit(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardPerDeposit
}

// RewardPerWithdraw returns the reward amount for each withdrawal
func (k Keeper) RewardPerWithdraw(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardPerWithdraw
}

// DepositRewardDenom returns the denom for deposit reward
func (k Keeper) DepositRewardDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).RewardPerDeposit.Denom
}

// WithdrawRewardDenom returns the denom for withdrawal reward
func (k Keeper) WithdrawRewardDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).RewardPerWithdraw.Denom
}
