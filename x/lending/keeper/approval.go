package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// HandleApproval performs the loan approval
func (k Keeper) HandleApproval(ctx sdk.Context, sender string, depositTxHash string, loan *types.Loan) error {
	pool := k.GetPool(ctx, loan.PoolId)
	if pool.AvailableAmount.LT(loan.BorrowAmount.Amount) {
		return types.ErrInsufficientLiquidity
	}

	amount := sdk.NewCoin(loan.BorrowAmount.Denom, loan.BorrowAmount.Amount.Sub(loan.OriginationFee))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(loan.Borrower), sdk.NewCoins(amount)); err != nil {
		return err
	}

	if types.HasOriginationFee(pool) {
		originationFee := sdk.NewCoin(loan.BorrowAmount.Denom, loan.OriginationFee)
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(k.OriginationFeeCollector(ctx)), sdk.NewCoins(originationFee)); err != nil {
			return err
		}
	}

	// initiate signing request for repayment cet adaptor signatures from DCM
	if err := k.InitiateRepaymentCetSigningRequest(ctx, loan.VaultAddress); err != nil {
		return err
	}

	// update pool
	k.AfterPoolBorrowed(ctx, loan.PoolId, loan.BorrowAmount)

	loan.DisburseAt = ctx.BlockTime()
	loan.Status = types.LoanStatus_Open
	k.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApprove,
			sdk.NewAttribute(types.AttributeKeySender, sender),
			sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
			sdk.NewAttribute(types.AttributeKeyDepositTxHash, depositTxHash),
		),
	)

	return nil
}
