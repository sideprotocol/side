package types

import (
	"encoding/hex"
	fmt "fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
)

var (
	// denom prefix for yToken
	Y_TOKEN_DENOM_PREFIX = "y"

	// OneYear represents the seconds in one year
	OneYear = 365 * 24 * 3600

	// initial borrow index
	InitialBorrowIndex = sdkmath.LegacyOneDec()

	// price precision
	PricePrecision = "0.001"

	// maximum length of the referrer name
	MaxReferrerNameLength = 70

	// referral code regex: 8 alphanumeric characters
	ReferralCodeRegex = regexp.MustCompile("^[a-zA-Z0-9]{8}$")
)

// GetExchangeRate calculates the yToken exchange rate according to the given params
// Formula:
// exchange rate = (totalAvailable + total borrowed - total reserve) / totalYTokens
func GetExchangeRate(totalAvailable sdkmath.Int, totalBorrowed sdkmath.Int, totalReserve sdkmath.Int, totalYTokens sdkmath.Int) sdkmath.LegacyDec {
	if totalYTokens.IsZero() {
		return sdkmath.LegacyOneDec()
	}

	return sdkmath.LegacyNewDecFromInt(totalAvailable.Add(totalBorrowed).Sub(totalReserve)).Quo(totalYTokens.ToLegacyDec())
}

// GetInterest calculates the loan interest based on the given borrow index
func GetInterest(borrowAmount sdkmath.Int, startBorrowIndex sdkmath.LegacyDec, borrowIndex sdkmath.LegacyDec) sdkmath.Int {
	return borrowAmount.ToLegacyDec().Mul(borrowIndex).Quo(startBorrowIndex).TruncateInt().Sub(borrowAmount)
}

// GetTotalInterest calculates the total loan interest based on the given params
func GetTotalInterest(borrowAmount sdkmath.Int, maturity int64, borrowAPR sdkmath.LegacyDec, blocksPerYear uint64) sdkmath.Int {
	totalBlocks := uint64(maturity) * blocksPerYear / uint64(OneYear)

	borrowRatePerBlock := borrowAPR.Quo(sdkmath.LegacyNewDec(int64(blocksPerYear)))
	borrowIndexRatio := sdkmath.LegacyOneDec().Add(borrowRatePerBlock)

	return borrowAmount.ToLegacyDec().Mul(borrowIndexRatio.Power(totalBlocks)).TruncateInt().Sub(borrowAmount)
}

// GetProtocolFee calculates the protocol fee based on the given interest and reserve factor
func GetProtocolFee(interest sdkmath.Int, reserveFactor sdkmath.LegacyDec) sdkmath.Int {
	return interest.ToLegacyDec().Mul(reserveFactor).TruncateInt()
}

// GetLiquidationPrice calculates the liquidation price according to the liquidation LTV
// Formula:
// 1. collateral is the base price asset:
// liquidation price = (borrow amount + interest) / lltv / collateral amount
// 2. collateral is NOT the base price asset:
// liquidation price = collateral amount * lltv / (borrow amount + interest)
func GetLiquidationPrice(collateralAmount sdkmath.Int, collateralAssetDecimals int, borrowAmount sdkmath.Int, borrowAssetDecimals int, maturity int64, borrowAPR sdkmath.LegacyDec, blocksPerYear uint64, lltv sdkmath.LegacyDec, collateralIsBaseAsset bool) sdkmath.LegacyDec {
	interest := GetTotalInterest(borrowAmount, maturity, borrowAPR, blocksPerYear)

	var liquidationPrice sdkmath.LegacyDec
	if collateralIsBaseAsset {
		liquidationPrice = borrowAmount.Add(interest).Mul(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).ToLegacyDec().Quo(lltv).QuoInt(collateralAmount).QuoInt(sdkmath.NewIntWithDecimal(1, borrowAssetDecimals))
	} else {
		liquidationPrice = collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, borrowAssetDecimals)).ToLegacyDec().Mul(lltv).QuoInt(borrowAmount.Add(interest)).QuoInt(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals))
	}

	return NormalizePrice(liquidationPrice, collateralIsBaseAsset)
}

