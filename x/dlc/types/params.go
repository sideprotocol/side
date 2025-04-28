package types

import (
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
)

var (
	// default nonce queue size for price events
	DefaultPriceEventNonceQueueSize = uint32(20)

	// BTCUSD price pair
	BTCUSDPricePair = "BTCUSD"

	// default price interval for BTCUSD
	DefaultBTCUSDPriceInterval = sdkmath.LegacyNewDec(100)

	// default nonce queue size for date events
	DefaultDateEventNonceQueueSize = uint32(180)

	// default date interval
	DefaultDateInterval = 24 * time.Hour // 1 day

	// default nonce queue size for lending events
	DefaultLendingEventNonceQueueSize = uint32(1000)

	// default oracle participant base number
	DefaultOracleParticipantBaseNum = uint32(50)

	// maximum oracle participant base number
	MaxOracleParticipantBaseNum = uint32(100)

	// default oracle participant number
	DefaultOracleParticipantNum = uint32(21)

	// minimum oracle participant number
	MinOracleParticipantNum = uint32(3)

	// default nonce generation batch size
	DefaultNonceGenerationBatchSize = uint32(200)
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		PriceEventNonceQueueSize: DefaultPriceEventNonceQueueSize,
		PriceIntervals: []PriceInterval{
			{
				PricePair: BTCUSDPricePair,
				Interval:  DefaultBTCUSDPriceInterval,
			},
		},
		DateEventNonceQueueSize:    DefaultDateEventNonceQueueSize,
		DateInterval:               DefaultDateInterval,
		LendingEventNonceQueueSize: DefaultLendingEventNonceQueueSize,
		OracleParticipantBaseNum:   DefaultOracleParticipantBaseNum,
		OracleParticipantNum:       DefaultOracleParticipantNum,
		NonceGenerationBatchSize:   DefaultNonceGenerationBatchSize,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates params
func (p Params) Validate() error {
	if p.PriceEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "price event nonce queue size must be greater than 0")
	}

	for _, pi := range p.PriceIntervals {
		if err := validatePriceInterval(pi); err != nil {
			return err
		}
	}

	if p.DateEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "date event nonce queue size must be greater than 0")
	}

	if p.DateInterval <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "date interval must be greater than 0")
	}

	if p.LendingEventNonceQueueSize == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "lending event nonce queue size must be greater than 0")
	}

	if p.OracleParticipantBaseNum > MaxOracleParticipantBaseNum {
		return errorsmod.Wrapf(ErrInvalidParams, "oracle participant base number can not be greater than %d", MaxOracleParticipantBaseNum)
	}

	if p.OracleParticipantNum < MinOracleParticipantNum || p.OracleParticipantNum > p.OracleParticipantBaseNum {
		return errorsmod.Wrapf(ErrInvalidParams, "oracle participant number must be between [%d, %d]", MinOracleParticipantNum, p.OracleParticipantBaseNum)
	}

	if p.NonceGenerationBatchSize < 2 {
		return errorsmod.Wrapf(ErrInvalidParams, "nonce generation batch size can not be less than 2")
	}

	return nil
}

// validatePriceInterval validates the given price interval
func validatePriceInterval(priceInterval PriceInterval) error {
	if len(priceInterval.PricePair) == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "empty price pair")
	}

	if priceInterval.PricePair != strings.ToUpper(priceInterval.PricePair) {
		return errorsmod.Wrap(ErrInvalidParams, "price pair must be in uppercase")
	}

	if !priceInterval.Interval.IsPositive() {
		return errorsmod.Wrap(ErrInvalidParams, "invalid price interval")
	}

	return nil
}
