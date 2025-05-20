package keeper

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// TransferVault performs the vault asset transfer from the source version to the destination version
func (k Keeper) TransferVault(ctx sdk.Context, sourceVersion uint64, destVersion uint64, assetType types.AssetType, psbts []string, targetUtxoNum uint32) error {
	sourceVault := k.GetVaultByAssetTypeAndVersion(ctx, assetType, sourceVersion)
	if sourceVault == nil {
		return types.ErrVaultDoesNotExist
	}

	destVault := k.GetVaultByAssetTypeAndVersion(ctx, assetType, destVersion)
	if destVault == nil {
		return types.ErrVaultDoesNotExist
	}

	// handle pre-built psbts if any
	if len(psbts) > 0 {
		for i := range psbts {
			p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(psbts[i])), true)

			if err := k.handleTransferVaultTx(ctx, p, sourceVault, destVault, assetType); err != nil {
				return err
			}

			signingReq := &types.SigningRequest{
				Address:      k.authority,
				Sequence:     k.IncrementSigningRequestSequence(ctx),
				Type:         assetType,
				Txid:         p.UnsignedTx.TxHash().String(),
				Psbt:         psbts[i],
				CreationTime: ctx.BlockTime(),
				Status:       types.SigningStatus_SIGNING_STATUS_PENDING,
			}

			k.SetSigningRequest(ctx, signingReq)
		}

		return nil
	}

	feeRate := k.GetFeeRate(ctx)
	if err := k.CheckFeeRate(ctx, feeRate); err != nil {
		return err
	}

	var err error
	var signingReq *types.SigningRequest

	switch assetType {
	case types.AssetType_ASSET_TYPE_BTC:
		signingReq, err = k.BuildTransferVaultBtcSigningRequest(ctx, sourceVault, destVault, targetUtxoNum, feeRate.Value)
		if err != nil {
			return err
		}

	case types.AssetType_ASSET_TYPE_RUNES:
		signingReq, err = k.BuildTransferVaultRunesSigningRequest(ctx, sourceVault, destVault, targetUtxoNum, feeRate.Value)
		if err != nil {
			return err
		}
	}

	k.SetSigningRequest(ctx, signingReq)

	return nil
}

// handleTransferVaultTx handles the pre-built tx for the vault transfer
func (k Keeper) handleTransferVaultTx(ctx sdk.Context, p *psbt.Packet, sourceVault, destVault *types.Vault, assetType types.AssetType) error {
	if err := k.checkUtxoCount(ctx, len(p.UnsignedTx.TxIn)); err != nil {
		return err
	}

	txHash := p.UnsignedTx.TxHash().String()

	if assetType == types.AssetType_ASSET_TYPE_RUNES {
		if edicts, err := types.ParseRunes(p.UnsignedTx); err != nil || len(edicts) != types.RunesEdictNum {
			return types.ErrInvalidRunes
		}
	}

	runeBalances := make(types.RuneBalances, 0)

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
			runeBalances = runeBalances.Merge(utxo.Runes)
		}

		_ = k.SpendUTXO(ctx, hash, uint64(vout))
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
					IsLocked:     true,
				}

				k.SetUTXO(ctx, utxo)
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
					IsLocked:     true,
					Runes:        runeBalances,
				}

				k.SetUTXO(ctx, utxo)
			}
		}
	}

	// mark minted
	k.addToMintHistory(ctx, p.UnsignedTx.TxHash().String())

	return nil
}

