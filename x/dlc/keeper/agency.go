package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// CreateAgency initiates the agency creation request
func (k Keeper) CreateAgency(ctx sdk.Context, participants []string, threshold uint32) error {
	agency := types.Agency{
		Id:           k.IncrementAgencyId(ctx),
		Participants: participants,
		Threshold:    threshold,
		Status:       types.AgencyStatus_Agency_Status_Pending,
	}

	k.SetAgency(ctx, &agency)

	return nil
}

// SubmitAgencyPubKey performs the agency public key submission
func (k Keeper) SubmitAgencyPubKey(ctx sdk.Context, sender string, pubKey string, signature string) error {
	// TODO

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
