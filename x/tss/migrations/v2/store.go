package v2

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/types"
)

// MigrateStore migrates the x/tss module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateDKGRequests(ctx, storeKey, cdc)
	migrateSigningRequests(ctx, storeKey, cdc)

	return nil
}

// migrateDKGRequests performs the DKG requests migration
func migrateDKGRequests(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DKGRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dkgRequest types.DKGRequest
		cdc.MustUnmarshal(iterator.Value(), &dkgRequest)

		// set the DKG request by status
		store.Set(types.DKGRequestByStatusKey(dkgRequest.Status, dkgRequest.Id), []byte{})
	}
}

// migrateSigningRequests performs the signing reqeusts migration
func migrateSigningRequests(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.SigningRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signingRequest types.SigningRequest
		cdc.MustUnmarshal(iterator.Value(), &signingRequest)

		// set the signing request by status
		store.Set(types.SigningRequestByStatusKey(signingRequest.Status, signingRequest.Id), []byte{})
	}
}