// ToBeLiquidated returns true if the given price satisfies the liquidation price, false otherwise
func ToBeLiquidated(price sdkmath.LegacyDec, liquidationPrice sdkmath.LegacyDec, collateralIsBaseAsset bool) bool {
	if collateralIsBaseAsset {
		return price.LTE(liquidationPrice)
	}

	return price.GTE(liquidationPrice)
}

// CheckLTV returns true if the collateral amount and borrow amount satisfy the max LTV limitation by the given price, false otherwise
func CheckLTV(collateralAmount sdkmath.Int, collateralAssetDecimals int, borrowAmount sdkmath.Int, borrowAssetDecimals int, maxLTV sdkmath.LegacyDec, price sdkmath.LegacyDec, collateralIsBaseAsset bool) bool {
	if collateralIsBaseAsset {
		return collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, borrowAssetDecimals)).ToLegacyDec().Mul(maxLTV).Mul(price).QuoInt(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).TruncateInt().GTE(borrowAmount)
	}

	return collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, borrowAssetDecimals)).ToLegacyDec().Mul(maxLTV).Quo(price).QuoInt(sdkmath.NewIntWithDecimal(1, collateralAssetDecimals)).TruncateInt().GTE(borrowAmount)
}

// GetPricePair gets the price pair from the given pool config
func GetPricePair(poolConfig PoolConfig) string {
	if poolConfig.CollateralAsset.IsBasePriceAsset {
		return fmt.Sprintf("%s%s", poolConfig.CollateralAsset.PriceSymbol, poolConfig.LendingAsset.PriceSymbol)
	}

	return fmt.Sprintf("%s%s", poolConfig.LendingAsset.PriceSymbol, poolConfig.CollateralAsset.PriceSymbol)
}

// FormatPrice formats the given price
// Assume that the given price is valid
func FormatPrice(price sdkmath.LegacyDec) string {
	decimalPrice, _ := decimal.NewFromString(price.String())

	return decimalPrice.String()
}

// FormatPrice formats the given price with the specified pair
// Assume that the given price is valid
func FormatPriceWithPair(price sdkmath.LegacyDec, pair string) string {
	return fmt.Sprintf("%s%s", FormatPrice(price), pair)
}

// NormalizePrice normalizes the given price
func NormalizePrice(price sdkmath.LegacyDec, collateralIsBaseAsset bool) sdkmath.LegacyDec {
	adjust := sdkmath.LegacyMustNewDecFromStr(PricePrecision)
	n := digitOrZeroCount(price)

	for range n.Int64() {
		adjust = adjust.MulInt64(10)
	}

	for range -n.Int64() {
		adjust = adjust.QuoInt64(10)
	}

	if !collateralIsBaseAsset {
		return price.Quo(adjust).TruncateDec().Mul(adjust)
	}

	return price.Add(adjust).Quo(adjust).TruncateDec().Mul(adjust)
}

// YTokenDenom returns the yToken denom from the given pool id
func YTokenDenom(poolId string) string {
	return fmt.Sprintf("%s%s", Y_TOKEN_DENOM_PREFIX, poolId)
}

// PoolIdFromYTokenDenom returns the pool id from the given yToken denom
func PoolIdFromYTokenDenom(denom string) string {
	return strings.TrimPrefix(denom, Y_TOKEN_DENOM_PREFIX)
}

// ToLiquidationAssetMeta converts the given asset metadata to the corresponding liquidation asset metadata
func ToLiquidationAssetMeta(metadata AssetMetadata) liquidationtypes.AssetMetadata {
	return liquidationtypes.AssetMetadata{
		Denom:            metadata.Denom,
		Symbol:           metadata.Symbol,
		Decimals:         metadata.Decimals,
		PriceSymbol:      metadata.PriceSymbol,
		IsBasePriceAsset: metadata.IsBasePriceAsset,
	}
}

// AdaptorPointFromSecret gets the corresponding adaptor point from the given secret
func AdaptorPointFromSecret(secret []byte) string {
	return hex.EncodeToString(adaptor.SecretToPubKey(secret))
}

