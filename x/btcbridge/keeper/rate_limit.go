package keeper

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// SetRateLimit sets the rate limit
func (k Keeper) SetRateLimit(ctx sdk.Context, rateLimit *types.RateLimit) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(rateLimit)

	store.Set(types.RateLimitKey, bz)
}

// HasRateLimit returns true if the rate limit exists, false otherwise
func (k Keeper) HasRateLimit(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RateLimitKey)
}

// GetRateLimit gets the current rate limit
func (k Keeper) GetRateLimit(ctx sdk.Context) *types.RateLimit {
	store := ctx.KVStore(k.storeKey)

	var rateLimit types.RateLimit
	bz := store.Get(types.RateLimitKey)
	k.cdc.MustUnmarshal(bz, &rateLimit)

	return &rateLimit
}

// SetAddressRateLimitDetails sets the per address rate limit details
func (k Keeper) SetAddressRateLimitDetails(ctx sdk.Context, rateLimitDetails *types.AddressRateLimitDetails) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(rateLimitDetails)

	store.Set(types.RateLimitByAddressKey(rateLimitDetails.Address), bz)
}

// HasAddressRateLimitDetails returns true if the rate limit details exist for the given address, false otherwise
func (k Keeper) HasAddressRateLimitDetails(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RateLimitByAddressKey(address))
}

// GetAddressRateLimitDetails gets the rate limit details by the given address
func (k Keeper) GetAddressRateLimitDetails(ctx sdk.Context, address string) *types.AddressRateLimitDetails {
	store := ctx.KVStore(k.storeKey)

	var rateLimitDetails types.AddressRateLimitDetails
	bz := store.Get(types.RateLimitByAddressKey(address))
	k.cdc.MustUnmarshal(bz, &rateLimitDetails)

	return &rateLimitDetails
}

// RemoveAllAddressRateLimitDetails removes all address rate limit details
func (k Keeper) RemoveAllAddressRateLimitDetails(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	k.IterateAddressRateLimitDetails(ctx, func(rateLimitDetails *types.AddressRateLimitDetails) (stop bool) {
		store.Delete(types.RateLimitByAddressKey(rateLimitDetails.Address))
		return false
	})
}

// IterateAddressRateLimitDetails iterates through all per address rate limit details
func (k Keeper) IterateAddressRateLimitDetails(ctx sdk.Context, cb func(rateLimitDetails *types.AddressRateLimitDetails) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.RateLimitByAddressKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var addressRateLimitDetails types.AddressRateLimitDetails
		k.cdc.MustUnmarshal(iterator.Value(), &addressRateLimitDetails)

		if cb(&addressRateLimitDetails) {
			break
		}
	}
}

// HandleRateLimit performs the rate limit handling
func (k Keeper) HandleRateLimit(ctx sdk.Context, address string, amount sdk.Coin) error {
	// only the BTC rate limit is supported currently
	if amount.Denom != k.BtcDenom(ctx) {
		return nil
	}

	// check rate limit
	if err := k.CheckRateLimit(ctx, address, amount.Amount.Int64()); err != nil {
		return err
	}

	// update rate limit
	k.UpdateRateLimitUsedQuotas(ctx, address, amount.Amount.Int64())

	return nil
}

// CheckRateLimit checks if the given address and amount satisfies the rate limit
func (k Keeper) CheckRateLimit(ctx sdk.Context, address string, amount int64) error {
	if !k.HasRateLimit(ctx) {
		return nil
	}

	rateLimit := k.GetRateLimit(ctx)

	// check per address rate limit
	if err := k.CheckAddressRateLimit(ctx, rateLimit, address, amount); err != nil {
		return err
	}

	// check global rate limit
	if err := k.CheckGlobalRateLimit(ctx, rateLimit, amount); err != nil {
		return err
	}

	return nil
}

// CheckGlobalRateLimit checks if the given amount satisfies the global rate limit
func (k Keeper) CheckGlobalRateLimit(ctx sdk.Context, rateLimit *types.RateLimit, amount int64) error {
	if !types.GlobalRateLimitEnabled(rateLimit) {
		return nil
	}

	globalRateLimit := rateLimit.GlobalRateLimit

	if globalRateLimit.Used+amount > globalRateLimit.Quota {
		return errorsmod.Wrapf(types.ErrRateLimitReached, "global rate limit reached; period: %s-%s, quota: %d, remaining: %d", globalRateLimit.StartTime, globalRateLimit.EndTime, globalRateLimit.Quota, globalRateLimit.Quota-globalRateLimit.Used)
	}

	return nil
}

