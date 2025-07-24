package v2

import (
	"encoding/hex"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// MigrateStore migrates the x/dlc module state from the consensus version 1 to
// version 2
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, tssKeeper types.TSSKeeper, cdc codec.BinaryCodec) error {
	migrateDLCEvents(ctx, storeKey, cdc)
	migrateDCMsAndOracles(ctx, storeKey, tssKeeper, cdc)

	return nil
}

// migrateDLCEvents performs the dlc events migration
func migrateDLCEvents(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var dlcEvent types.DLCEvent
		cdc.MustUnmarshal(iterator.Value(), &dlcEvent)

		// set event by status
		store.Set(types.EventByStatusKey(dlcEvent.HasTriggered, dlcEvent.Id), []byte{})
	}
}

// migrateDCMsAndOracles performs the DCMs and oracles migration
func migrateDCMsAndOracles(ctx sdk.Context, storeKey storetypes.StoreKey, tssKeeper types.TSSKeeper, cdc codec.BinaryCodec) {
	tssKeeper.IterateDKGRequests(ctx, func(req *tsstypes.DKGRequest) (stop bool) {
		if req.Status == tsstypes.DKGStatus_DKG_STATUS_COMPLETED {
			// dcm or oracle pub key
			pubKey := tssKeeper.GetDKGPubKeys(ctx, req.Id)[0]
			pubKeyBz, _ := hex.DecodeString(pubKey)

			if req.Type == types.DKG_TYPE_DCM {
				updateDCM(ctx, storeKey, req.Id, pubKeyBz, cdc)
			} else if req.Type == types.DKG_TYPE_NONCE {
				updateOracle(ctx, storeKey, req.Id, pubKeyBz, cdc)
			}
		}

		return false
	})
}

// updateDCM updates the given dcm
func updateDCM(ctx sdk.Context, storeKey storetypes.StoreKey, dkgId uint64, pubKey []byte, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	bz := store.Get(types.DCMByPubKeyKey(pubKey))
	id := sdk.BigEndianToUint64(bz)

	// unmarshal to DCM v1
	dcmBz := store.Get(types.DCMKey(id))
	var dcmV1 types.DCMV1
	cdc.MustUnmarshal(dcmBz, &dcmV1)

	// build new DCM
	dcm := &types.DCM{
		Id:     id,
		DkgId:  dkgId,
		Desc:   dcmV1.Desc,
		Pubkey: dcmV1.Pubkey,
		Time:   dcmV1.Time,
		Status: dcmV1.Status,
	}

	// update DCM
	store.Set(types.DCMKey(id), cdc.MustMarshal(dcm))
}

// updateOracle updates the given oracle
func updateOracle(ctx sdk.Context, storeKey storetypes.StoreKey, dkgId uint64, pubKey []byte, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	bz := store.Get(types.OracleByPubKeyKey(pubKey))
	id := sdk.BigEndianToUint64(bz)

	// unmarshal to oracle v1
	oracleBz := store.Get(types.OracleKey(id))
	var oracleV1 types.DLCOracleV1
	cdc.MustUnmarshal(oracleBz, &oracleV1)

	// build new oracle
	oracle := &types.DLCOracle{
		Id:         id,
		DkgId:      dkgId,
		Desc:       oracleV1.Desc,
		Pubkey:     oracleV1.Pubkey,
		NonceIndex: oracleV1.NonceIndex,
		Time:       oracleV1.Time,
		Status:     oracleV1.Status,
	}

	// update oracle
	store.Set(types.OracleKey(id), cdc.MustMarshal(oracle))
}
