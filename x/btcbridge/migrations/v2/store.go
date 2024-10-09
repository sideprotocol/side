package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	rollbackBlockHeader(ctx, storeKey, cdc)
	markPendingSigningRequestsFailed(ctx, storeKey, cdc)
	migrateRunes(ctx, storeKey, cdc)

	return nil
}

// rollbackBlockHeader rolls back a block header due to a wrongly forked block at the height 3077817 of testnet3
// The new best block header will be the one at the height 3077817 minus 1
func rollbackBlockHeader(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	forkedBlockHeight := uint64(3077817)
	newBestBlockHeight := forkedBlockHeight - 1

	forkedBlockHeaderBz := store.Get(types.BtcBlockHeaderHeightKey(forkedBlockHeight))
	if forkedBlockHeaderBz == nil {
		panic(types.ErrInvalidBlockHeader)
	}

	var forkedBlockHeader types.BlockHeader
	cdc.MustUnmarshal(forkedBlockHeaderBz, &forkedBlockHeader)

	newBestBlockHeaderBz := store.Get(types.BtcBlockHeaderHeightKey(newBestBlockHeight))
	if newBestBlockHeaderBz == nil {
		panic(types.ErrInvalidBlockHeader)
	}

	var newBestBlockHeader types.BlockHeader
	cdc.MustUnmarshal(newBestBlockHeaderBz, &newBestBlockHeader)

	// remove the forked block header
	store.Delete(types.BtcBlockHeaderHashKey(forkedBlockHeader.Hash))
	store.Delete(types.BtcBlockHeaderHeightKey(forkedBlockHeader.Height))

	// set the new best block header
	bz := cdc.MustMarshal(&newBestBlockHeader)
	store.Set(types.BtcBestBlockHeaderKey, bz)
}

// markPendingSigningRequestsFailed marks the previous pending signing requests failed to avoid bo be constantly fetched from TSSigner
func markPendingSigningRequestsFailed(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcSigningRequestPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signingRequest types.SigningRequest
		cdc.MustUnmarshal(iterator.Value(), &signingRequest)

		if signingRequest.Status == types.SigningStatus_SIGNING_STATUS_PENDING {
			signingRequest.Status = types.SigningStatus_SIGNING_STATUS_FAILED

			sequence := sdk.BigEndianToUint64(iterator.Key())
			bz := cdc.MustMarshal(&signingRequest)

			store.Set(types.BtcSigningRequestKey(sequence), bz)
		}
	}
}

// migrateRunes migrates the runes to make the runes utxos sorted by amount
func migrateRunes(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcOwnerRunesUtxoKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		value := iterator.Value()

		hash := string(key[len(key)-64-8 : len(key)-8])
		vout := sdk.BigEndianToUint64(key[len(key)-8:])

		bz := store.Get(types.BtcUtxoKey(hash, vout))
		var utxo types.UTXO
		cdc.MustUnmarshal(bz, &utxo)

		id := string(key[1+len(utxo.Address) : len(key)-64-8])
		amount := string(value)

		// delete the original key
		store.Delete(key)

		// set the new key
		store.Set(types.BtcOwnerRunesUtxoKey(utxo.Address, id, amount, hash, vout), []byte{})
	}
}
