package types

import (
	"context"

	"github.com/btcsuite/btcd/btcutil"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktype "github.com/cosmos/cosmos-sdk/x/bank/types"

	auctiontypes "github.com/sideprotocol/side/x/auction/types"
	btcbridgetypes "github.com/sideprotocol/side/x/btcbridge/types"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
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
	SetDenomMetaData(ctx context.Context, denomMetaData banktype.Metadata)

	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error

	HasSupply(ctx context.Context, denom string) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// OracleKeeper defines the expected oracle keeper interface
type OracleKeeper interface {
	GetPrice(ctx sdk.Context, pair string) (sdkmath.Int, error)
}

// AuctionKeeper defines the expected auction keeper interface
type AuctionKeeper interface {
	CreateAuction(ctx sdk.Context, auction *auctiontypes.Auction)
}

// DLCKeeper defines the expected DLC keeper interface
type DLCKeeper interface {
	HasEvent(ctx sdk.Context, id uint64) bool
	GetEvent(ctx sdk.Context, id uint64) *dlctypes.DLCPriceEvent
	GetEventByPrice(ctx sdk.Context, price sdkmath.Int) *dlctypes.DLCPriceEvent
	GetAttestationByEvent(ctx sdk.Context, eventId uint64) *dlctypes.DLCAttestation

	HasAgency(ctx sdk.Context, id uint64) bool
	GetAgency(ctx sdk.Context, id uint64) *dlctypes.Agency

	TriggerEvent(ctx sdk.Context, id uint64)

	SetPrice(ctx sdk.Context, pair string, price string)
}

// BtcBridgeKeeper defines the expected BtcBridge keeper interface
type BtcBridgeKeeper interface {
	ValidateTransaction(ctx sdk.Context, tx string, prevTx string, blockHash string, proof []string) (*btcutil.Tx, *btcutil.Tx, error)
	GetFeeRate(ctx sdk.Context) *btcbridgetypes.FeeRate
}
