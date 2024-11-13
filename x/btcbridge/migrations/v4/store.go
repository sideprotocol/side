package v4

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// MigrateStore migrates the x/btcbridge module state from the consensus version 3 to
// version 4
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateUtxos(ctx, storeKey, cdc)

	return nil
}

// migrateUtxos migrates the utxos to delete the locked ones
func migrateUtxos(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BtcUtxoKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var utxo types.UTXO
		cdc.MustUnmarshal(iterator.Value(), &utxo)

		if utxo.IsLocked {
			store.Delete(types.BtcOwnerUtxoKey(utxo.Address, utxo.Txid, utxo.Vout))
			store.Delete(types.BtcOwnerUtxoByAmountKey(utxo.Address, utxo.Amount, utxo.Txid, utxo.Vout))

			for _, r := range utxo.Runes {
				store.Delete(types.BtcOwnerRunesUtxoKey(utxo.Address, r.Id, r.Amount, utxo.Txid, utxo.Vout))
			}

			bz := store.Get(types.BtcSigningRequestByTxHashKey(utxo.Txid))
			if bz != nil {
				signingRequestBz := store.Get(types.BtcSigningRequestKey(sdk.BigEndianToUint64(bz)))

				var signingRequest types.SigningRequest
				cdc.MustUnmarshal(signingRequestBz, &signingRequest)

				if signingRequest.Status != types.SigningStatus_SIGNING_STATUS_CONFIRMED &&
					signingRequest.Status != types.SigningStatus_SIGNING_STATUS_FAILED {
					continue
				}
			}

			store.Delete(types.BtcUtxoKey(utxo.Txid, utxo.Vout))
		}
	}
}
