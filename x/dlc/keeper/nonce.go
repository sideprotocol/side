package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// HandleNonces performs the nonces handling
func (k Keeper) HandleNonces(ctx sdk.Context, oraclePubKey string, nonces []string, intent int32) error {
	for _, nonce := range nonces {
		if err := k.HandleNonce(ctx, oraclePubKey, nonce, intent); err != nil {
			return err
		}
	}

	return nil
}

// HandleNonce performs the nonce handling
func (k Keeper) HandleNonce(ctx sdk.Context, oraclePubKey string, nonce string, intent int32) error {
	nonceBytes, _ := hex.DecodeString(nonce)
	if k.HasNonce(ctx, nonceBytes) {
		return errorsmod.Wrap(types.ErrInvalidNonce, "nonce already exists")
	}

	oraclePKBytes, _ := hex.DecodeString(oraclePubKey)
	if !k.HasOracleByPubKey(ctx, oraclePKBytes) {
		return types.ErrOracleDoesNotExist
	}

	oracle := k.GetOracleByPubKey(ctx, oraclePKBytes)

	dlcNonce := &types.DLCNonce{
		Index:        k.IncrementNonceIndex(ctx, oracle.Id),
		Nonce:        nonce,
		OraclePubkey: oraclePubKey,
		Time:         ctx.BlockTime(),
	}

	eventType := types.GetEventTypeFromIntent(intent)

	dlcEvent := &types.DLCEvent{
		Id:           k.IncrementEventId(ctx),
		Type:         eventType,
		Nonce:        nonce,
		Pubkey:       oraclePubKey,
		HasTriggered: false,
		OutcomeIndex: types.DefaultOutcomeIndex,
		PublishAt:    ctx.BlockTime(),
	}

	switch eventType {
	case types.DlcEventType_PRICE, types.DlcEventType_DATE:
		// not implemented currently
		return nil

	case types.DlcEventType_LENDING:
		// description and outcomes will be updated when bound to a loan

		k.AddLendingEventToPendingQueue(ctx, dlcEvent)
	}

	k.SetNonce(ctx, dlcNonce, oracle.Id)
	k.SetNonceByValue(ctx, nonceBytes)
	k.SetEvent(ctx, dlcEvent)

	return nil
}

// GetNonceIndex gets the current nonce index of the given oracle
func (k Keeper) GetNonceIndex(ctx sdk.Context, oracleId uint64) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.NonceIndexKey(oracleId))
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementNonceIndex increments the nonce index and returns the new index
func (k Keeper) IncrementNonceIndex(ctx sdk.Context, oracleId uint64) uint64 {
	store := ctx.KVStore(k.storeKey)

	index := k.GetNonceIndex(ctx, oracleId) + 1
	store.Set(types.NonceIndexKey(oracleId), sdk.Uint64ToBigEndian(index))

	return index
}

// HasNonce returns true if the given nonce exists, false otherwise
func (k Keeper) HasNonce(ctx sdk.Context, nonce []byte) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.NonceByValueKey(nonce))
}

// GetNonce gets the nonce by the given oracle id and index
func (k Keeper) GetNonce(ctx sdk.Context, oracleId uint64, index uint64) *types.DLCNonce {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.NonceKey(oracleId, index))
	var nonce types.DLCNonce
	k.cdc.MustUnmarshal(bz, &nonce)

	return &nonce
}

// SetNonce sets the given nonce
func (k Keeper) SetNonce(ctx sdk.Context, nonce *types.DLCNonce, oracleId uint64) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(nonce)
	store.Set(types.NonceKey(oracleId, nonce.Index), bz)
}

// SetNonceByValue sets the given nonce value
func (k Keeper) SetNonceByValue(ctx sdk.Context, nonce []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.NonceByValueKey(nonce), []byte{})
}

// GetNonces gets nonces of the given oracle
func (k Keeper) GetNonces(ctx sdk.Context, oracleId uint64) []*types.DLCNonce {
	nonces := make([]*types.DLCNonce, 0)

	k.IterateNonces(ctx, oracleId, func(nonce *types.DLCNonce) (stop bool) {
		nonces = append(nonces, nonce)
		return false
	})

	return nonces
}

// IterateNonces iterates through all nonces of the given oracle
func (k Keeper) IterateNonces(ctx sdk.Context, oracleId uint64, cb func(nonce *types.DLCNonce) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.NonceKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var nonce types.DLCNonce
		k.cdc.MustUnmarshal(iterator.Value(), &nonce)

		if cb(&nonce) {
			break
		}
	}
}
