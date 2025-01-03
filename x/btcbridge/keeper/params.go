package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// DepositEnabled returns true if deposit enabled, false otherwise
func (k Keeper) DepositEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).DepositEnabled
}

// WithdrawEnabled returns true if withdrawal enabled, false otherwise
func (k Keeper) WithdrawEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).WithdrawEnabled
}

// ProtocolDepositFeeEnabled returns true if the protocol fee is required for deposit, false otherwise
func (k Keeper) ProtocolDepositFeeEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).ProtocolFees.DepositFee > 0
}

// ProtocolWithdrawFeeEnabled returns true if the protocol fee is required for withdrawal, false otherwise
func (k Keeper) ProtocolWithdrawFeeEnabled(ctx sdk.Context) bool {
	return k.GetParams(ctx).ProtocolFees.WithdrawFee > 0
}

// ProtocolFeeCollector gets the protocol fee collector
func (k Keeper) ProtocolFeeCollector(ctx sdk.Context) string {
	return k.GetParams(ctx).ProtocolFees.Collector
}

// BtcDenom gets the btc denomination
func (k Keeper) BtcDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).BtcVoucherDenom
}

// IsTrustedBtcRelayer returns true if the given address is a trusted btc relayer, false otherwise
func (k Keeper) IsTrustedBtcRelayer(ctx sdk.Context, addr string) bool {
	trustedBtcRelayers := k.GetParams(ctx).TrustedBtcRelayers
	if len(trustedBtcRelayers) == 0 {
		return true
	}

	for _, relayer := range trustedBtcRelayers {
		if relayer == addr {
			return true
		}
	}

	return false
}

// IsTrustedNonBtcRelayer returns true if the given address is a trusted non-btc relayer, false otherwise
func (k Keeper) IsTrustedNonBtcRelayer(ctx sdk.Context, addr string) bool {
	for _, relayer := range k.GetParams(ctx).TrustedNonBtcRelayers {
		if relayer == addr {
			return true
		}
	}

	return false
}

// IsTrustedOracle returns true if the given address is a trusted oracle, false otherwise
func (k Keeper) IsTrustedOracle(ctx sdk.Context, addr string) bool {
	for _, oracle := range k.GetParams(ctx).TrustedOracles {
		if oracle == addr {
			return true
		}
	}

	return false
}

// GetVaultByAssetTypeAndVersion gets the vault by the given asset type and version
func (k Keeper) GetVaultByAssetTypeAndVersion(ctx sdk.Context, assetType types.AssetType, version uint64) *types.Vault {
	for _, v := range k.GetParams(ctx).Vaults {
		if v.AssetType == assetType && v.Version == version {
			return v
		}
	}

	return nil
}

// GetVaultVersionByAddress gets the vault version of the given address
func (k Keeper) GetVaultVersionByAddress(ctx sdk.Context, address string) (uint64, bool) {
	for _, v := range k.GetParams(ctx).Vaults {
		if v.Address == address {
			return v.Version, true
		}
	}

	return 0, false
}

// GetMaxUtxoNum gets the maximum utxo number for the signing request
func (k Keeper) GetMaxUtxoNum(ctx sdk.Context) int {
	params := k.GetParams(ctx)

	return int(params.WithdrawParams.MaxUtxoNum)
}
