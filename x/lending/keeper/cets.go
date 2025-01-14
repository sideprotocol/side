package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// SetCETs sets the given CETs
func (k Keeper) SetCETs(ctx sdk.Context, loanId string, cets *types.CETs) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(cets)
	store.Set(types.LoanCETsKey(loanId), bz)
}

// GetCETs gets the specified CETs
func (k Keeper) GetCETs(ctx sdk.Context, loanId string) *types.CETs {
	store := ctx.KVStore(k.storeKey)

	var cets types.CETs
	bz := store.Get(types.LoanCETsKey(loanId))
	k.cdc.MustUnmarshal(bz, &cets)

	return &cets
}

// GetLiquidationBorrowerAdaptorSignature gets the liquidation borrower adaptor signature of the given loan
func (k Keeper) GetLiquidationBorrowerAdaptorSignature(ctx sdk.Context, loanId string) string {
	cets := k.GetCETs(ctx, loanId)

	return cets.LiquidationBorrowerAdaptorSignature
}
