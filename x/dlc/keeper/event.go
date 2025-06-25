package keeper

import (
	"encoding/base64"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
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

// HasEvent returns true if the given event exists, false otherwise
func (k Keeper) HasEvent(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.EventKey(id))
}

// GetEvent gets the event by the given id
func (k Keeper) GetEvent(ctx sdk.Context, id uint64) *types.DLCEvent {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EventKey(id))
	var event types.DLCEvent
	k.cdc.MustUnmarshal(bz, &event)

	return &event
}

// SetEvent sets the given event
func (k Keeper) SetEvent(ctx sdk.Context, event *types.DLCEvent) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(event)

	store.Set(types.EventKey(event.Id), bz)
}

// GetPendingLendingEventCount gets the pending lending event count
func (k Keeper) GetPendingLendingEventCount(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.PendingLendingEventCountKey)

	return sdk.BigEndianToUint64(bz)
}

// IncreasePendingLendingEventCount increases the pending lending event count by 1
func (k Keeper) IncreasePendingLendingEventCount(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	count := k.GetPendingLendingEventCount(ctx)

	store.Set(types.PendingLendingEventCountKey, sdk.Uint64ToBigEndian(count+1))
}

// DecreasePendingLendingEventCount decreases the pending lending event count by 1
func (k Keeper) DecreasePendingLendingEventCount(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	count := k.GetPendingLendingEventCount(ctx)
	if count == 0 {
		return
	}

	store.Set(types.PendingLendingEventCountKey, sdk.Uint64ToBigEndian(count-1))
}

// AddLendingEventToPendingQueue adds the specified lending event to the pending queue
func (k Keeper) AddLendingEventToPendingQueue(ctx sdk.Context, event *types.DLCEvent) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.PendingLendingEventKey(event.Id), []byte{})

	k.IncreasePendingLendingEventCount(ctx)
}

// RemoveLendingEventFromPendingQueue removes the specified lending event from the pending queue
func (k Keeper) RemoveLendingEventFromPendingQueue(ctx sdk.Context, event *types.DLCEvent) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.PendingLendingEventKey(event.Id))

	k.DecreasePendingLendingEventCount(ctx)
}

// GetAvailableLendingEvent gets an available lending event and removes it from the pending queue if any
func (k Keeper) GetAvailableLendingEvent(ctx sdk.Context) *types.DLCEvent {
	var lendingEvent *types.DLCEvent

	k.IteratePendingLendingEvents(ctx, func(event *types.DLCEvent) (stop bool) {
		lendingEvent = event
		return true
	})

	if lendingEvent != nil {
		k.RemoveLendingEventFromPendingQueue(ctx, lendingEvent)
	}

	return lendingEvent
}

// TriggerDLCEvent triggers the given event
func (k Keeper) TriggerDLCEvent(ctx sdk.Context, id uint64, outcomeIndex int) {
	event := k.GetEvent(ctx, id)

	event.HasTriggered = true
	event.OutcomeIndex = int32(outcomeIndex)
	event.TriggerAt = ctx.BlockTime()

	k.SetEvent(ctx, event)

	k.tssKeeper.InitiateSigningRequest(
		ctx,
		types.ModuleName,
		types.ToScopedId(event.Id),
		tsstypes.SigningType_SIGNING_TYPE_SCHNORR_WITH_COMMITMENT,
		int32(types.SigningIntent_SIGNING_INTENT_DEFAULT),
		event.Pubkey,
		[]string{base64.StdEncoding.EncodeToString(types.GetEventOutcomeHash(event, outcomeIndex))},
		&tsstypes.SigningOptions{Nonce: event.Nonce},
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTriggerDLCEvent,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", event.Id)),
			sdk.NewAttribute(types.AttributeKeyDLCEventType, fmt.Sprintf("%d", event.Type)),
			sdk.NewAttribute(types.AttributeKeyOutcome, event.Outcomes[outcomeIndex]),
		),
	)
}

// GetAllEvents gets all events
func (k Keeper) GetAllEvents(ctx sdk.Context) []*types.DLCEvent {
	events := make([]*types.DLCEvent, 0)

	k.IterateEvents(ctx, func(event *types.DLCEvent) (stop bool) {
		events = append(events, event)
		return false
	})

	return events
}

// GetEvents gets events according to the specified status
func (k Keeper) GetEvents(ctx sdk.Context, triggered bool) []*types.DLCEvent {
	events := make([]*types.DLCEvent, 0)

	k.IterateEventsByStatus(ctx, triggered, func(event *types.DLCEvent) (stop bool) {
		events = append(events, event)
		return false
	})

	return events
}

// IterateEventsByStatus iterates through events by the given status
func (k Keeper) IterateEventsByStatus(ctx sdk.Context, triggered bool, cb func(event *types.DLCEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.DLCEvent
		k.cdc.MustUnmarshal(iterator.Value(), &event)

		if event.HasTriggered == triggered && cb(&event) {
			break
		}
	}
}

// IterateEvents iterates through all events
func (k Keeper) IterateEvents(ctx sdk.Context, cb func(event *types.DLCEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.EventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var event types.DLCEvent
		k.cdc.MustUnmarshal(iterator.Value(), &event)

		if cb(&event) {
			break
		}
	}
}

// IteratePendingLendingEvents iterates through the pending lending events
func (k Keeper) IteratePendingLendingEvents(ctx sdk.Context, cb func(event *types.DLCEvent) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.PendingLendingEventKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		event := k.GetEvent(ctx, sdk.BigEndianToUint64(key[1:]))

		if cb(event) {
			break
		}
	}
}
