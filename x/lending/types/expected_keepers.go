package types

import (
	"context"

	"github.com/btcsuite/btcd/btcutil"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI

	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
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

// LiquidationKeeper defines the expected liquidation keeper interface
type LiquidationKeeper interface {
	GetLiquidation(ctx sdk.Context, id uint64) *liquidationtypes.Liquidation

	CreateLiquidation(ctx sdk.Context, liquidation *liquidationtypes.Liquidation) *liquidationtypes.Liquidation

	SetPrice(ctx sdk.Context, pair string, price string)

	SetLiquidatedDebtHandler(handler liquidationtypes.LiquidatedDebtHandler)
}

// DLCKeeper defines the expected DLC keeper interface
type DLCKeeper interface {
	HasEvent(ctx sdk.Context, id uint64) bool
	GetEvent(ctx sdk.Context, id uint64) *dlctypes.DLCEvent
	HasEventByPrice(ctx sdk.Context, price sdkmath.Int) bool
	GetEventByPrice(ctx sdk.Context, price sdkmath.Int) *dlctypes.DLCEvent
	HasEventByDate(ctx sdk.Context, date int64) bool
	GetEventByDate(ctx sdk.Context, date int64) *dlctypes.DLCEvent
	GetAvailableLendingEvent(ctx sdk.Context) *dlctypes.DLCEvent

	GetAttestationByEvent(ctx sdk.Context, eventId uint64) *dlctypes.DLCAttestation

	HasDCM(ctx sdk.Context, id uint64) bool
	GetDCM(ctx sdk.Context, id uint64) *dlctypes.DCM

	SetEvent(ctx sdk.Context, event *dlctypes.DLCEvent)
	TriggerDLCEvent(ctx sdk.Context, id uint64, outcomeIndex int)

	SetPrice(ctx sdk.Context, pair string, price string)
}

// BtcBridgeKeeper defines the expected BtcBridge keeper interface
type BtcBridgeKeeper interface {
	ValidateTransaction(ctx sdk.Context, tx string, prevTx string, blockHash string, proof []string) (*btcutil.Tx, *btcutil.Tx, error)
	GetFeeRate(ctx sdk.Context) *btcbridgetypes.FeeRate
}
