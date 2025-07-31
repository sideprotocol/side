package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// HandleApproval performs the loan approval
func (k Keeper) HandleApproval(ctx sdk.Context, loan *types.Loan) error {
	pool := k.GetPool(ctx, loan.PoolId)
	if pool.AvailableAmount.LT(loan.BorrowAmount.Amount) {
		return types.ErrInsufficientLiquidity
	}

	// initiate signing request for repayment cet adaptor signatures from DCM
	if err := k.InitiateRepaymentCetSigningRequest(ctx, loan.VaultAddress); err != nil {
		return err
	}

	amount := sdk.NewCoin(loan.BorrowAmount.Denom, loan.BorrowAmount.Amount.Sub(loan.OriginationFee))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(loan.Borrower), sdk.NewCoins(amount)); err != nil {
		// unexpected error
		return types.ErrUnexpected
	}

	if loan.OriginationFee.IsPositive() {
		if err := k.handleOriginationFee(ctx, loan); err != nil {
			// unexpected error
			return types.ErrUnexpected
		}
	}

	// update pool
	k.AfterPoolBorrowed(ctx, loan.PoolId, loan.Maturity, loan.BorrowAmount)

	// update starting borrow index
	tranche, _ := types.GetTranche(pool.Tranches, loan.Maturity)
	loan.StartBorrowIndex = tranche.BorrowIndex

	// update total interest and protocol fee
	loan.Interest = types.GetTotalInterest(loan.BorrowAmount.Amount, loan.MaturityTime-ctx.BlockTime().Unix(), loan.BorrowAPR, k.GetBlocksPerYear(ctx))
	loan.ProtocolFee = types.GetProtocolFee(loan.Interest, pool.Config.ReserveFactor)

	loan.DisburseAt = ctx.BlockTime()
	loan.Status = types.LoanStatus_Open
	k.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApprove,
			sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return nil
}

// handleOriginationFee handles the origination fee for the given loan
func (k Keeper) handleOriginationFee(ctx sdk.Context, loan *types.Loan) error {
	originationFee := sdk.NewCoin(loan.BorrowAmount.Denom, loan.OriginationFee)
	referralFee := sdk.NewCoin(loan.BorrowAmount.Denom, sdkmath.ZeroInt())

	if loan.Referrer != nil {
		referralFee.Amount = originationFee.Amount.ToLegacyDec().Mul(loan.Referrer.ReferralFeeFactor).TruncateInt()
		originationFee = originationFee.Sub(referralFee)
	}

	if originationFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(k.OriginationFeeCollector(ctx)), sdk.NewCoins(originationFee)); err != nil {
			return err
		}
	}

	if referralFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(loan.Referrer.Address), sdk.NewCoins(referralFee)); err != nil {
			return err
		}
	}

	return nil
}
