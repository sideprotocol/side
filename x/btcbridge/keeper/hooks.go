package keeper

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// AfterDeposit performs the extended logic after deposit
func (k Keeper) AfterDeposit(ctx sdk.Context, addr string, amount sdk.Coin, tx *wire.MsgTx) error {
	// distribute deposit reward if enabled
	if k.incentiveKeeper.DepositIncentiveEnabled(ctx) {
		_ = k.incentiveKeeper.DistributeDepositReward(ctx, addr)
	}

	// perform IBC transfer if enabled
	script := types.GetIBCTransferScript(tx)
	if len(script) != 0 {
		// parse channel id and recipient
		channelId, recipient, err := types.ParseIBCTransferScript(script)
		if err != nil {
			// no error
			return nil
		}

		// transfer
		sequence, err := k.IBCTransfer(ctx, addr, recipient, amount, channelId)

		// emit event
		event := sdk.NewEvent(
			types.EventTypeIBCTransfer,
			sdk.NewAttribute(types.AttributeKeyChannelId, channelId),
		)
		if err == nil {
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyPacketSequence, fmt.Sprintf("%d", sequence)))
		} else {
			event = event.AppendAttributes(sdk.NewAttribute(types.AttributeKeyErrorMsg, err.Error()))
		}

		ctx.EventManager().EmitEvent(event)
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
