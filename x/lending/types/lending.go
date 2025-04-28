package types

import (
	"encoding/hex"
	fmt "fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/crypto/adaptor"
)

var (
	// OneYear represents the seconds in one year
	OneYear = 365 * 24 * 3600

	// initial borrow index
	InitialBorrowIndex = sdkmath.LegacyOneDec()
)

// GetExchangeRate calculates the sToken exchange rate according to the given params
// Formula:
// exchange rate = (totalAvailable + total borrowed) / totalSTokens
func GetExchangeRate(totalAvailable sdkmath.Int, totalBorrowed sdkmath.Int, totalSTokens sdkmath.Int) sdkmath.LegacyDec {
	if totalSTokens.IsZero() {
		return sdkmath.LegacyOneDec()
	}

	return sdkmath.LegacyNewDecFromInt(totalAvailable.Add(totalBorrowed)).Quo(totalSTokens.ToLegacyDec())
}

// GetInterest calculates the loan interest based on the given borrow index
func GetInterest(borrowAmount sdkmath.Int, startBorrowIndex sdkmath.LegacyDec, borrowIndex sdkmath.LegacyDec) sdkmath.Int {
	return borrowAmount.ToLegacyDec().Mul(borrowIndex).Quo(startBorrowIndex).TruncateInt().Sub(borrowAmount)
}

// GetTotalInterest calculates the total loan interest based on the given params
func GetTotalInterest(borrowAmount sdkmath.Int, maturity int64, borrowAPR uint32, blocksPerYear uint64) sdkmath.Int {
	totalBlocks := uint64(maturity) * blocksPerYear / uint64(OneYear)

	borrowRatePerBlock := sdkmath.LegacyNewDec(int64(borrowAPR)).Quo(sdkmath.LegacyNewDec(1000)).Quo(sdkmath.LegacyNewDec(int64(blocksPerYear)))
	borrowIndexRatio := sdkmath.LegacyOneDec().Add(borrowRatePerBlock)

	return borrowAmount.ToLegacyDec().Mul(borrowIndexRatio.Power(totalBlocks)).TruncateInt().Sub(borrowAmount)
}

// GetProtocolFee calculates the protocol fee based on the given interest and reserve factor
func GetProtocolFee(interest sdkmath.Int, reserveFactor uint32) sdkmath.Int {
	return interest.Mul(sdkmath.NewInt(int64(reserveFactor))).Quo(Permille)
}

// GetLiquidationPrice calculates the liquidation price according to the liquidation LTV
// Formula:
// liquidation price = (borrow amount + interest) / lltv / collateral amount
func GetLiquidationPrice(collateralAmount sdkmath.Int, borrowAmount sdkmath.Int, maturity int64, borrowAPR uint32, blocksPerYear uint64, lltv uint32) sdkmath.Int {
	interest := GetTotalInterest(borrowAmount, maturity, borrowAPR, blocksPerYear)
	liquidationPrice := borrowAmount.Add(interest).Mul(sdkmath.NewInt(100000000)).Mul(Percent).Quo(sdkmath.NewInt(int64(lltv))).Quo(collateralAmount).Quo(sdkmath.NewInt(1000000))

	// price precision
	precision := sdkmath.NewInt(100)

	return liquidationPrice.Quo(precision).Mul(precision)
}

// GetMaturityTime gets the actual maturity time according to the given maturity time
func GetMaturityTime(originMaturityTime int64) int64 {
	if originMaturityTime%(24*int64(time.Hour)) == 0 {
		return originMaturityTime
	}

	return time.Unix(originMaturityTime, 0).Truncate(24 * time.Hour).Add(24 * time.Hour).Unix()
}

// AdaptorPointFromSecret gets the corresponding adaptor point from the given secret
func AdaptorPointFromSecret(secret []byte) string {
	return hex.EncodeToString(adaptor.SecretToPubKey(secret))
}

// GetPricePair gets the price pair from the given pool config
func GetPricePair(poolConfig PoolConfig) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(poolConfig.CollateralAsset.PriceSymbol), strings.ToUpper(poolConfig.LendingAsset.PriceSymbol))
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

// HasRequestFee returns true if the request fee set in the given pool, false otherwise
func HasRequestFee(pool *LendingPool) bool {
	return pool.Config.RequestFee.IsPositive()
}