// HasSupplyCap returns true if the supply cap set in the given pool, false otherwise
func HasSupplyCap(pool *LendingPool) bool {
	return pool.Config.SupplyCap.IsPositive()
}

// HasBorrowCap returns true if the borrow cap set in the given pool, false otherwise
func HasBorrowCap(pool *LendingPool) bool {
	return pool.Config.BorrowCap.IsPositive()
}

// HasMinBorrowAmountLimit returns true if the min borrow amount set in the given pool, false otherwise
func HasMinBorrowAmountLimit(pool *LendingPool) bool {
	return pool.Config.MinBorrowAmount.IsPositive()
}

// HasMaxBorrowAmountLimit returns true if the max borrow amount set in the given pool, false otherwise
func HasMaxBorrowAmountLimit(pool *LendingPool) bool {
	return pool.Config.MaxBorrowAmount.IsPositive()
}

// HasReferralFee returns true if the referral fee exists, false otherwise
func HasReferralFee(loan *Loan) bool {
	return loan.Referrer != nil && loan.Referrer.ReferralFeeFactor.IsPositive()
}

// CheckSupplyCap checks if the supply cap will be exceeded for the given deposit amount
func CheckSupplyCap(pool *LendingPool, depositAmount sdkmath.Int) error {
	if HasSupplyCap(pool) && pool.Supply.Amount.Add(depositAmount).GT(pool.Config.SupplyCap) {
		return ErrSupplyCapExceeded
	}

	return nil
}

// CheckBorrowCap checks if the borrow cap will be exceeded for the given borrow amount
func CheckBorrowCap(pool *LendingPool, borrowAmount sdkmath.Int) error {
	if HasBorrowCap(pool) && pool.BorrowedAmount.Add(borrowAmount).GT(pool.Config.BorrowCap) {
		return ErrBorrowCapExceeded
	}

	return nil
}

// CheckBorrowAmountLimit checks if the borrow amount satisfies limits for the given pool
func CheckBorrowAmountLimit(pool *LendingPool, borrowAmount sdkmath.Int) error {
	if HasMinBorrowAmountLimit(pool) && borrowAmount.LT(pool.Config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount cannot be less than min borrow amount")
	}

	if HasMaxBorrowAmountLimit(pool) && borrowAmount.GT(pool.Config.MaxBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount cannot be greater than max borrow amount")
	}

	return nil
}

// LoanDisbursed returns true if the loan has been disbursed, false otherwise
func LoanDisbursed(loan *Loan) bool {
	return loan.Status >= LoanStatus_Open
}

// CollateralRedeemable returns true if the collateral is redeemable, false otherwise
func CollateralRedeemable(loan *Loan) bool {
	if len(loan.Authorizations) > 0 {
		return loan.Status == LoanStatus_Rejected
	}

	return loan.Status == LoanStatus_Requested || loan.Status == LoanStatus_Cancelled || loan.Status == LoanStatus_Rejected
}

// GetTrancheConfig gets the corresponding tranche config according to the given maturity
func GetTrancheConfig(tranches []PoolTrancheConfig, maturity int64) (*PoolTrancheConfig, bool) {
	for _, tranche := range tranches {
		if tranche.Maturity == maturity {
			return &tranche, true
		}
	}

	return nil, false
}

// GetTranche gets the corresponding tranche according to the given maturity
func GetTranche(tranches []PoolTranche, maturity int64) (*PoolTranche, bool) {
	for _, tranche := range tranches {
		if tranche.Maturity == maturity {
			return &tranche, true
		}
	}

	return nil, false
}

// NewTranche creates a new tranche from the given tranche config
func NewTranche(config PoolTrancheConfig) PoolTranche {
	return PoolTranche{
		Maturity:    config.Maturity,
		BorrowIndex: InitialBorrowIndex,
	}
}

// NewTranches initializes the pool tranches from the given tranche configs
func NewTranches(trancheConfigs []PoolTrancheConfig) []PoolTranche {
	tranches := make([]PoolTranche, len(trancheConfigs))

	for i, config := range trancheConfigs {
		tranches[i] = NewTranche(config)
	}

	return tranches
}