// BuildTransferVaultBtcSigningRequest builds the signing request to transfer btc of the given vault
func (k Keeper) BuildTransferVaultBtcSigningRequest(ctx sdk.Context, sourceVault *types.Vault, destVault *types.Vault, targetUtxoNum uint32, feeRate int64) (*types.SigningRequest, error) {
	utxos := make([]*types.UTXO, 0)

	maxUtxoNum := uint32(k.GetMaxUtxoNum(ctx))
	if targetUtxoNum > maxUtxoNum {
		targetUtxoNum = maxUtxoNum
	}

	k.IterateUnlockedUTXOsByAddr(ctx, sourceVault.Address, func(addr string, utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)

		return len(utxos) >= int(targetUtxoNum)
	})

	if len(utxos) == 0 {
		return nil, types.ErrInsufficientUTXOs
	}

	p, recipientUTXO, err := types.BuildTransferAllBtcPsbt(utxos, destVault.Address, feeRate)
	if err != nil {
		return nil, err
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	txHash := p.UnsignedTx.TxHash().String()

	// spend the involved utxos
	_ = k.SpendUTXOs(ctx, utxos)

	// lock the recipient(change) utxo
	k.lockChangeUTXOs(ctx, txHash, recipientUTXO)

	signingReq := &types.SigningRequest{
		Address:      k.authority,
		Sequence:     k.IncrementSigningRequestSequence(ctx),
		Type:         types.AssetType_ASSET_TYPE_BTC,
		Txid:         txHash,
		Psbt:         psbtB64,
		CreationTime: ctx.BlockTime(),
		Status:       types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	return signingReq, nil
}

// BuildTransferVaultRunesSigningRequest builds the signing request to transfer runes of the given vault
func (k Keeper) BuildTransferVaultRunesSigningRequest(ctx sdk.Context, sourceVault *types.Vault, destVault *types.Vault, targetUtxoNum uint32, feeRate int64) (*types.SigningRequest, error) {
	runesUtxos := make([]*types.UTXO, 0)
	runeBalances := make(types.RuneBalances, 0)

	maxUtxoNum := uint32(k.GetMaxUtxoNum(ctx))
	if targetUtxoNum > maxUtxoNum {
		targetUtxoNum = maxUtxoNum
	}

	k.IterateUnlockedUTXOsByAddr(ctx, sourceVault.Address, func(addr string, utxo *types.UTXO) (stop bool) {
		runesUtxos = append(runesUtxos, utxo)
		runeBalances = runeBalances.Merge(utxo.Runes)

		return len(runesUtxos) >= int(targetUtxoNum)
	})

	if len(runesUtxos) == 0 {
		return nil, types.ErrInsufficientUTXOs
	}

	sourceBtcVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, sourceVault.Version)
	if sourceBtcVault == nil {
		return nil, types.ErrVaultDoesNotExist
	}

	destBtcVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, destVault.Version)
	if destBtcVault == nil {
		return nil, types.ErrVaultDoesNotExist
	}

	btcUtxoIterator := k.GetUTXOIteratorByAddr(ctx, sourceBtcVault.Address)

	p, selectedUtxos, changeUtxo, runesRecipientUtxo, err := types.BuildTransferAllRunesPsbt(runesUtxos, btcUtxoIterator, destVault.Address, runeBalances, feeRate, destBtcVault.Address, k.GetMaxUtxoNum(ctx))
	if err != nil {
		return nil, err
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	txHash := p.UnsignedTx.TxHash().String()

	// spend the involved utxos
	_ = k.SpendUTXOs(ctx, runesUtxos)
	_ = k.SpendUTXOs(ctx, selectedUtxos)

	// lock the change utxos
	k.lockChangeUTXOs(ctx, txHash, changeUtxo, runesRecipientUtxo)

	signingReq := &types.SigningRequest{
		Address:      k.authority,
		Sequence:     k.IncrementSigningRequestSequence(ctx),
		Type:         types.AssetType_ASSET_TYPE_RUNES,
		Txid:         txHash,
		Psbt:         psbtB64,
		CreationTime: ctx.BlockTime(),
		Status:       types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	return signingReq, nil
}

// CheckVaults checks if the provided vaults are valid
func (k Keeper) CheckVaults(ctx sdk.Context, vaults []string, vaultTypes []types.AssetType) error {
	currentVaults := k.GetParams(ctx).Vaults

	if len(vaults) != len(vaultTypes) {
		return errorsmod.Wrap(types.ErrInvalidDKGCompletionRequest, "invalid vaults")
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

// SetVaultVersion sets the vault version
func (k Keeper) SetVaultVersion(ctx sdk.Context, version uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.VaultVersionKey, sdk.Uint64ToBigEndian(version))
}

// VaultTransferCompleted returns true if the asset transfer completed for the given vault, false otherwise
func (k Keeper) VaultTransferCompleted(ctx sdk.Context, vault string) bool {
	completed := true

	k.IterateUnlockedUTXOsByAddr(ctx, vault, func(addr string, utxo *types.UTXO) (stop bool) {
		completed = false
		return true
	})

	return completed
}

// VaultsTransferCompleted returns true if all asset transfer completed for the given vault version, false otherwise
func (k Keeper) VaultsTransferCompleted(ctx sdk.Context, version uint64) bool {
	btcVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, version).Address
	runesVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_RUNES, version).Address

	return k.VaultTransferCompleted(ctx, btcVault) && k.VaultTransferCompleted(ctx, runesVault)
}
