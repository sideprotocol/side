package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// SetFeeRate sets the bitcoin network fee rate
func (k Keeper) SetFeeRate(ctx sdk.Context, feeRate int64) {
	store := ctx.KVStore(k.storeKey)

	feeRateWithHeight := types.FeeRate{
		Value:  feeRate,
		Height: ctx.BlockHeight(),
	}

	store.Set(types.BtcFeeRateKey, k.cdc.MustMarshal(&feeRateWithHeight))
}

// GetFeeRate gets the bitcoin network fee rate
func (k Keeper) GetFeeRate(ctx sdk.Context) *types.FeeRate {
	store := ctx.KVStore(k.storeKey)

	var feeRate types.FeeRate
	bz := store.Get(types.BtcFeeRateKey)
	k.cdc.MustUnmarshal(bz, &feeRate)

	return &feeRate
}

// CheckFeeRate checks the given fee rate
func (k Keeper) CheckFeeRate(ctx sdk.Context, feeRate *types.FeeRate) error {
	if feeRate.Value == 0 || ctx.BlockHeight()-feeRate.Height > k.GetParams(ctx).FeeRateValidityPeriod {
		return types.ErrInvalidFeeRate
	}

	return nil
}
