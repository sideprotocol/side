package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// SetFeeRate sets the bitcoin network fee rate
func (k Keeper) SetFeeRate(ctx sdk.Context, feeRate int64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.BtcFeeRateKey, sdk.Uint64ToBigEndian(uint64(feeRate)))
}

// GetFeeRate gets the bitcoin network fee rate
func (k Keeper) GetFeeRate(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcFeeRateKey)

	return int64(sdk.BigEndianToUint64(bz))
}
