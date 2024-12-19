package v5

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 4 to
// version 5
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, bankKeeper types.BankKeeper) error {
	migrateFeeRate(ctx, storeKey, cdc)
	migrateDKGRequests(ctx, storeKey, cdc)
	setDenomMetadata(ctx, bankKeeper)

	return nil
}

// migrateFeeRate migrates the fee rate to add the `height` field
func migrateFeeRate(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	feeRateBz := store.Get(types.BtcFeeRateKey)
	if feeRateBz != nil {
		// add the height field
		feeRate := &types.FeeRate{
			Value:  int64(sdk.BigEndianToUint64(feeRateBz)),
			Height: ctx.BlockHeight(),
		}

		store.Set(types.BtcFeeRateKey, cdc.MustMarshal(feeRate))
	}
}

// migrateDKGRequests migrates the dkg requests to delete the deprecated fields
func migrateDKGRequests(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DKGRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var request types.DKGRequestV1
		cdc.MustUnmarshal(iterator.Value(), &request)

		requestV2 := &types.DKGRequest{
			Id:             request.Id,
			Participants:   request.Participants,
			Threshold:      request.Threshold,
			VaultTypes:     request.VaultTypes,
			EnableTransfer: request.EnableTransfer,
			TargetUtxoNum:  request.TargetUtxoNum,
			Expiration:     request.Expiration,
			Status:         request.Status,
		}

		store.Set(iterator.Key(), cdc.MustMarshal(requestV2))
	}
}

// setDenomMetadata sets the denom metadata for sat
func setDenomMetadata(ctx sdk.Context, bankKeeper types.BankKeeper) {
	metadata := banktypes.Metadata{
		Description: "BTC pegged token via the Side btc bridge",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "sat",
				Exponent: 0,
				Aliases:  []string{},
			},
			{
				Denom:    "sBTC",
				Exponent: 8,
				Aliases:  []string{},
			},
		},
		Base:    "sat",
		Display: "sBTC",
		Name:    "BTC pegged token on Side",
		Symbol:  "SBTC",
		URI:     "",
		URIHash: "",
	}

	bankKeeper.SetDenomMetaData(ctx, metadata)
}
