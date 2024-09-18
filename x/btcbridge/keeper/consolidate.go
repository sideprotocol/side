package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// ConsolidateVaults performs the UTXO consolidation for the given vaults
func (k Keeper) ConsolidateVaults(ctx sdk.Context, vaultVersion uint64, btcConsolidation *types.BtcConsolidation, runesConsolidations []*types.RunesConsolidation, feeRate int64) error {
	if btcConsolidation != nil {
		if err := k.handleBtcConsolidation(ctx, vaultVersion, btcConsolidation.TargetThreshold, btcConsolidation.MaxNum, feeRate); err != nil {
			return err
		}
	}

	for _, rc := range runesConsolidations {
		if err := k.handleRunesConsolidation(ctx, vaultVersion, rc.RuneId, rc.TargetThreshold, rc.MaxNum, feeRate); err != nil {
			return err
		}
	}

	return nil
}

// handleBtcConsolidation handles the given btc consolidation
func (k Keeper) handleBtcConsolidation(ctx sdk.Context, vaultVersion uint64, targetThreshold int64, maxNum uint32, feeRate int64) error {
	vault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, vaultVersion)
	if vault == nil {
		return types.ErrVaultDoesNotExist
	}

	targetUTXOs := k.GetUnlockedUTXOsByAddrAndThreshold(ctx, vault.Address, targetThreshold, maxNum)
	if len(targetUTXOs) == 0 {
		return types.ErrInsufficientUTXOs
	}

	p, recipientUTXO, err := types.BuildTransferAllBtcPsbt(targetUTXOs, vault.Address, feeRate)
	if err != nil {
		return err
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return types.ErrFailToSerializePsbt
	}

	// lock the involved utxos
	_ = k.LockUTXOs(ctx, targetUTXOs)

	// save the recipient utxo and mark minted
	k.saveUTXO(ctx, recipientUTXO)
	k.addToMintHistory(ctx, p.UnsignedTx.TxHash().String())

	// set signing request
	signingReq := &types.SigningRequest{
		Address:  k.authority,
		Sequence: k.IncrementRequestSequence(ctx),
		Txid:     p.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}
	k.SetSigningRequest(ctx, signingReq)

	// Emit events
	k.EmitEvent(ctx, k.authority,
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", signingReq.Sequence)),
		sdk.NewAttribute("txid", signingReq.Txid),
	)

	return nil
}

// handleRunesConsolidation handles the given runes consolidation
func (k Keeper) handleRunesConsolidation(ctx sdk.Context, vaultVersion uint64, runeId string, targetThreshold string, maxNum uint32, feeRate int64) error {
	vault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_RUNES, vaultVersion)
	if vault == nil {
		return types.ErrVaultDoesNotExist
	}

	btcVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, vaultVersion)
	if btcVault == nil {
		return types.ErrVaultDoesNotExist
	}

	targetRunesUTXOs, runeBalances := k.GetTargetRunesUTXOsByAddrAndThreshold(ctx, vault.Address, runeId, types.RuneAmountFromString(targetThreshold), maxNum)
	if len(targetRunesUTXOs) == 0 {
		return types.ErrInsufficientUTXOs
	}

	btcUtxoIterator := k.GetUTXOIteratorByAddr(ctx, btcVault.Address)

	p, selectedUtxos, changeUtxo, runesRecipientUtxo, err := types.BuildTransferAllRunesPsbt(targetRunesUTXOs, btcUtxoIterator, vault.Address, runeBalances, feeRate, btcVault.Address)
	if err != nil {
		return err
	}

	psbtB64, err := p.B64Encode()
	if err != nil {
		return types.ErrFailToSerializePsbt
	}

	// lock the involved utxos
	_ = k.LockUTXOs(ctx, targetRunesUTXOs)
	_ = k.LockUTXOs(ctx, selectedUtxos)

	// save the change utxo
	if changeUtxo != nil {
		k.saveUTXO(ctx, changeUtxo)
	}

	// save the runes recipient utxo
	k.saveUTXO(ctx, runesRecipientUtxo)

	// mark minted
	k.addToMintHistory(ctx, p.UnsignedTx.TxHash().String())

	// set signing request
	signingReq := &types.SigningRequest{
		Address:  k.authority,
		Sequence: k.IncrementRequestSequence(ctx),
		Txid:     p.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	// Emit events
	k.EmitEvent(ctx, k.authority,
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", signingReq.Sequence)),
		sdk.NewAttribute("txid", signingReq.Txid),
	)

	return nil
}