// HasOriginationFee returns true if the origination fee set in the given pool, false otherwise
func HasOriginationFee(pool *LendingPool) bool {
	return pool.Config.OriginationFee.IsPositive()
}

// HasReferralFee returns true if the referrer exists and the referral fee factor is not 0, false otherwise
func HasReferralFee(loan *Loan, pool *LendingPool) bool {
	return len(loan.Referrer) != 0 && pool.Config.ReferralFeeFactor > 0
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
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount can not be less than min borrow amount")
	}

	if HasMaxBorrowAmountLimit(pool) && borrowAmount.GT(pool.Config.MaxBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidAmount, "borrow amount can not be greater than max borrow amount")
	}

	return nil
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

// NewTranches initializes the pool tranches from the given tranche configs
func NewTranches(trancheConfigs []PoolTrancheConfig) []PoolTranche {
	tranches := make([]PoolTranche, len(trancheConfigs))

	for i, config := range trancheConfigs {
		tranches[i].Maturity = config.Maturity
		tranches[i].BorrowIndex = InitialBorrowIndex
	}

	return tranches
}

// ValidatePoolConfig validates the given pool config
func ValidatePoolConfig(config PoolConfig) error {
	if err := validateAssetMetadata(config.CollateralAsset); err != nil {
		return err
	}

	if err := validateAssetMetadata(config.LendingAsset); err != nil {
		return err
	}

	if config.SupplyCap.IsNil() || config.SupplyCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "supply cap can not be nil or negative")
	}

	if config.BorrowCap.IsNil() || config.BorrowCap.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow cap can not be nil or negative")
	}

	if config.MinBorrowAmount.IsNil() || config.MinBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "min borrow amount can not be nil or negative")
	}

	if config.MaxBorrowAmount.IsNil() || config.MaxBorrowAmount.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount can not be nil or negative")
	}

	if config.MinBorrowAmount.IsPositive() && config.MaxBorrowAmount.IsPositive() && config.MaxBorrowAmount.LT(config.MinBorrowAmount) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "max borrow amount can not be less than min borrow amount")
	}

	if err := validatePoolTranches(config.Tranches); err != nil {
		return err
	}

	if !config.RequestFee.IsValid() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid request fee")
	}

	if config.OriginationFee.IsNil() || config.OriginationFee.IsNegative() {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee can not be nil or negative")
	}

	if config.OriginationFee.IsPositive() && (!config.MinBorrowAmount.IsPositive() || config.OriginationFee.GTE(config.MinBorrowAmount)) {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "origination fee must be less than min borrow amount")
	}

	if config.ReserveFactor >= 1000 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid reserve factor")
	}

	if config.ReferralFeeFactor > 1000 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid referral fee factor")
	}

	if config.LiquidationThreshold == 0 || config.LiquidationThreshold >= 100 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid liquidation threshold")
	}

	if config.MaxLtv == 0 || config.MaxLtv >= 100 || config.MaxLtv >= config.LiquidationThreshold {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "invalid max ltv")
	}

	return nil
}

func validateAssetMetadata(metadata AssetMetadata) error {
	if err := sdk.ValidateDenom(metadata.Denom); err != nil {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset denom")
	}

	if len(metadata.Symbol) == 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset symbol")
	}

	if len(metadata.PriceSymbol) == 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset price symbol")
	}

	if metadata.Decimals < 0 {
		return errorsmod.Wrapf(ErrInvalidPoolConfig, "invalid asset decimals")
	}

	return nil
}

// validatePoolTrancheConfig validates the given tranche config
func validatePoolTranches(tranches []PoolTrancheConfig) error {
	if len(tranches) == 0 {
		return errorsmod.Wrap(ErrInvalidPoolConfig, "tranches can not be empty")
	}

	for _, tranche := range tranches {
		if tranche.Maturity <= 0 {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "maturity must be greater than 0")
		}

		if tranche.BorrowAPR == 0 || tranche.BorrowAPR >= 1000 {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "borrow apr must be between (0, 1000)")
		}

		if tranche.MinMaturityFactor == 0 || tranche.MinMaturityFactor > 1000 {
			return errorsmod.Wrap(ErrInvalidPoolConfig, "min maturity factor must be between (0, 1000]")
		}
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
