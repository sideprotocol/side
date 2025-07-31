package btcbridge

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/keeper"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

// EndBlocker called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handleDKGRequests(ctx, k)
	handleRefreshingRequests(ctx, k)

	handleIBCWithdrawRequests(ctx, k)
	handleBtcWithdrawRequests(ctx, k)

	updateRateLimit(ctx, k)

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

	// get fee rate
	feeRate := k.GetFeeRate(ctx)
	if err := k.CheckFeeRate(ctx, feeRate); err != nil {
		return
	}

	protocolFee := sdk.NewInt64Coin(k.BtcDenom(ctx), k.ProtocolWithdrawFee(ctx))
	protocoFeeCollector := sdk.MustAccAddressFromBech32(k.ProtocolFeeCollector(ctx))

	// handle the IBC withdrawal request
	for _, req := range pendingIBCWithdrawRequests {
		address := sdk.MustAccAddressFromBech32(req.Address)
		amount, _ := sdk.ParseCoinNormalized(req.Amount)

		// check if the balance is sufficient
		if k.BankKeeper().SpendableCoin(ctx, address, k.BtcDenom(ctx)).IsLT(amount) {
			k.Logger(ctx).Warn("failed to perform withdrawal from IBC", "address", req.Address, "amount", req.Amount, "err", "insufficient balance")

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// deduct protocol fee
		withdrawAmount, err := amount.SafeSub(protocolFee)
		if err != nil || withdrawAmount.Amount.Int64() < k.MinBTCWithdraw(ctx) || withdrawAmount.Amount.Int64() > k.MaxBTCWithdraw(ctx) {
			k.Logger(ctx).Info("failed to perform withdrawal from IBC", "address", req.Address, "amount", req.Amount, "err", "invalid withdrawal amount")

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// estimate the btc network fee
		networkFee, err := k.EstimateWithdrawalNetworkFee(ctx, req.Address, withdrawAmount, feeRate.Value)
		if err != nil {
			k.Logger(ctx).Info("failed to estimate network fee for withdrawal from IBC", "address", req.Address, "amount", req.Amount, "fee rate", feeRate.Value, "err", err)

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// deduct network fee
		withdrawAmount, err = withdrawAmount.SafeSub(networkFee)
		if err != nil || withdrawAmount.Amount.Int64() < k.MinBTCWithdraw(ctx) {
			k.Logger(ctx).Info("failed to perform withdrawal from IBC", "address", req.Address, "amount", req.Amount, "fee rate", feeRate.Value, "network fee", networkFee, "err", "invalid withdrawal amount")

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// handle rate limit
		if err := k.HandleRateLimit(ctx, req.Address, amount); err != nil {
			k.Logger(ctx).Info("failed to perform withdrawal from IBC", "address", req.Address, "amount", req.Amount, "err", err)

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// burn asset
		if err := k.BurnAsset(ctx, req.Address, withdrawAmount.Add(networkFee)); err != nil {
			k.Logger(ctx).Warn("failed to burn asset for withdrawal from IBC", "address", req.Address, "amount", req.Amount, "burned amount", withdrawAmount.Add(networkFee), "err", err)

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// transfer protocol fee to fee collector
		if err := k.BankKeeper().SendCoins(ctx, address, protocoFeeCollector, sdk.NewCoins(protocolFee)); err != nil {
			k.Logger(ctx).Warn("failed to transfer protocol fee for withdrawal from IBC", "address", req.Address, "amount", req.Amount, "protocol fee", protocolFee, "err", err)

			k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)
			continue
		}

		// set the withdrawal request
		withdrawRequest := k.NewWithdrawRequest(ctx, req.Address, withdrawAmount.String())
		k.SetWithdrawRequest(ctx, withdrawRequest)

		// add to the pending queue
		k.AddToBtcWithdrawRequestQueue(ctx, withdrawRequest)

		// remove from queue
		k.RemoveFromIBCWithdrawRequestQueue(ctx, req.ChannelId, req.Sequence)

		// Emit events
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeIBCWithdraw,
				sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
				sdk.NewAttribute(types.AttributeKeyAmount, withdrawAmount.String()),
				sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", withdrawRequest.Sequence)),
				sdk.NewAttribute(types.AttributeKeyChannelId, req.ChannelId),
				sdk.NewAttribute(types.AttributeKeyPacketSequence, fmt.Sprintf("%d", req.Sequence)),
			),
		)
	}
}

// handleRefreshingRequests performs the key refreshing request handling
func handleRefreshingRequests(ctx sdk.Context, k keeper.Keeper) {
	// get pending refreshing requests
	requests := k.GetPendingRefreshingRequests(ctx)

	for _, req := range requests {
		// check if the refreshing request expired
		if !req.ExpirationTime.IsZero() && !ctx.BlockTime().Before(req.ExpirationTime) {
			req.Status = types.RefreshingStatus_REFRESHING_STATUS_TIMEDOUT
			k.SetRefreshingRequest(ctx, req)

			continue
		}

		// check refreshing completions
		completions := k.GetRefreshingCompletions(ctx, req.Id)
		if len(completions) != len(k.GetRefreshingParticipants(ctx, req)) {
			continue
		}

		// update status
		req.Status = types.RefreshingStatus_REFRESHING_STATUS_COMPLETED
		k.SetRefreshingRequest(ctx, req)

		// update DKG participants and threshold
		dkgRequest := k.GetDKGRequest(ctx, req.DkgId)
		dkgRequest.Participants = k.GetRefreshingParticipants(ctx, req)
		dkgRequest.Threshold = req.Threshold
		k.SetDKGRequest(ctx, dkgRequest)

		// Emit events
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeRefreshingCompleted,
				sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", req.Id)),
				sdk.NewAttribute(types.AttributeKeyDKGId, fmt.Sprintf("%d", req.DkgId)),
			),
		)
	}
}

