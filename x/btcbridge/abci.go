package btcbridge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/keeper"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleBtcWithdrawRequests(ctx, k)
	handleDKGRequests(ctx, k)
	handleVaultTransfer(ctx, k)
}

// handleBtcWithdrawRequests performs the batch btc withdrawal request handling
func handleBtcWithdrawRequests(ctx sdk.Context, k keeper.Keeper) {
	p := k.GetParams(ctx)

	// check if withdrawal is enabled
	if !p.WithdrawEnabled {
		return
	}

	// check block height
	if ctx.BlockHeight()%p.WithdrawParams.BtcBatchWithdrawPeriod != 0 {
		return
	}

	// get the pending btc withdrawal request
	pendingWithdrawRequests := k.GetPendingBtcWithdrawRequests(ctx, p.WithdrawParams.MaxBtcBatchWithdrawNum)
	if len(pendingWithdrawRequests) == 0 {
		return
	}

	feeRate := k.GetFeeRate(ctx)
	if feeRate == 0 {
		k.Logger(ctx).Error("invalid fee rate", feeRate)
		return
	}

	vault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_BTC)
	if vault == nil {
		k.Logger(ctx).Error("btc vault does not exist")
		return
	}

	signingRequest, err := k.BuildBtcBatchWithdrawSigningRequest(ctx, pendingWithdrawRequests, feeRate, vault.Address)
	if err != nil {
		k.Logger(ctx).Error("failed to build signing request", "err", err)
		return
	}

	for _, req := range pendingWithdrawRequests {
		// update withdrawal request
		req.Txid = signingRequest.Txid
		k.SetWithdrawRequest(ctx, req)

		// remove from the pending queue
		k.RemoveFromBtcWithdrawRequestQueue(ctx, req)

		// emit event
		k.EmitEvent(ctx, req.Address,
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", req.Sequence)),
			sdk.NewAttribute("txid", req.Txid),
		)
	}
}

// handleDKGRequests performs the DKG request handling
func handleDKGRequests(ctx sdk.Context, k keeper.Keeper) {
	pendingDKGRequests := k.GetPendingDKGRequests(ctx)

	for _, req := range pendingDKGRequests {
		// check if the DKG request expired
		if !ctx.BlockTime().Before(*req.Expiration) {
			req.Status = types.DKGRequestStatus_DKG_REQUEST_STATUS_TIMEDOUT
			k.SetDKGRequest(ctx, req)

			continue
		}

		// handle DKG completion requests
		completionRequests := k.GetDKGCompletionRequests(ctx, req.Id)
		if len(completionRequests) != len(req.Participants) {
			continue
		}

		// check if the DKG completion requests are valid
		if !types.CheckDKGCompletionRequests(completionRequests) {
			req.Status = types.DKGRequestStatus_DKG_REQUEST_STATUS_FAILED
			k.SetDKGRequest(ctx, req)

			continue
		}

		// update vaults
		k.UpdateVaults(ctx, completionRequests[0].Vaults, req.VaultTypes)

		// update status
		req.Status = types.DKGRequestStatus_DKG_REQUEST_STATUS_COMPLETED
		k.SetDKGRequest(ctx, req)
	}
}

// handleVaultTransfer performs the vault asset transfer
func handleVaultTransfer(ctx sdk.Context, k keeper.Keeper) {
	completedDKGRequests := k.GetDKGRequests(ctx, types.DKGRequestStatus_DKG_REQUEST_STATUS_COMPLETED)

	for _, req := range completedDKGRequests {
		if req.EnableTransfer {
			completions := k.GetDKGCompletionRequests(ctx, req.Id)
			dkgVaultVersion, _ := k.GetVaultVersionByAddress(ctx, completions[0].Vaults[0])

			sourceVersion := dkgVaultVersion - 1
			destVersion := k.GetLatestVaultVersion(ctx)

			if k.VaultsTransferCompleted(ctx, sourceVersion) {
				continue
			}

			sourceBtcVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_BTC, sourceVersion).Address
			sourceRunesVault := k.GetVaultByAssetTypeAndVersion(ctx, types.AssetType_ASSET_TYPE_RUNES, sourceVersion).Address

			// transfer runes
			if !k.VaultTransferCompleted(ctx, sourceRunesVault) {
				if err := k.TransferVault(ctx, sourceVersion, destVersion, types.AssetType_ASSET_TYPE_RUNES, nil, req.TargetUtxoNum, req.FeeRate); err != nil {
					k.Logger(ctx).Error("transfer vault errored", "source version", sourceVersion, "destination version", destVersion, "asset type", types.AssetType_ASSET_TYPE_RUNES, "target utxo num", req.TargetUtxoNum, "fee rate", req.FeeRate, "err", err)
					continue
				}
			}

			// transfer btc only when runes transfer completed
			if k.VaultTransferCompleted(ctx, sourceRunesVault) && !k.VaultTransferCompleted(ctx, sourceBtcVault) {
				if err := k.TransferVault(ctx, sourceVersion, destVersion, types.AssetType_ASSET_TYPE_BTC, nil, req.TargetUtxoNum, req.FeeRate); err != nil {
					k.Logger(ctx).Error("transfer vault errored", "source version", sourceVersion, "destination version", destVersion, "asset type", types.AssetType_ASSET_TYPE_BTC, "target utxo num", req.TargetUtxoNum, "fee rate", req.FeeRate, "err", err)
					continue
				}
			}

			// reenable bridge functions if disabled when all asset transfer completed
			if k.VaultsTransferCompleted(ctx, sourceVersion) {
				if req.DisableBridge {
					k.EnableBridge(ctx)
				}
			}
		}
	}
}
