package keeper

import (
	"bytes"
	"time"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetNextDKGRequestID gets the next DKG request ID
func (keeper Keeper) GetNextDKGRequestID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.DKGRequestIDKey)
	if bz == nil {
		return 1
	}

	return sdk.BigEndianToUint64(bz) + 1
}

// SetDKGRequestID sets the current DKG request ID
func (keeper Keeper) SetDKGRequestID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.DKGRequestIDKey, sdk.Uint64ToBigEndian(id))
}

// SetDKGRequest sets the given DKG request
func (k Keeper) SetDKGRequest(ctx sdk.Context, req *types.DKGRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGRequestKey(req.Id), bz)
}

// GetDKGRequest gets the DKG request by the given id
func (k Keeper) GetDKGRequest(ctx sdk.Context, id uint64) *types.DKGRequest {
	store := ctx.KVStore(k.storeKey)

	var req types.DKGRequest
	bz := store.Get(types.DKGRequestKey(id))
	k.cdc.MustUnmarshal(bz, &req)

	return &req
}

// GetDKGRequests gets the DKG requests by the given status
func (k Keeper) GetDKGRequests(ctx sdk.Context, status types.DKGRequestStatus) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		if req.Status == status {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetPendingDKGRequests gets the pending DKG requests
func (k Keeper) GetPendingDKGRequests(ctx sdk.Context) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		if req.Status == types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING {
			requests = append(requests, req)
		}

		return false
	})

	return requests
}

// GetAllDKGRequests gets all DKG requests
func (k Keeper) GetAllDKGRequests(ctx sdk.Context) []*types.DKGRequest {
	requests := make([]*types.DKGRequest, 0)

	k.IterateDKGRequests(ctx, func(req *types.DKGRequest) (stop bool) {
		requests = append(requests, req)
		return false
	})

	return requests
}

// IterateDKGRequests iterates through all DKG requests
func (k Keeper) IterateDKGRequests(ctx sdk.Context, cb func(req *types.DKGRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.DKGRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// GetDKGRequestExpirationTime gets the expiration time of the DKG request
func (k Keeper) GetDKGRequestExpirationTime(ctx sdk.Context) *time.Time {
	creationTime := ctx.BlockTime()
	timeout := k.GetParams(ctx).TssParams.DkgTimeoutPeriod

	expiration := creationTime.Add(timeout)

	return &expiration
}

// SetDKGCompletionRequest sets the given DKG completion request
func (k Keeper) SetDKGCompletionRequest(ctx sdk.Context, req *types.DKGCompletionRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.DKGCompletionRequestKey(req.Id, req.ConsensusAddress), bz)
}

// HasDKGCompletionRequest returns true if the given completion request exists, false otherwise
func (k Keeper) HasDKGCompletionRequest(ctx sdk.Context, id uint64, consAddress string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DKGCompletionRequestKey(id, consAddress))
}

// GetDKGCompletionRequests gets DKG completion requests by the given id
func (k Keeper) GetDKGCompletionRequests(ctx sdk.Context, id uint64) []*types.DKGCompletionRequest {
	requests := make([]*types.DKGCompletionRequest, 0)

	k.IterateDKGCompletionRequests(ctx, id, func(req *types.DKGCompletionRequest) (stop bool) {
		requests = append(requests, req)
		return false
	})

	return requests
}

// IterateDKGCompletionRequests iterates through all DKG completion requests by the given id
func (k Keeper) IterateDKGCompletionRequests(ctx sdk.Context, id uint64, cb func(req *types.DKGCompletionRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.DKGCompletionRequestKeyPrefix, sdk.Uint64ToBigEndian(id)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var req types.DKGCompletionRequest
		k.cdc.MustUnmarshal(iterator.Value(), &req)

		if cb(&req) {
			break
		}
	}
}

