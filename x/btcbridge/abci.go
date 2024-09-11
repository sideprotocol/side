package btcbridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/keeper"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

// EndBlocker called at every block to handle DKG requests
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
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

		// transfer vaults if the EnableTransfer flag set
		if req.EnableTransfer {
			err := transferVaults(ctx, k, req.TargetUtxoNum, req.FeeRate)

			// reenable bridge when successfully transferred
			if err == nil && req.DisableBridge {
				k.EnableBridge(ctx)
			}
		}
	}
}

// transferVaults performs the vault asset transfer (possibly partially)
func transferVaults(ctx sdk.Context, k keeper.Keeper, targetUtxoNum uint32, feeRate string) error {
	latestVaultVersion := k.GetLatestVaultVersion(ctx)

	if err := k.TransferVault(ctx, latestVaultVersion-1, latestVaultVersion, types.AssetType_ASSET_TYPE_RUNES, nil, targetUtxoNum, feeRate); err != nil {
		k.Logger(ctx).Error("transfer vault errored", "source version", latestVaultVersion-1, "destination version", latestVaultVersion, "asset type", types.AssetType_ASSET_TYPE_RUNES, "target utxo num", targetUtxoNum, "fee rate", feeRate, "err", err)

		return err
	}

	if err := k.TransferVault(ctx, latestVaultVersion-1, latestVaultVersion, types.AssetType_ASSET_TYPE_BTC, nil, targetUtxoNum, feeRate); err != nil {
		k.Logger(ctx).Error("transfer vault errored", "source version", latestVaultVersion-1, "destination version", latestVaultVersion, "asset type", types.AssetType_ASSET_TYPE_BTC, "target utxo num", targetUtxoNum, "fee rate", feeRate, "err", err)

		return err
	}

	return nil
}
