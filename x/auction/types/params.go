package types

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	// default price drop period
	DefaultPriceDropPeriod = time.Duration(10) * time.Minute // 10min

	// default initial discount
	DefaultInitialDiscount = uint32(3) // 3%

	// default minimum amount for bid
	DefaultMinBidAmount = uint64(100000) // 100000sat

	// default liquidation bonus
	DefaultLiquidationBonus = uint32(50) // 5%

	// default protocol liquidation fee
	DefaultProtocolLiquidationFee = uint32(100) // 10%
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		PriceDropPeriod:                 DefaultPriceDropPeriod,
		InitialDiscount:                 DefaultInitialDiscount,
		MinBidAmount:                    DefaultMinBidAmount,
		LiquidationBonus:                DefaultLiquidationBonus,
		ProtocolLiquidationFee:          DefaultProtocolLiquidationFee,
		ProtocolLiquidationFeeCollector: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if p.PriceDropPeriod <= 0 {
		return errorsmod.Wrap(ErrInvalidParams, "price drop period must be greater than 0")
	}

	if p.InitialDiscount == 0 || p.InitialDiscount >= 100 {
		return errorsmod.Wrap(ErrInvalidParams, "initial discount must be between (0, 100)")
	}

	if p.MinBidAmount == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "minimum bid amount must be greater than 0")
	}

	if p.LiquidationBonus == 0 || p.LiquidationBonus >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid liquidation bonus")
	}

	if p.ProtocolLiquidationFee >= 1000 {
		return errorsmod.Wrap(ErrInvalidParams, "invalid protocol liquidation fee")
	}

	if _, err := sdk.AccAddressFromBech32(p.ProtocolLiquidationFeeCollector); err != nil {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol liquidation fee collector: %v", err)
	}

	return nil
}