// InitiateDKG initiates the DKG request by the specified params
func (k Keeper) InitiateDKG(ctx sdk.Context, participants []*types.DKGParticipant, threshold uint32, vaultTypes []types.AssetType) (*types.DKGRequest, error) {
	for _, p := range participants {
		consAddress, _ := sdk.ConsAddressFromHex(p.ConsensusAddress)

		validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddress)
		if !found {
			return nil, sdkerrors.Wrap(types.ErrInvalidDKGParams, "non validator")
		}

		if validator.Status != stakingtypes.Bonded {
			return nil, sdkerrors.Wrap(types.ErrInvalidDKGParams, "validator not bonded")
		}
	}

	req := &types.DKGRequest{
		Id:           k.GetNextDKGRequestID(ctx),
		Participants: participants,
		Threshold:    threshold,
		VaultTypes:   vaultTypes,
		Expiration:   k.GetDKGRequestExpirationTime(ctx),
		Status:       types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING,
	}

	k.SetDKGRequest(ctx, req)
	k.SetDKGRequestID(ctx, req.Id)

	return req, nil
}

// CompleteDKG completes the DKG request by the DKG participant
// The DKG request will be completed when all participants submit the valid completion request before timeout
func (k Keeper) CompleteDKG(ctx sdk.Context, req *types.DKGCompletionRequest) error {
	dkgReq := k.GetDKGRequest(ctx, req.Id)
	if dkgReq == nil {
		return types.ErrDKGRequestDoesNotExist
	}

	if !types.ParticipantExists(dkgReq.Participants, req.ConsensusAddress) {
		return types.ErrUnauthorizedDKGCompletionRequest
	}

	if k.HasDKGCompletionRequest(ctx, req.Id, req.ConsensusAddress) {
		return types.ErrDKGCompletionRequestExists
	}

	if dkgReq.Status != types.DKGRequestStatus_DKG_REQUEST_STATUS_PENDING {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid dkg request status")
	}

	if !ctx.BlockTime().Before(*dkgReq.Expiration) {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "dkg request expired")
	}

	if err := k.CheckVaults(ctx, req.Vaults, dkgReq.VaultTypes); err != nil {
		return err
	}

	consAddress, _ := sdk.ConsAddressFromHex(req.ConsensusAddress)
	validator, found := k.stakingKeeper.GetValidatorByConsAddr(ctx, consAddress)
	if !found {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "non validator")
	}

	if validator.Status != stakingtypes.Bonded {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "validator not bonded")
	}

	pubKey, err := validator.ConsPubKey()
	if err != nil {
		return err
	}

	if !types.VerifySignature(req.Signature, pubKey.Bytes(), req) {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid signature")
	}

	k.SetDKGCompletionRequest(ctx, req)

	return nil
}

// TransferVault performs the vault asset transfer from the source version to the destination version
func (k Keeper) TransferVault(ctx sdk.Context, sourceVersion uint64, destVersion uint64, assetType types.AssetType, psbts []string) error {
	sourceVault := k.GetVaultByAssetTypeAndVersion(ctx, assetType, sourceVersion)
	if sourceVault == nil {
		return types.ErrVaultDoesNotExist
	}

	destVault := k.GetVaultByAssetTypeAndVersion(ctx, assetType, destVersion)
	if destVault == nil {
		return types.ErrVaultDoesNotExist
	}

	for i := range psbts {
		p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(psbts[i])), true)

		if err := k.handleTransferTx(ctx, p, sourceVault, destVault, assetType); err != nil {
			return err
		}

		signingReq := &types.BitcoinWithdrawRequest{
			Address:  k.authority,
			Sequence: k.IncrementRequestSequence(ctx),
			Txid:     p.UnsignedTx.TxHash().String(),
			Psbt:     psbts[i],
			Status:   types.WithdrawStatus_WITHDRAW_STATUS_CREATED,
		}

		k.SetWithdrawRequest(ctx, signingReq)
	}

	return nil
}

