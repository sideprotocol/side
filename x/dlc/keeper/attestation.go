package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/hash"
	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/dlc/types"
)

// LiquidatioinOutcome represents the dlc outcome string for liquidation
const LiquidatioinOutcome = "Liquidated"

// HandleAttestation performs the attestation handling
func (k Keeper) HandleAttestation(ctx sdk.Context, sender string, eventId uint64, signature string) error {
	if !k.HasEvent(ctx, eventId) {
		return types.ErrEventDoesNotExist
	}

	event := k.GetEvent(ctx, eventId)

	pubKeyBytes, _ := hex.DecodeString(event.Pubkey)
	sigBytes, _ := hex.DecodeString(signature)
	msg := hash.Sha256(sdk.Uint64ToBigEndian(event.TriggerPrice.Uint64()))

	if !schnorr.Verify(sigBytes, msg, pubKeyBytes) {
		return errorsmod.Wrap(types.ErrInvalidSignature, "failed to verify the signature")
	}

	attestation := types.DLCAttestation{
		Id:        k.IncrementAttestationId(ctx),
		EventId:   eventId,
		Outcome:   LiquidatioinOutcome,
		Pubkey:    event.Pubkey,
		Signature: signature,
		Time:      ctx.BlockTime(),
	}

	k.SetAttestation(ctx, &attestation)

	return nil
}

// GetAttestationId gets the current attestation id
func (k Keeper) GetAttestationId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AttestationIdKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementAttestationId increments the attestation id
func (k Keeper) IncrementAttestationId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetAttestationId(ctx) + 1
	store.Set(types.AttestationIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// GetAttestation gets the attestation by the given id
func (k Keeper) GetAttestation(ctx sdk.Context, id uint64) *types.DLCAttestation {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AttestationKey(id))
	var attestation types.DLCAttestation
	k.cdc.MustUnmarshal(bz, &attestation)

	return &attestation
}

// SetAttestation sets the given attestation
func (k Keeper) SetAttestation(ctx sdk.Context, attestation *types.DLCAttestation) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(attestation)
	store.Set(types.AttestationKey(attestation.Id), bz)
}

// GetAttestations gets attestations
func (k Keeper) GetAttestations(ctx sdk.Context) []*types.DLCAttestation {
	attestations := make([]*types.DLCAttestation, 0)

	k.IterateAttestations(ctx, func(attestation *types.DLCAttestation) (stop bool) {
		attestations = append(attestations, attestation)
		return false
	})

	return attestations
}

// IterateAttestations iterates through all attestations
func (k Keeper) IterateAttestations(ctx sdk.Context, cb func(attestation *types.DLCAttestation) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.AttestationKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var attestation types.DLCAttestation
		k.cdc.MustUnmarshal(iterator.Value(), &attestation)

		if cb(&attestation) {
			break
		}
	}
}
