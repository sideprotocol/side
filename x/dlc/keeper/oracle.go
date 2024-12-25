package keeper

import (
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
		Status:       types.DLCOracleStatus_Oracle_Status_Pending,
	}

	k.SetOracle(ctx, &oracle)

	return nil
}

// SubmitOraclePubKey performs the oracle public key submission
func (k Keeper) SubmitOraclePubKey(ctx sdk.Context, sender string, pubKey string, signature string) error {
	// TODO

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
