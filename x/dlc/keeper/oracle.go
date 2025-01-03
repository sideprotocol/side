package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// CreateOracle initiates the oracle creation request
func (k Keeper) CreateOracle(ctx sdk.Context, participants []string, threshold uint32) error {
	oracle := types.DLCOracle{
		Id:           k.IncrementOracleId(ctx),
		Participants: participants,
		Threshold:    threshold,
		Time:         ctx.BlockTime(),
		Status:       types.DLCOracleStatus_Oracle_Status_Pending,
	}

	k.SetOracle(ctx, &oracle)

	return nil
}

// SubmitOraclePubKey performs the oracle public key submission
func (k Keeper) SubmitOraclePubKey(ctx sdk.Context, sender string, pubKey string, oracleId uint64, oraclePubKey string, signature string) error {
	oracle := k.GetOracle(ctx, oracleId)
	if oracle == nil {
		return types.ErrOracleDoesNotExist
	}

	if !types.ParticipantExists(oracle.Participants, pubKey) {
		return types.ErrUnauthorizedParticipant
	}

	pubKeyBytes, _ := hex.DecodeString(pubKey)

	if k.HasPendingOraclePubKey(ctx, oracleId, pubKeyBytes) {
		return types.ErrPendingOraclePubKeyExists
	}

	if oracle.Status != types.DLCOracleStatus_Oracle_Status_Pending {
		return types.ErrInvalidOracleStatus
	}

	if !ctx.BlockTime().Before(oracle.Time.Add(k.GetDKGTimeoutPeriod(ctx))) {
		return errorsmod.Wrap(types.ErrDKGTimedOut, "oracle dkg timed out")
	}

	oraclePubKeyBytes, _ := hex.DecodeString(oraclePubKey)
	sigBytes, _ := hex.DecodeString(signature)
	sigMsg := types.GetSigMsg(oracleId, oraclePubKeyBytes)

	if !types.VerifySignature(sigBytes, pubKeyBytes, sigMsg) {
		return errorsmod.Wrap(types.ErrInvalidSignature, "signature verification failed")
	}

	k.SetPendingOraclePubKey(ctx, oracleId, pubKeyBytes, oraclePubKeyBytes)

	return nil
}

// GetOracleId gets the current oracle id
func (k Keeper) GetOracleId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OracleIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementOracleId increments the oracle id and returns the new id
func (k Keeper) IncrementOracleId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetOracleId(ctx) + 1
	store.Set(types.OracleIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasOracleByPubKey returns true if the given oracle exists, false otherwise
func (k Keeper) HasOracleByPubKey(ctx sdk.Context, pubKey []byte) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.OracleByPubKeyKey(pubKey))
}

// GetOracle gets the oracle by the given id
func (k Keeper) GetOracle(ctx sdk.Context, id uint64) *types.DLCOracle {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OracleKey(id))
	var oracle types.DLCOracle
	k.cdc.MustUnmarshal(bz, &oracle)

	return &oracle
}

// GetOracleByPubKey gets the oracle by the given public key
func (k Keeper) GetOracleByPubKey(ctx sdk.Context, pubKey []byte) *types.DLCOracle {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.OracleByPubKeyKey(pubKey))
	if bz == nil {
		return nil
	}

	return k.GetOracle(ctx, sdk.BigEndianToUint64(bz))
}

// SetOracle sets the given oracle
func (k Keeper) SetOracle(ctx sdk.Context, oracle *types.DLCOracle) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(oracle)
	store.Set(types.OracleKey(oracle.Id), bz)
}

// HasPendingOraclePubKey returns true if the given pending oracle pubkey exists, false otherwise
func (k Keeper) HasPendingOraclePubKey(ctx sdk.Context, oracleId uint64, pubKey []byte) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.PendingOraclePubKeyKey(oracleId, pubKey))
}

// SetPendingOraclePubKey sets the pending oracle public key
func (k Keeper) SetPendingOraclePubKey(ctx sdk.Context, oracleId uint64, pubKey []byte, oraclePubKey []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PendingOraclePubKeyKey(oracleId, pubKey), oraclePubKey)
}

// GetOracles gets oracles by the given status
func (k Keeper) GetOracles(ctx sdk.Context, status types.DLCOracleStatus) []*types.DLCOracle {
	oracles := make([]*types.DLCOracle, 0)

	k.IterateOracles(ctx, func(oracle *types.DLCOracle) (stop bool) {
		if oracle.Status == status {
			oracles = append(oracles, oracle)
		}

		return false
	})

	return oracles
}

// GetPendingOraclePubKeys gets pending oracle pub keys by the given oracle id
func (k Keeper) GetPendingOraclePubKeys(ctx sdk.Context, oracleId uint64) [][]byte {
	pubKeys := make([][]byte, 0)

	k.IteratePendingOraclePubKeys(ctx, oracleId, func(pubKey []byte) (stop bool) {
		pubKeys = append(pubKeys, pubKey)

		return false
	})

	return pubKeys
}

// IterateOracles iterates through all oracles
func (k Keeper) IterateOracles(ctx sdk.Context, cb func(oracle *types.DLCOracle) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.OracleKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var oracle types.DLCOracle
		k.cdc.MustUnmarshal(iterator.Value(), &oracle)

		if cb(&oracle) {
			break
		}
	}
}

// IteratePendingOraclePubKeys iterates through all pending oracle pub keys by the given oracle id
func (k Keeper) IteratePendingOraclePubKeys(ctx sdk.Context, oracleId uint64, cb func(pubKey []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.PendingOraclePubKeyKeyPrefix, sdk.Uint64ToBigEndian(oracleId)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if cb(iterator.Value()) {
			break
		}
	}
}