// CheckAddressRateLimit checks if the given amount satisfies the per address rate limit for the given address
func (k Keeper) CheckAddressRateLimit(ctx sdk.Context, rateLimit *types.RateLimit, address string, amount int64) error {
	if !types.AddressRateLimitEnabled(rateLimit) {
		return nil
	}

	addressRateLimit := rateLimit.AddressRateLimit

	if !k.HasAddressRateLimitDetails(ctx, address) {
		if amount > addressRateLimit.Quota {
			return errorsmod.Wrapf(types.ErrRateLimitReached, "address rate limit reached; period: %s-%s, quota: %d, remaining: %d", addressRateLimit.StartTime, addressRateLimit.EndTime, addressRateLimit.Quota, addressRateLimit.Quota)
		}

		return nil
	}

	rateLimitDetails := k.GetAddressRateLimitDetails(ctx, address)
	if rateLimitDetails.Used+amount > addressRateLimit.Quota {
		return errorsmod.Wrapf(types.ErrRateLimitReached, "address rate limit reached; period: %s-%s, quota: %d, remaining: %d", &addressRateLimit.StartTime, addressRateLimit.EndTime, addressRateLimit.Quota, addressRateLimit.Quota-rateLimitDetails.Used)
	}

	return nil
}

// UpdateRateLimitUsedQuotas updates the used quotas of the rate limit by the given delta
func (k Keeper) UpdateRateLimitUsedQuotas(ctx sdk.Context, address string, amount int64) {
	rateLimit := k.GetRateLimit(ctx)

	// update global rate limit
	rateLimit.GlobalRateLimit.Used += amount

	// update address rate limit
	var rateLimitDetails *types.AddressRateLimitDetails
	if !k.HasAddressRateLimitDetails(ctx, address) {
		rateLimitDetails = &types.AddressRateLimitDetails{
			Address: address,
			Used:    amount,
		}
	} else {
		rateLimitDetails = k.GetAddressRateLimitDetails(ctx, address)
		rateLimitDetails.Used += amount
	}

	k.SetRateLimit(ctx, rateLimit)
	k.SetAddressRateLimitDetails(ctx, rateLimitDetails)
}

// UpdateRateLimitTotalQuotas updates the total quotas of the rate limit
func (k Keeper) UpdateRateLimitTotalQuotas(ctx sdk.Context, globalSupplyPercentageQuota uint32, addressRateLimitQuota int64) {
	rateLimit := k.GetRateLimit(ctx)

	// update global rate limit
	rateLimit.GlobalRateLimit.Quota = k.GetGlobalRateLimitQuota(ctx, globalSupplyPercentageQuota)

	// update address rate limit
	rateLimit.AddressRateLimit.Quota = addressRateLimitQuota

	k.SetRateLimit(ctx, rateLimit)
}

// NewRateLimit creates the rate limit for the new epoch
func (k Keeper) NewRateLimit(ctx sdk.Context) *types.RateLimit {
	return &types.RateLimit{
		GlobalRateLimit:  k.NewGlobalRateLimit(ctx),
		AddressRateLimit: k.NewAddressRateLimit(ctx),
	}
}

// NewGlobalRateLimit creates the global rate limit for the new epoch
func (k Keeper) NewGlobalRateLimit(ctx sdk.Context) types.GlobalRateLimit {
	return types.GlobalRateLimit{
		StartTime: ctx.BlockTime(),
		EndTime:   ctx.BlockTime().Add(k.GlobalRateLimitPeriod(ctx)),
		Quota:     k.GetGlobalRateLimitQuota(ctx, k.GlobalRateLimitSupplyPercentageQuota(ctx)),
	}
}

// NewAddressRateLimit creates the per address rate limit for the new epoch
func (k Keeper) NewAddressRateLimit(ctx sdk.Context) types.AddressRateLimit {
	return types.AddressRateLimit{
		StartTime: ctx.BlockTime(),
		EndTime:   ctx.BlockTime().Add(k.AddressRateLimitPeriod(ctx)),
		Quota:     k.AddressRateLimitQuota(ctx),
	}
}

// GetGlobalRateLimitQuota gets the global rate limit quota according to the given supply percentage quota
func (k Keeper) GetGlobalRateLimitQuota(ctx sdk.Context, supplyPercentageQuota uint32) int64 {
	if supplyPercentageQuota == 100 {
		// no limit
		return 0
	}

	// get the current sBTC supply
	supply := k.bankKeeper.GetSupply(ctx, k.BtcDenom(ctx)).Amount.Int64()

	return supply * int64(supplyPercentageQuota) / 100
}
