package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
	oracletypes "github.com/sideprotocol/side/x/oracle/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, oracleStoreKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateParams(ctx, storeKey, cdc)
	migrateBlockHeaders(ctx, storeKey, oracleStoreKey, cdc)

	return nil
}

// migrateParams performs the params migration
func migrateParams(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	// get current params
	var paramsV1 types.ParamsV1
	bz := store.Get(types.ParamsStoreKey)
	cdc.MustUnmarshal(bz, &paramsV1)

	// build new params
	params := &types.Params{
		DepositConfirmationDepth:  types.DefaultDepositConfirmationDepth,
		WithdrawConfirmationDepth: types.DefaultWithdrawConfirmationDepth,
		MaxAcceptableBlockDepth:   paramsV1.MaxAcceptableBlockDepth,
		BtcVoucherDenom:           paramsV1.BtcVoucherDenom,
		DepositEnabled:            paramsV1.DepositEnabled,
		WithdrawEnabled:           paramsV1.WithdrawEnabled,
		TrustedNonBtcRelayers:     paramsV1.TrustedNonBtcRelayers,
		TrustedFeeProviders:       paramsV1.TrustedFeeProviders,
		FeeRateValidityPeriod:     paramsV1.FeeRateValidityPeriod,
		Vaults:                    paramsV1.Vaults,
		WithdrawParams:            paramsV1.WithdrawParams,
		ProtocolLimits:            paramsV1.ProtocolLimits,
		ProtocolFees:              paramsV1.ProtocolFees,
		TssParams:                 paramsV1.TssParams,
		IbcParams: types.IBCParams{
			TimeoutHeightOffset: types.DefaultIBCTimeoutHeightOffset,
			TimeoutDuration:     types.DefaultIBCTimeoutDuration,
		},
	}

	bz = cdc.MustMarshal(params)
	store.Set(types.ParamsStoreKey, bz)
}

// migrateBlockHeaders performs the block headers migration from btcbridge to oracle
func migrateBlockHeaders(ctx sdk.Context, storeKey storetypes.StoreKey, oracleStoreKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)
	oracleStore := ctx.KVStore(oracleStoreKey)

	iterator := storetypes.KVStorePrefixIterator(store, BtcBlockHeaderPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var blockHeader oracletypes.BlockHeader
		cdc.MustUnmarshal(iterator.Value(), &blockHeader)

		// migrate to oracle
		oracleStore.Set(oracletypes.BitcoinHeaderKey(blockHeader.Hash), iterator.Value())
		oracleStore.Set(oracletypes.BitcoinBlockHeaderHeightKey(blockHeader.Height), []byte(blockHeader.Hash))

		// remove block header
		store.Delete(iterator.Key())
		store.Delete(BtcBlockHeaderHeightKey(uint64(blockHeader.Height)))
	}

	bestBlockHeader := store.Get(BtcBestBlockHeaderKey)
	if bestBlockHeader != nil {
		// set best block header to oracle
		oracleStore.Set(oracletypes.BitcoinBestBlockHeaderKey, bestBlockHeader)

		// remove best block header
		store.Delete(BtcBestBlockHeaderKey)
	}
}
