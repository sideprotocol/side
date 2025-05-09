package btcbridge

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/keeper"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

// EndBlocker called at every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleIBCWithdrawRequests(ctx, k)
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

	// get the pending btc withdrawal requests
	pendingWithdrawRequests := k.GetPendingBtcWithdrawRequests(ctx, p.WithdrawParams.MaxBtcBatchWithdrawNum)
	if len(pendingWithdrawRequests) == 0 {
		return
	}

	feeRate := k.GetFeeRate(ctx)
	if err := k.CheckFeeRate(ctx, feeRate); err != nil {
		k.Logger(ctx).Info("invalid fee rate", "value", feeRate.Value, "height", feeRate.Height)
		return
	}

	vault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_BTC)
	if vault == nil {
		k.Logger(ctx).Info("btc vault does not exist")
		return
	}

	signingRequest, err := k.BuildBtcBatchWithdrawSigningRequest(ctx, pendingWithdrawRequests, feeRate.Value, vault.Address)
	if err != nil {
		k.Logger(ctx).Info("failed to build signing request", "err", err)
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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateSigning,
			sdk.NewAttribute(types.AttributeKeyId, signingRequest.Txid),
			sdk.NewAttribute(types.AttributeKeySigners, strings.Join(types.GetSigners(signingRequest.Psbt), types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(types.GetSigHashes(signingRequest.Psbt), types.AttributeValueSeparator)),
		),
	)
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
				if err := k.TransferVault(ctx, sourceVersion, destVersion, types.AssetType_ASSET_TYPE_RUNES, nil, req.TargetUtxoNum); err != nil {
					k.Logger(ctx).Info("failed to transfer vault", "source version", sourceVersion, "destination version", destVersion, "asset type", types.AssetType_ASSET_TYPE_RUNES, "target utxo num", req.TargetUtxoNum, "err", err)
					continue
				}
			}

			// transfer btc only when runes transfer completed
			if k.VaultTransferCompleted(ctx, sourceRunesVault) && !k.VaultTransferCompleted(ctx, sourceBtcVault) {
				if err := k.TransferVault(ctx, sourceVersion, destVersion, types.AssetType_ASSET_TYPE_BTC, nil, req.TargetUtxoNum); err != nil {
					k.Logger(ctx).Info("failed to transfer vault", "source version", sourceVersion, "destination version", destVersion, "asset type", types.AssetType_ASSET_TYPE_BTC, "target utxo num", req.TargetUtxoNum, "err", err)
					continue
				}
			}

			if k.VaultsTransferCompleted(ctx, sourceVersion) {
				k.Logger(ctx).Info("vaults transfer completed", "source version", sourceVersion, "destination version", destVersion)
			}
		}
	}
}

// handleIBCWithdrawRequests handles BTC withdrawal requests via IBC
func handleIBCWithdrawRequests(ctx sdk.Context, k keeper.Keeper) {
	// get the pending IBC withdrawal requests
	pendingIBCWithdrawRequests := k.GetPendingIBCWithdrawRequests(ctx, k.MaxBtcBatchWithdrawNum(ctx))
	if len(pendingIBCWithdrawRequests) == 0 {
		return
	}

	// handle the IBC withdrawal request
	for _, req := range pendingIBCWithdrawRequests {
		var err error

		address := sdk.MustAccAddressFromBech32(req.Address)
		amount, _ := sdk.ParseCoinNormalized(req.Amount)

		if k.ProtocolWithdrawFeeEnabled(ctx) {
			// deduct the protocol fee and get the actual withdrawal amount
			amount, err = k.HandleWithdrawProtocolFee(ctx, address, amount)
			if err != nil {
				k.Logger(ctx).Info("failed to handle protocol fee for ibc withdrawal", "address", address, "amount", amount, "err", err)
				continue
			}
		}

		withdrawRequest, err := k.HandleWithdrawal(ctx, req.Address, amount)
		if err != nil {
			k.Logger(ctx).Info("failed to handle ibc withdrawal", "address", address, "amount", amount, "err", err)
			continue
		}

		// remove from queue
		k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)

		// Emit events
		k.EmitEvent(ctx, req.Address,
			sdk.NewAttribute("amount", amount.String()),
			sdk.NewAttribute("sequence", fmt.Sprintf("%d", withdrawRequest.Sequence)),
			sdk.NewAttribute("txid", withdrawRequest.Txid),
			sdk.NewAttribute("channel_id", req.ChannelId),
			sdk.NewAttribute("channel_sequence", fmt.Sprintf("%d", req.Sequence)),
		)
	}
}
