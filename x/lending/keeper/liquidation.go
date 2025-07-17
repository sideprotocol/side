package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// HandleLiquidationSignatures handles the liquidation signatures
// Assume that signatures have already been verified
func (k Keeper) HandleLiquidationSignatures(ctx sdk.Context, loanId string, signatures []string) error {
	if !k.HasLoan(ctx, loanId) {
		return types.ErrLoanDoesNotExist
	}

	loan := k.GetLoan(ctx, loanId)
	if loan.Status != types.LoanStatus_Liquidated {
		return errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not liquidated")
	}

	dlcMeta := k.GetDLCMeta(ctx, loanId)
	if len(dlcMeta.LiquidationCet.DCMSignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	dlcMeta.LiquidationCet.DCMSignatures = signatures
	k.SetDLCMeta(ctx, loanId, dlcMeta)

	return nil
}

// handleDefaultLiquidationSignatures handles the default liquidation signatures
// Assume that signatures have already been verified
func (k Keeper) handleDefaultLiquidationSignatures(ctx sdk.Context, loanId string, signatures []string) error {
	if !k.HasLoan(ctx, loanId) {
		return types.ErrLoanDoesNotExist
	}

	loan := k.GetLoan(ctx, loanId)
	if loan.Status != types.LoanStatus_Defaulted {
		return errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not defaulted")
	}

	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
	if len(dlcMeta.DefaultLiquidationCet.DCMSignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	dlcMeta.DefaultLiquidationCet.DCMSignatures = signatures
	k.SetDLCMeta(ctx, loanId, dlcMeta)

	return nil
}

// HandleLiquidatedDebt handles the liquidated debt for the liquidated loan
func (k Keeper) HandleLiquidatedDebt(ctx sdk.Context, liquidationId uint64, loanId string, moduleAccount string, debtAmount sdk.Coin) error {
	loan := k.GetLoan(ctx, loanId)
	pool := k.GetPool(ctx, loan.PoolId)

	interest := k.GetCurrentInterest(ctx, loan).Amount

	principal := sdk.NewCoin(debtAmount.Denom, sdkmath.ZeroInt())
	if debtAmount.Amount.GT(interest) {
		// split debt to principal and interest
		principal = debtAmount.SubAmount(interest)
	} else {
		// consider debt as interest
		interest = debtAmount.Amount
	}

	protocolFee := types.GetProtocolFee(interest, pool.Config.ReserveFactor)

	referralFee := sdkmath.ZeroInt()
	actualProtocolFee := protocolFee
	if protocolFee.IsPositive() && loan.Referrer != nil {
		referralFee = protocolFee.ToLegacyDec().Mul(loan.Referrer.ReferralFeeFactor).TruncateInt()
		actualProtocolFee = protocolFee.Sub(referralFee)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, moduleAccount, types.ModuleName, sdk.NewCoins(debtAmount.SubAmount(protocolFee))); err != nil {
		return err
	}

	if actualProtocolFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(k.ProtocolFeeCollector(ctx)), sdk.NewCoins(sdk.NewCoin(debtAmount.Denom, actualProtocolFee))); err != nil {
			return err
		}
	}

	if referralFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(loan.Referrer.Address), sdk.NewCoins(sdk.NewCoin(debtAmount.Denom, referralFee))); err != nil {
			return err
		}
	}

	k.AfterPoolRepaid(ctx, loan.PoolId, loan.Maturity, principal, interest, protocolFee, actualProtocolFee)

	k.DeductLiquidationAccruedInterest(ctx, loan)

	// emit referral event
	if referralFee.IsPositive() {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReferral,
				sdk.NewAttribute(types.AttributeKeyLoanId, loanId),
				sdk.NewAttribute(types.AttributeKeyReferralCode, loan.Referrer.ReferralCode),
				sdk.NewAttribute(types.AttributeKeyReferrerAddress, loan.Referrer.Address),
				sdk.NewAttribute(types.AttributeKeyReferralFeeFactor, loan.Referrer.ReferralFeeFactor.String()),
				sdk.NewAttribute(types.AttributeKeyReferralFee, referralFee.String()),
			),
		)
	}

	return nil
}

// DeductLiquidationAccruedInterest deducts the interest accrued during the loan liquidation from total borrowed
func (k Keeper) DeductLiquidationAccruedInterest(ctx sdk.Context, loan *types.Loan) {
	interest := k.GetLiquidationAccruedInterest(ctx, loan)

	k.DecreaseTotalBorrowed(ctx, loan.PoolId, loan.Maturity, interest)
}

// GetLiquidationAccruedInterest gets the current accrued interest during the loan liquidation
func (k Keeper) GetLiquidationAccruedInterest(ctx sdk.Context, loan *types.Loan) sdkmath.Int {
	currentTotalInterest := types.GetInterest(loan.BorrowAmount.Amount, loan.StartBorrowIndex, k.GetCurrentBorrowIndex(ctx, loan))

	return currentTotalInterest.Sub(k.GetCurrentInterest(ctx, loan).Amount)
}
