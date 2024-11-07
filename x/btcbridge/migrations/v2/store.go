package v2

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	rollbackBlockHeader(ctx, storeKey, cdc)

	migrateSigningRequests(ctx, storeKey, cdc)
	migrateRunes(ctx, storeKey, cdc)

	return nil
}

// rollbackBlockHeader rolls back a block header due to a wrongly forked block at the height 3077817 of testnet3
// The new best block header will be the one at the height 3077817 minus 1
func rollbackBlockHeader(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	forkedBlockHeight := uint64(3077817)
	newBestBlockHeight := forkedBlockHeight - 1

	forkedBlockHash := store.Get(types.BtcBlockHeaderHeightKey(forkedBlockHeight))
	if forkedBlockHash == nil {
		panic(types.ErrInvalidBlockHeader)
	}

	forkedBlockHeaderBz := store.Get(types.BtcBlockHeaderHashKey(string(forkedBlockHash)))
	var forkedBlockHeader types.BlockHeader
	cdc.MustUnmarshal(forkedBlockHeaderBz, &forkedBlockHeader)

	newBestBlockHash := store.Get(types.BtcBlockHeaderHeightKey(newBestBlockHeight))
	if newBestBlockHash == nil {
		panic(types.ErrInvalidBlockHeader)
	}

	newBestBlockHeaderBz := store.Get(types.BtcBlockHeaderHashKey(string(newBestBlockHash)))
	var newBestBlockHeader types.BlockHeader
	cdc.MustUnmarshal(newBestBlockHeaderBz, &newBestBlockHeader)

	// remove the forked block header
	store.Delete(types.BtcBlockHeaderHashKey(forkedBlockHeader.Hash))
	store.Delete(types.BtcBlockHeaderHeightKey(forkedBlockHeader.Height))

	// set the new best block header
	bz := cdc.MustMarshal(&newBestBlockHeader)
	store.Set(types.BtcBestBlockHeaderKey, bz)
}

// migrateSigningRequests migrates the signing requests to add the status store and new `CreationTime` field
// Note: the migration will NOT add the status store for pending signing requests to avoid bo be fetched from Shutter(TSSigner)
func migrateSigningRequests(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BtcSigningRequestPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signingRequestV1 types.SigningRequestV1
		cdc.MustUnmarshal(iterator.Value(), &signingRequestV1)

		// add the new `CreationTime` field
		signingRequest := &types.SigningRequest{
			Address:      signingRequestV1.Address,
			Sequence:     signingRequestV1.Sequence,
			Txid:         signingRequestV1.Txid,
			Psbt:         signingRequestV1.Psbt,
			CreationTime: time.Time{},
			Status:       signingRequestV1.Status,
		}

		bz := cdc.MustMarshal(signingRequest)
		store.Set(types.BtcSigningRequestKey(signingRequest.Sequence), bz)

		if signingRequest.Status != types.SigningStatus_SIGNING_STATUS_PENDING {
			// add the status store
			store.Set(types.BtcSigningRequestByStatusKey(signingRequest.Status, signingRequest.Sequence), []byte{})
		}
	}
}

// migrateRunes migrates the runes to make the runes utxos sorted by amount
func migrateRunes(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BtcOwnerRunesUtxoKeyPrefix)
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
