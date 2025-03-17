package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/auction/types"
)

// HandleBiddedAsset handles the bidded asset for the liquidated loan
func (k Keeper) HandleBiddedAsset(ctx sdk.Context, loanId string, moduleAccount string, asset sdk.Coin) error {
	loan := k.GetLoan(ctx, loanId)

	if asset.Amount.LTE(loan.Interest) {
		// TODO
		return nil
	}

	poolAmount := asset.SubAmount(loan.ProtocolFee)
	protocolFee := sdk.NewCoin(asset.Denom, loan.ProtocolFee)

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, moduleAccount, types.ModuleName, sdk.NewCoins(poolAmount)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(k.GetParams(ctx).ProtocolFeeCollector), sdk.NewCoins(protocolFee)); err != nil {
		return err
	}

	k.AfterPoolRepaid(ctx, loan.PoolId, asset.SubAmount(loan.Interest), loan.Interest.Sub(loan.ProtocolFee))

	return nil
}
