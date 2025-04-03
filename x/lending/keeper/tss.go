package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, signingRequestId uint64, loanId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	switch types.SigningIntent(intent) {
	case types.SigningIntent_SIGNING_INTENT_REPAYMENT:
		return k.HandleRepaymentAdaptorSignatures(ctx, loanId, signatures)

	case types.SigningIntent_SIGNING_INTENT_LIQUIDATION:
		return k.HandleLiquidationSignatures(ctx, loanId, signatures)

	case types.SigningIntent_SIGNING_INTENT_DEFAULT_LIQUIDATION:
		return k.handleDefaultLiquidationSignatures(ctx, loanId, signatures)

	case types.SigningIntent_SIGNING_INTENT_CANCELLATION:
		return k.HandleCancellationSignatures(ctx, loanId, signatures)

	default:
		return types.ErrInvalidSigningIntent
	}
}
