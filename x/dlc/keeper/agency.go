package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// CreateAgency initiates the agency creation request
func (k Keeper) CreateAgency(ctx sdk.Context, participants []string, threshold uint32) (*types.Agency, error) {
	agency := &types.Agency{
		Id:           k.IncrementAgencyId(ctx),
		Participants: participants,
		Threshold:    threshold,
		Time:         ctx.BlockTime(),
		Status:       types.AgencyStatus_Agency_Status_Pending,
	}

	k.SetAgency(ctx, agency)

	return agency, nil
}

// SubmitAgencyPubKey performs the agency public key submission
func (k Keeper) SubmitAgencyPubKey(ctx sdk.Context, sender string, pubKey string, agencyId uint64, agencyPubKey string, signature string) error {
	agency := k.GetAgency(ctx, agencyId)
	if agency == nil {
		return types.ErrAgencyDoesNotExist
	}

	if !types.ParticipantExists(agency.Participants, pubKey) {
		return types.ErrUnauthorizedParticipant
	}

	pubKeyBytes, _ := hex.DecodeString(pubKey)

	if k.HasPendingAgencyPubKey(ctx, agencyId, pubKeyBytes) {
		return types.ErrPendingAgencyPubKeyExists
	}

	if agency.Status != types.AgencyStatus_Agency_Status_Pending {
		return types.ErrInvalidAgencyStatus
	}

	if !ctx.BlockTime().Before(agency.Time.Add(k.GetDKGTimeoutPeriod(ctx))) {
		return errorsmod.Wrap(types.ErrDKGTimedOut, "agency dkg timed out")
	}

	agencyPubKeyBytes, _ := hex.DecodeString(agencyPubKey)
	sigBytes, _ := hex.DecodeString(signature)
	sigMsg := types.GetSigMsg(agencyId, agencyPubKeyBytes)

	if !types.VerifySignature(sigBytes, pubKeyBytes, sigMsg) {
		return errorsmod.Wrap(types.ErrInvalidSignature, "signature verification failed")
	}

	k.SetPendingAgencyPubKey(ctx, agencyId, pubKeyBytes, agencyPubKeyBytes)

	return nil
}

// GetAgencyId gets the current agency id
func (k Keeper) GetAgencyId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AgencyIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementAgencyId increments the agency id and returns the new id
func (k Keeper) IncrementAgencyId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetAgencyId(ctx) + 1
	store.Set(types.AgencyIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// GetAgency gets the agency by the given id
func (k Keeper) GetAgency(ctx sdk.Context, id uint64) *types.Agency {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AgencyKey(id))
	var agency types.Agency
	k.cdc.MustUnmarshal(bz, &agency)

	return &agency
}

// SetAgency sets the given agency
func (k Keeper) SetAgency(ctx sdk.Context, agency *types.Agency) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(agency)
	store.Set(types.AgencyKey(agency.Id), bz)
}

// HasPendingAgencyPubKey returns true if the given pending agency pubkey exists, false otherwise
func (k Keeper) HasPendingAgencyPubKey(ctx sdk.Context, agencyId uint64, pubKey []byte) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.PendingAgencyPubKeyKey(agencyId, pubKey))
}

// SetPendingAgencyPubKey sets the pending agency public key
func (k Keeper) SetPendingAgencyPubKey(ctx sdk.Context, agencyId uint64, pubKey []byte, agencyPubKey []byte) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PendingAgencyPubKeyKey(agencyId, pubKey), agencyPubKey)
}

// GetAgencies gets agencies by the given status
func (k Keeper) GetAgencies(ctx sdk.Context, status types.AgencyStatus) []*types.Agency {
	agencies := make([]*types.Agency, 0)

	k.IterateAgencies(ctx, func(agency *types.Agency) (stop bool) {
		if agency.Status == status {
			agencies = append(agencies, agency)
		}

		return false
	})

	return agencies
}

// GetPendingAgencyPubKeys gets pending agency pub keys by the given agency id
func (k Keeper) GetPendingAgencyPubKeys(ctx sdk.Context, agencyId uint64) [][]byte {
	pubKeys := make([][]byte, 0)

	k.IteratePendingAgencyPubKeys(ctx, agencyId, func(pubKey []byte) (stop bool) {
		pubKeys = append(pubKeys, pubKey)

		return false
	})

	return pubKeys
}

// IterateAgencies iterates through all agencies
func (k Keeper) IterateAgencies(ctx sdk.Context, cb func(agency *types.Agency) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.AgencyKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var agency types.Agency
		k.cdc.MustUnmarshal(iterator.Value(), &agency)

		if cb(&agency) {
			break
		}
	}
}

// IteratePendingAgencyPubKeys iterates through all pending agency pub keys by the given agency id
func (k Keeper) IteratePendingAgencyPubKeys(ctx sdk.Context, agencyId uint64, cb func(pubKey []byte) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.PendingAgencyPubKeyKeyPrefix, sdk.Uint64ToBigEndian(agencyId)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if cb(iterator.Value()) {
			break
		}
	}
}
