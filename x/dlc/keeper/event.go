package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

// GetEventId gets the current event id
func (k Keeper) GetEventId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EventIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementEventId increments the event id
func (k Keeper) IncrementEventId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetEventId(ctx) + 1
	store.Set(types.EventIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// GetCurrentEventPrice gets the current event price
func (k Keeper) GetCurrentEventPrice(ctx sdk.Context, pair string) int64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.CurrentEventPriceKey(pair))
	if bz == nil {
		return 0
	}

	return int64(sdk.BigEndianToUint64(bz))
}

// SetCurrentEventPrice sets the current event price for the given pair
func (k Keeper) SetCurrentEventPrice(ctx sdk.Context, pair string, price sdkmath.Int) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.CurrentEventPriceKey(pair), sdk.Uint64ToBigEndian(price.Uint64()))
}

// HasEvent returns true if the given event exists, false otherwise
func (k Keeper) HasEvent(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.EventKey(id))
}

// GetEvent gets the event by the given id
func (k Keeper) GetEvent(ctx sdk.Context, id uint64) *types.DLCPriceEvent {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EventKey(id))
	var event types.DLCPriceEvent
	k.cdc.MustUnmarshal(bz, &event)

	return &event
}

// GetEventByPrice gets the event by the given price
func (k Keeper) GetEventByPrice(ctx sdk.Context, price sdkmath.Int) *types.DLCPriceEvent {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EventByPriceKey(price))
	if bz == nil {
		return nil
	}

	return k.GetEvent(ctx, sdk.BigEndianToUint64(bz))
}

// SetEvent sets the given event
func (k Keeper) SetEvent(ctx sdk.Context, event *types.DLCPriceEvent) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(event)
	store.Set(types.EventKey(event.Id), bz)

	store.Set(types.EventByPriceKey(event.TriggerPrice), sdk.Uint64ToBigEndian(event.Id))
}

// TriggerEvent sets the given event to triggered
func (k Keeper) TriggerEvent(ctx sdk.Context, event *types.DLCPriceEvent) {
	store := ctx.KVStore(k.storeKey)

	event.HasTriggered = true

	bz := k.cdc.MustMarshal(event)
	store.Set(types.EventKey(event.Id), bz)
}

// GetAllEvents gets all events
func (k Keeper) GetAllEvents(ctx sdk.Context) []*types.DLCPriceEvent {
	events := make([]*types.DLCPriceEvent, 0)

	k.IterateEvents(ctx, func(event *types.DLCPriceEvent) (stop bool) {
		events = append(events, event)
		return false
	})

	return events
}

// GetEvents gets events according to the specified status
func (k Keeper) GetEvents(ctx sdk.Context, triggered bool) []*types.DLCPriceEvent {
	events := make([]*types.DLCPriceEvent, 0)

	k.IterateEventsByStatus(ctx, triggered, func(event *types.DLCPriceEvent) (stop bool) {
		events = append(events, event)
		return false
	})

	return events
}

// IterateEventsByStatus iterates through events by the given status
func (k Keeper) IterateEventsByStatus(ctx sdk.Context, triggered bool, cb func(event *types.DLCPriceEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.DLCPriceEvent
		k.cdc.MustUnmarshal(iterator.Value(), &event)

		if event.HasTriggered == triggered && cb(&event) {
			break
		}
	}
}

// IterateEvents iterates through all events
func (k Keeper) IterateEvents(ctx sdk.Context, cb func(event *types.DLCPriceEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.DLCPriceEvent
		k.cdc.MustUnmarshal(iterator.Value(), &event)

		if cb(&event) {
			break
		}
	}
}