// updateRateLimit updates the rate limit
func updateRateLimit(ctx sdk.Context, k keeper.Keeper) {
	if !k.HasRateLimit(ctx) {
		if k.RateLimitParamsSet(ctx) {
			// initialize the rate limit if the params are set
			k.SetRateLimit(ctx, k.NewRateLimit(ctx))
		}

		return
	}

	rateLimit := k.GetRateLimit(ctx)

	// if the current global rate limit epoch has ended, proceed to the next one
	if !ctx.BlockTime().Before(rateLimit.GlobalRateLimit.EndTime) {
		globalRateLimit := rateLimit.GlobalRateLimit
		newGlobalRateLimit := k.NewGlobalRateLimit(ctx)

		rateLimit.GlobalRateLimit = newGlobalRateLimit

		// emit events
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeGlobalRateLimitUpdated,
				sdk.NewAttribute(types.AttributeKeyPreviousStartTime, globalRateLimit.StartTime.String()),
				sdk.NewAttribute(types.AttributeKeyPreviousEndTime, globalRateLimit.EndTime.String()),
				sdk.NewAttribute(types.AttributeKeyPreviousQuota, fmt.Sprintf("%d", globalRateLimit.Quota)),
				sdk.NewAttribute(types.AttributeKeyPreviousUsed, fmt.Sprintf("%d", globalRateLimit.Used)),
				sdk.NewAttribute(types.AttributeKeyStartTime, newGlobalRateLimit.StartTime.String()),
				sdk.NewAttribute(types.AttributeKeyEndTime, newGlobalRateLimit.EndTime.String()),
				sdk.NewAttribute(types.AttributeKeyQuota, fmt.Sprintf("%d", newGlobalRateLimit.Quota)),
			),
		)
	}

	// if the current per address rate limit epoch has ended, proceed to the next one
	if !ctx.BlockTime().Before(rateLimit.AddressRateLimit.EndTime) {
		// remove the current address rate limit details
		k.RemoveAllAddressRateLimitDetails(ctx)

		rateLimit.AddressRateLimit = k.NewAddressRateLimit(ctx)
	}

	// update rate limit
	k.SetRateLimit(ctx, rateLimit)
}
