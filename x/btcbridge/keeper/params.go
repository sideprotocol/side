package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// DepositConfirmationDepth gets the confirmation depth for deposit transactions
func (k Keeper) DepositConfirmationDepth(ctx sdk.Context) int32 {
	return k.GetParams(ctx).DepositConfirmationDepth
}

// WithdrawConfirmationDepth gets the confirmation depth for withdrawal transactions
func (k Keeper) WithdrawConfirmationDepth(ctx sdk.Context) int32 {
	return k.GetParams(ctx).WithdrawConfirmationDepth
}

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

// ProtocolDepositFee gets the protocol fee for deposit
func (k Keeper) ProtocolDepositFee(ctx sdk.Context) int64 {
	return k.GetParams(ctx).ProtocolFees.DepositFee
}

// ProtocolWithdrawFee gets the protocol fee for withdrawal
func (k Keeper) ProtocolWithdrawFee(ctx sdk.Context) int64 {
	return k.GetParams(ctx).ProtocolFees.WithdrawFee
}

// MinBTCDeposit gets the minimum deposit amount for BTC
func (k Keeper) MinBTCDeposit(ctx sdk.Context) int64 {
	return k.GetParams(ctx).ProtocolLimits.BtcMinDeposit
}

// MinBTCWithdraw gets the minimum withdrawal amount for BTC
func (k Keeper) MinBTCWithdraw(ctx sdk.Context) int64 {
	return k.GetParams(ctx).ProtocolLimits.BtcMinWithdraw
}

// MaxBTCWithdraw gets the maximum withdrawal amount for BTC
func (k Keeper) MaxBTCWithdraw(ctx sdk.Context) int64 {
	return k.GetParams(ctx).ProtocolLimits.BtcMaxWithdraw
}

// BtcDenom gets the btc denomination
func (k Keeper) BtcDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).BtcVoucherDenom
}

// MaxBtcBatchWithdrawNum gets the maximum btc batch withdrawal number
func (k Keeper) MaxBtcBatchWithdrawNum(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).WithdrawParams.MaxBtcBatchWithdrawNum
}

// RateLimitParamsSet returns true if the rate limit params are set, false otherwise
func (k Keeper) RateLimitParamsSet(ctx sdk.Context) bool {
	// check if the global rate limit period is set (or address rate limit period)
	return k.GetParams(ctx).RateLimitParams.GlobalRateLimitParams.Period > 0
}

// GlobalRateLimitPeriod gets the period of the global rate limit
func (k Keeper) GlobalRateLimitPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).RateLimitParams.GlobalRateLimitParams.Period
}

// GlobalRateLimitSupplyPercentageQuota gets the supply percentage quota of the global rate limit
func (k Keeper) GlobalRateLimitSupplyPercentageQuota(ctx sdk.Context) uint32 {
	return k.GetParams(ctx).RateLimitParams.GlobalRateLimitParams.SupplyPercentageQuota
}

// AddressRateLimitPeriod gets the period of the per address rate limit
func (k Keeper) AddressRateLimitPeriod(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).RateLimitParams.AddressRateLimitParams.Period
}

// AddressRateLimitQuota gets the quota of the per address rate limit
func (k Keeper) AddressRateLimitQuota(ctx sdk.Context) int64 {
	return k.GetParams(ctx).RateLimitParams.AddressRateLimitParams.Quota
}

// IBCPortId gets the IBC port id
func (k Keeper) IBCPortId(ctx sdk.Context) string {
	return k.GetParams(ctx).IbcParams.PortId
}

// IBCTimeoutHeightOffset gets the IBC timeout height offset
func (k Keeper) IBCTimeoutHeightOffset(ctx sdk.Context) uint64 {
	return k.GetParams(ctx).IbcParams.TimeoutHeightOffset
}

// IBCTimeoutDuration gets the IBC timeout duration
func (k Keeper) IBCTimeoutDuration(ctx sdk.Context) time.Duration {
	return k.GetParams(ctx).IbcParams.TimeoutDuration
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

// IsTrustedFeeProvider returns true if the given address is a trusted fee provider, false otherwise
func (k Keeper) IsTrustedFeeProvider(ctx sdk.Context, addr string) bool {
	for _, provider := range k.GetParams(ctx).TrustedFeeProviders {
		if provider == addr {
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