// handleTransferTx handles the specified tx for the vault transfer
func (k Keeper) handleTransferTx(ctx sdk.Context, p *psbt.Packet, sourceVault, destVault *types.Vault, assetType types.AssetType) error {
	txHash := p.UnsignedTx.TxHash().String()

	if assetType == types.AssetType_ASSET_TYPE_RUNES {
		if edicts, err := types.ParseRunes(p.UnsignedTx); err != nil || len(edicts) != types.RunesEdictNum {
			return types.ErrInvalidRunes
		}
	}

	runeBalances := make([]*types.RuneBalance, 0)

	for i, ti := range p.UnsignedTx.TxIn {
		hash := ti.PreviousOutPoint.Hash.String()
		vout := ti.PreviousOutPoint.Index

		if !k.HasUTXO(ctx, hash, uint64(vout)) {
			return types.ErrUTXODoesNotExist
		}

		if k.IsUTXOLocked(ctx, hash, uint64(vout)) {
			return types.ErrUTXOLocked
		}

		utxo := k.GetUTXO(ctx, hash, uint64(vout))
		if !bytes.Equal(utxo.PubKeyScript, p.Inputs[i].WitnessUtxo.PkScript) || utxo.Amount != uint64(p.Inputs[i].WitnessUtxo.Value) {
			return types.ErrInvalidPsbt
		}

		vault := types.SelectVaultByAddress(k.GetParams(ctx).Vaults, utxo.Address)
		if vault == nil {
			return types.ErrVaultDoesNotExist
		}

		if vault.Version != sourceVault.Version {
			return types.ErrInvalidVaultVersion
		}

		if assetType == types.AssetType_ASSET_TYPE_BTC && vault.AssetType != sourceVault.AssetType {
			return types.ErrInvalidVault
		}

		if assetType == types.AssetType_ASSET_TYPE_RUNES && vault.AssetType == types.AssetType_ASSET_TYPE_RUNES {
			runeBalances = append(runeBalances, utxo.Runes...)
		}
	}

	for i, out := range p.UnsignedTx.TxOut {
		if !txscript.IsNullData(out.PkScript) {
			vault := types.SelectVaultByPkScript(k.GetParams(ctx).Vaults, out.PkScript)
			if vault == nil {
				return types.ErrVaultDoesNotExist
			}

			if vault.Version != destVault.Version {
				return types.ErrInvalidVault
			}

			if assetType == types.AssetType_ASSET_TYPE_BTC && vault.AssetType != destVault.AssetType {
				return types.ErrInvalidVault
			}

			if vault.AssetType == types.AssetType_ASSET_TYPE_RUNES && i != 1 {
				return types.ErrInvalidRunes
			}

			if vault.AssetType == types.AssetType_ASSET_TYPE_BTC {
				utxo := &types.UTXO{
					Txid:         txHash,
					Vout:         uint64(i),
					Address:      vault.Address,
					Amount:       uint64(out.Value),
					PubKeyScript: out.PkScript,
					IsLocked:     false,
				}

				k.saveUTXO(ctx, utxo)
			}

			if vault.AssetType == types.AssetType_ASSET_TYPE_RUNES {
				if len(runeBalances) == 0 {
					return types.ErrInvalidRunes
				}

				utxo := &types.UTXO{
					Txid:         txHash,
					Vout:         uint64(i),
					Address:      vault.Address,
					Amount:       uint64(out.Value),
					PubKeyScript: out.PkScript,
					IsLocked:     false,
					Runes:        types.GetCompactRuneBalances(runeBalances),
				}

				k.saveUTXO(ctx, utxo)
			}
		}
	}

	return nil
}

// CheckVaults checks if the provided vaults are valid
func (k Keeper) CheckVaults(ctx sdk.Context, vaults []string, vaultTypes []types.AssetType) error {
	currentVaults := k.GetParams(ctx).Vaults

	if len(vaults) != len(vaultTypes) {
		return sdkerrors.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid vaults")
	}

	for _, v := range vaults {
		if types.SelectVaultByAddress(currentVaults, v) != nil {
			return types.ErrInvalidDKGCompletionRequest
		}
	}

	return nil
}

// UpdateVaults updates the asset vaults of the btc bridge
// Assume that vaults are validated and match vault types
func (k Keeper) UpdateVaults(ctx sdk.Context, newVaults []string, vaultTypes []types.AssetType) {
	params := k.GetParams(ctx)

	version := k.IncreaseVaultVersion(ctx)

	for i, v := range newVaults {
		newVault := &types.Vault{
			Address:   v,
			AssetType: vaultTypes[i],
			Version:   version,
		}

		params.Vaults = append(params.Vaults, newVault)
	}

	k.SetParams(ctx, params)
}

// IncreaseVaultVersion increases the vault version by 1
func (k Keeper) IncreaseVaultVersion(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	version := k.GetLatestVaultVersion(ctx)

	store.Set(types.VaultVersionKey, sdk.Uint64ToBigEndian(version+1))

	return version + 1
}

// GetLatestVaultVersion gets the latest vault version
func (k Keeper) GetLatestVaultVersion(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.VaultVersionKey)
	if bz != nil {
		return sdk.BigEndianToUint64(bz)
	}

	return 0
}
