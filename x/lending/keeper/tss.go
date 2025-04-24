package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, id uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	switch types.SigningIntent(intent) {
	case types.SigningIntent_SIGNING_INTENT_REPAYMENT:
		return k.HandleRepaymentAdaptorSignatures(ctx, scopedId, signatures)

	case types.SigningIntent_SIGNING_INTENT_LIQUIDATION:
		return k.HandleLiquidationSignatures(ctx, scopedId, signatures)

	case types.SigningIntent_SIGNING_INTENT_DEFAULT_LIQUIDATION:
		return k.handleDefaultLiquidationSignatures(ctx, scopedId, signatures)

	case types.SigningIntent_SIGNING_INTENT_REDEMPTION:
		return k.HandleRedemptionSignatures(ctx, types.FromScopedId(scopedId), signatures)

	default:
		return types.ErrInvalidSigningIntent
	}
}
