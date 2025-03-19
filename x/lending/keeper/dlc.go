package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
)

// SetDLCMeta sets the given dlc meta
func (k Keeper) SetDLCMeta(ctx sdk.Context, loanId string, dlcMeta *types.DLCMeta) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(dlcMeta)
	store.Set(types.LoanDLCMetaKey(loanId), bz)
}

// GetDLCMeta gets the specified dlc meta
func (k Keeper) GetDLCMeta(ctx sdk.Context, loanId string) *types.DLCMeta {
	store := ctx.KVStore(k.storeKey)

	var dlcMeta types.DLCMeta
	bz := store.Get(types.LoanDLCMetaKey(loanId))
	k.cdc.MustUnmarshal(bz, &dlcMeta)

	return &dlcMeta
}

// GetCetInfos gets the related cet infos of the given loan
func (k Keeper) GetCetInfos(ctx sdk.Context, loanId string, collateralAmount sdk.Coin) ([]*types.CetInfo, error) {
	loan := k.GetLoan(ctx, loanId)
	pool := k.GetPool(ctx, loan.PoolId)

	multisigScript, _ := types.CreateMultisigScript([]string{loan.BorrowerPubKey, loan.Agency})

	var liquidationEvent *dlctypes.DLCEvent
	if loan.LiquidationEventId != 0 {
		liquidationEvent = k.dlcKeeper.GetEvent(ctx, loan.LiquidationEventId)
	} else if collateralAmount.Amount.IsPositive() {
		liquidationPrice := types.GetLiquidationPrice(collateralAmount.Amount, loan.BorrowAmount.Amount, sdkmath.NewInt(int64(pool.Config.LiquidationThreshold)))
		liquidationEvent = k.dlcKeeper.GetEventByPrice(ctx, liquidationPrice)
	}

	defaultLiquidationEvent := k.dlcKeeper.GetEvent(ctx, loan.DefaultLiquidationEventId)
	repaymentEvent := k.dlcKeeper.GetEvent(ctx, loan.RepaymentEventId)

	liquidationCetInfo, _ := types.GetCetInfo(liquidationEvent, 0, multisigScript)
	defaultLiquidationCetInfo, _ := types.GetCetInfo(defaultLiquidationEvent, 0, multisigScript)
	repaymentCetInfo, _ := types.GetCetInfo(repaymentEvent, 0, multisigScript)

	return []*types.CetInfo{
		liquidationCetInfo,
		defaultLiquidationCetInfo,
		repaymentCetInfo,
	}, nil
}
