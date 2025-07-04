package keeper

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"slices"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// CreateOracle creates a new oracle with the given pub key
// Assume that the pub key is valid
func (k Keeper) CreateOracle(ctx sdk.Context, pubKey string) error {
	pubKeyBz, _ := hex.DecodeString(pubKey)
	if k.HasOracleByPubKey(ctx, pubKeyBz) {
		return types.ErrOracleAlreadyExists
	}

	oracle := &types.DLCOracle{
		Id:     k.IncrementOracleId(ctx),
		Pubkey: pubKey,
		Time:   ctx.BlockTime(),
		Status: types.DLCOracleStatus_Oracle_status_Enable,
	}

	k.SetOracle(ctx, oracle)
	k.SetOracleByPubKey(ctx, oracle.Id, pubKeyBz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateOracle,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", oracle.Id)),
			sdk.NewAttribute(types.AttributeKeyPubKey, oracle.Pubkey),
		),
	)

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

// SetOracleByPubKey sets the given oracle by pub key
func (k Keeper) SetOracleByPubKey(ctx sdk.Context, oracleId uint64, pubKey []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.OracleByPubKeyKey(pubKey), sdk.Uint64ToBigEndian(oracleId))
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

// GetOracleParticipantBaseSet gets the oracle participant base set
func (k Keeper) GetOracleParticipantBaseSet(ctx sdk.Context) []string {
	allowedOracleParticipants := k.AllowedOracleParticipants(ctx)

	baseParticipants := k.tssKeeper.AllowedDKGParticipants(ctx)
	if len(baseParticipants) == 0 {
		return allowedOracleParticipants
	}

	oracleParticipantBaseSet := []string{}
	for _, participant := range allowedOracleParticipants {
		if slices.Contains(baseParticipants, participant) {
			oracleParticipantBaseSet = append(oracleParticipantBaseSet, participant)
		}
	}

	return oracleParticipantBaseSet
}

// GetOracleParticipants gets oracle participants
// First select alive participants, then fallback to random participants if not sufficient
func (k Keeper) GetOracleParticipants(ctx sdk.Context) []string {
	aliveParticipants := k.GetAliveOracleParticipants(ctx)
	if len(aliveParticipants) != 0 {
		return aliveParticipants
	}

	return k.GetRandomOracleParticipants(ctx)
}

// GetRandomOracleParticipants gets oracle participants randomly
func (k Keeper) GetRandomOracleParticipants(ctx sdk.Context) []string {
	baseParticipants := k.GetOracleParticipantBaseSet(ctx)
	participantNum := int(k.OracleParticipantNum(ctx))

	return k.SelectOracleParticipants(ctx, baseParticipants, participantNum)
}

// GetAliveOracleParticipants gets alive oracle participants randomly
func (k Keeper) GetAliveOracleParticipants(ctx sdk.Context) []string {
	aliveParticipants := k.GetAllAliveOracleParticipants(ctx)
	participantNum := int(k.OracleParticipantNum(ctx))

	return k.SelectOracleParticipants(ctx, aliveParticipants, participantNum)
}

// GetAllAliveOracleParticipants gets all alive oracle participants
func (k Keeper) GetAllAliveOracleParticipants(ctx sdk.Context) []string {
	// get base participants
	baseParticipants := k.GetOracleParticipantBaseSet(ctx)

	// filter alive participants
	aliveParticipants := []string{}
	for _, participant := range baseParticipants {
		if k.IsOracleParticipantAlive(ctx, participant) {
			aliveParticipants = append(aliveParticipants, participant)
		}
	}

	return aliveParticipants
}

// SelectOracleParticipants selects oracle participants randomly from the base set based on the specified participant num
func (k Keeper) SelectOracleParticipants(ctx sdk.Context, baseOracleParticipants []string, participantNum int) []string {
	if len(baseOracleParticipants) < participantNum {
		return nil
	}

	if len(baseOracleParticipants) == participantNum {
		return baseOracleParticipants
	}

	selectedParticipants := []string{}

	rand := rand.New(rand.NewSource(ctx.BlockTime().Unix()))
	selectedIndices := rand.Perm(len(baseOracleParticipants))[0:participantNum]
	for _, index := range selectedIndices {
		selectedParticipants = append(selectedParticipants, baseOracleParticipants[index])
	}

	return selectedParticipants
}