// ValidatePoolConfig validates the given pool config
func ValidatePoolConfig(config PoolConfig) error {
	if err := validateAssetsMetadata(config.CollateralAsset, config.LendingAsset); err != nil {
		return err
	}

	if config.SupplyCap.IsNil() || config.SupplyCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "supply cap cannot be nil or negative")
	}

	if config.BorrowCap.IsNil() || config.BorrowCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow cap cannot be nil or negative")
	}

	if config.MinBorrowAmount.IsNil() || config.MinBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "min borrow amount cannot be nil or negative")
	}

	if config.MaxBorrowAmount.IsNil() || config.MaxBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount cannot be nil or negative")
	}

	if config.MinBorrowAmount.IsPositive() && config.MaxBorrowAmount.IsPositive() && config.MaxBorrowAmount.LT(config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount cannot be less than min borrow amount")
	}

	if err := validatePoolTranches(config.Tranches); err != nil {
		return err
	}

	if !config.RequestFee.IsValid() || !config.RequestFee.IsPositive() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid request fee")
	}

	if config.OriginationFeeFactor.IsNegative() || config.OriginationFeeFactor.GTE(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid origination fee factor")
	}

	if config.ReserveFactor.IsNegative() || config.ReserveFactor.GTE(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid reserve factor")
	}

	if !config.LiquidationThreshold.IsPositive() || config.LiquidationThreshold.GTE(sdkmath.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid liquidation threshold")
	}

	if !config.MaxLtv.IsPositive() || config.MaxLtv.GTE(sdkmath.LegacyOneDec()) || config.MaxLtv.GTE(config.LiquidationThreshold) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid max ltv")
	}

	return nil
}

// ValidatePoolConfigUpdate validates the given pool config update
func ValidatePoolConfigUpdate(config PoolConfig, newConfig PoolConfig) error {
	if newConfig.CollateralAsset != config.CollateralAsset {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "collateral asset cannot be updated")
	}

	if newConfig.LendingAsset != config.LendingAsset {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "lending asset cannot be updated")
	}

	return nil
}

// validateAssetsMetadata validates the given assets metadata
func validateAssetsMetadata(collateralAsset AssetMetadata, lendingAsset AssetMetadata) error {
	if err := validateAssetMetadata(collateralAsset); err != nil {
		return err
	}

	if err := validateAssetMetadata(lendingAsset); err != nil {
		return err
	}

	if collateralAsset.IsBasePriceAsset == lendingAsset.IsBasePriceAsset {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "conflicting base price asset")
	}

	return nil
}

// validateAssetMetadata validates the given asset metadata
func validateAssetMetadata(metadata AssetMetadata) error {
	if err := sdk.ValidateDenom(metadata.Denom); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset denom")
	}

	if len(metadata.Symbol) == 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset symbol")
	}

	if metadata.Decimals < 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset decimals")
	}

	if len(metadata.PriceSymbol) == 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset price symbol")
	}

	return nil
}

// validatePoolTrancheConfig validates the given tranche config
func validatePoolTranches(tranches []PoolTrancheConfig) error {
	if len(tranches) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "tranches cannot be empty")
	}

	trancheMap := make(map[int64]bool)

	for _, tranche := range tranches {
		if tranche.Maturity <= 0 {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "maturity must be greater than 0")
		}

		if trancheMap[tranche.Maturity] {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "duplicate maturity")
		}

		if !tranche.BorrowAPR.IsPositive() || tranche.BorrowAPR.GTE(sdkmath.LegacyOneDec()) {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow apr must be between (0, 1)")
		}

		trancheMap[tranche.Maturity] = true
	}

	return nil
}

// ToScopedId converts the given local id to the scoped id
func ToScopedId(id uint64) string {
	return fmt.Sprintf("%d", id)
}

// FromScopedId converts the scoped id to the local id
// Assume that the scoped id is valid
func FromScopedId(scopedId string) uint64 {
	id, _ := strconv.ParseUint(scopedId, 10, 64)
	return id
}
