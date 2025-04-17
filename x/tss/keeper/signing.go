package keeper

import (
	"fmt"
	"strings"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/tss/types"
)

// GetSigningRequestId returns the signing request id
func (k Keeper) GetSigningRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.SigningRequestIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementSigningRequestId increments the signing request id and returns the new id
func (k Keeper) IncrementSigningRequestId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetSigningRequestId(ctx) + 1
	store.Set(types.SigningRequestIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// SetSigningRequest sets the signing request
func (k Keeper) SetSigningRequest(ctx sdk.Context, signingRequest *types.SigningRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(signingRequest)

	store.Set(types.SigningRequestKey(signingRequest.Id), bz)
}

// HasSigningRequest returns true if the given signing request exists, false otherwise
func (k Keeper) HasSigningRequest(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.SigningRequestKey(id))
}

// GetSigningRequest returns the signing request by the given id
func (k Keeper) GetSigningRequest(ctx sdk.Context, id uint64) *types.SigningRequest {
	store := ctx.KVStore(k.storeKey)

	var signingRequest types.SigningRequest
	bz := store.Get(types.SigningRequestKey(id))
	k.cdc.MustUnmarshal(bz, &signingRequest)

	return &signingRequest
}

// GetSigningRequests gets the signing requests by the given status
func (k Keeper) GetSigningRequests(ctx sdk.Context, status types.SigningStatus) []*types.SigningRequest {
	requests := make([]*types.SigningRequest, 0)

	k.IterateSigningRequests(ctx, func(req *types.SigningRequest) (stop bool) {
		if req.Status == status {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// IterateSigningRequests iterates through all signing requests
func (k Keeper) IterateSigningRequests(ctx sdk.Context, cb func(signingRequest *types.SigningRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.SigningRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signingRequest types.SigningRequest
		k.cdc.MustUnmarshal(iterator.Value(), &signingRequest)

		if cb(&signingRequest) {
			break
		}
	}
}

// InitiateSigningRequest initiates the signing request with the specified params
func (k Keeper) InitiateSigningRequest(ctx sdk.Context, module string, scopedId string, ty types.SigningType, intent int32, pubKey string, sigHashes []string, options *types.SigningOptions) *types.SigningRequest {
	req := &types.SigningRequest{
		Id:           k.IncrementSigningRequestId(ctx),
		Module:       module,
		ScopedId:     scopedId,
		Type:         ty,
		Intent:       intent,
		PubKey:       pubKey,
		SigHashes:    sigHashes,
		Options:      options,
		CreationTime: ctx.BlockTime(),
		Status:       types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	k.SetSigningRequest(ctx, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateSigning,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
			sdk.NewAttribute(types.AttributeKeyModule, module),
			sdk.NewAttribute(types.AttributeKeyScopedId, scopedId),
			sdk.NewAttribute(types.AttributeKeyType, fmt.Sprintf("%d", ty)),
			sdk.NewAttribute(types.AttributeKeyIntent, fmt.Sprintf("%d", intent)),
			sdk.NewAttribute(types.AttributeKeyPubKey, pubKey),
			sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(sigHashes, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyOption, types.GetSigningOption(ty, options)),
		),
	)

	return req
}
