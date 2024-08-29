package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DepositEnabled returns true if deposit enabled, false otherwise
func (k Keeper) DepositEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).DepositEnabled
}

// WithdrawEnabled returns true if withdrawal enabled, false otherwise
func (k Keeper) WithdrawEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).WithdrawEnabled
}

// ProtocolDepositFeeEnabled returns true if the protocol fee is required for deposit, false otherwise
func (k Keeper) ProtocolDepositFeeEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).ProtocolFees.DepositFee > 0
}

// ProtocolWithdrawFeeEnabled returns true if the protocol fee is required for withdrawal, false otherwise
func (k Keeper) ProtocolWithdrawFeeEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).ProtocolFees.WithdrawFee > 0
}
