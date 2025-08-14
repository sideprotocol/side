package ante

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BtcBridgeKeeper defines the expected btcbridge keeper interface
type BtcBridgeKeeper interface {
	BtcDenom(ctx sdk.Context) string

	FeeSponsorshipEnabled(ctx sdk.Context) bool
	MaxSponsorFee(ctx sdk.Context) sdk.Coins
}

// FeeGrantKeeper defines the expected feegrant keeper.
type FeeGrantKeeper interface {
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}
