package types

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here

	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)

	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error

	HasSupply(ctx context.Context, denom string) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// OracleKeeper defines the expected oracle keeper interface
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, pair string) (sdkmath.LegacyDec, error)
}

// TSSKeeper defines the expected TSS keeper interface
type TSSKeeper interface {
	InitiateSigningRequest(ctx sdk.Context, module string, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, sigHashes []string, options *tsstypes.SigningOptions) *tsstypes.SigningRequest

	RegisterSigningRequestCompletedHandler(module string, handler tsstypes.SigningRequestCompletedHandler)
}

// BtcBridgeKeeper defines the expected BtcBridge keeper interface
type BtcBridgeKeeper interface {
	GetFeeRate(ctx sdk.Context) *btcbridgetypes.FeeRate
	CheckFeeRate(ctx sdk.Context, feeRate *btcbridgetypes.FeeRate) error
}

// LendingKeeper defines the expected lending keeper interface
type LendingKeeper interface {
	GetLiquidationAccruedInterest(ctx sdk.Context, loanId string) sdkmath.Int

	HandleLiquidatedDebt(ctx sdk.Context, liquidationId uint64, loanId string, moduleAccount string, debtAmount sdk.Coin, accruedInterestDuringLiquidation sdkmath.Int) error
}
