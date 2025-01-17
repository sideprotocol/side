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

// RewardPerDeposit returns the reward amount for each deposit
func (k Keeper) RewardPerDeposit(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardPerDeposit
}

// RewardPerWithdraw returns the reward amount for each withdrawal
func (k Keeper) RewardPerWithdraw(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).RewardPerWithdraw
}
