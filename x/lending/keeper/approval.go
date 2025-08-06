package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// HandleApproval performs the loan approval or rejection
func (k Keeper) HandleApproval(ctx sdk.Context, loan *types.Loan) error {
	pool := k.GetPool(ctx, loan.PoolId)
	authorizationId := k.GetAuthorizationId(ctx, loan.VaultAddress)

	// check if the maturity time already reached
	if ctx.BlockTime().Unix() >= loan.MaturityTime {
		k.reject(ctx, loan, authorizationId, types.ErrMaturityTimeReached)
		return nil
	}

	if authorizationId > 0 {
		// get the current price
		currentPrice, err := k.GetPrice(ctx, types.GetPricePair(pool.Config))
		if err != nil {
			return nil
		}

		// check if liquidation price reached
		if types.ToBeLiquidated(currentPrice, loan.LiquidationPrice, pool.Config.CollateralAsset.IsBasePriceAsset) {
			k.reject(ctx, loan, authorizationId, types.ErrLiquidationPriceReached)
			return nil
		}

		// try to approve loan if the repayment cet signed by DCM and all deposit txs verified
		if k.RepaymentCetSigned(ctx, loan.VaultAddress) && k.DepositsVerified(ctx, k.GetAuthorization(ctx, loan.VaultAddress, authorizationId)) {
			// check LTV
			if !types.CheckLTV(loan.CollateralAmount, int(pool.Config.CollateralAsset.Decimals), loan.BorrowAmount.Amount, int(pool.Config.LendingAsset.Decimals), pool.Config.MaxLtv, currentPrice, pool.Config.CollateralAsset.IsBasePriceAsset) {
				k.reject(ctx, loan, authorizationId, types.ErrInsufficientCollateral)
				return nil
			}

			// check if the borrow cap already reached
			if err := types.CheckBorrowCap(pool, loan.BorrowAmount.Amount); err != nil {
				k.reject(ctx, loan, authorizationId, err)
				return nil
			}

			// approve loan
			if err := k.approve(ctx, loan); err != nil {
				if errors.Is(err, types.ErrUnexpected) {
					return err
				}

				k.reject(ctx, loan, authorizationId, err)
				return nil
			}

			// set the authorization status
			loan := k.GetLoan(ctx, loan.VaultAddress)
			loan.Authorizations[authorizationId-1].Status = types.AuthorizationStatus_AUTHORIZATION_STATUS_AUTHORIZED
			k.SetLoan(ctx, loan)
		}
	}

	return nil
}

// approve performs the loan disbursement
func (k Keeper) approve(ctx sdk.Context, loan *types.Loan) error {
	pool := k.GetPool(ctx, loan.PoolId)
	if pool.AvailableAmount.LT(loan.BorrowAmount.Amount) {
		return types.ErrInsufficientLiquidity
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

// reject performs the loan rejection
func (k Keeper) reject(ctx sdk.Context, loan *types.Loan, authorizationId uint64, reason error) {
	if authorizationId > 0 {
		loan.Authorizations[authorizationId-1].Status = types.AuthorizationStatus_AUTHORIZATION_STATUS_REJECTED
	}

	loan.Status = types.LoanStatus_Rejected
	k.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReject,
			sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyAuthorizationId, fmt.Sprintf("%d", authorizationId)),
			sdk.NewAttribute(types.AttributeKeyReason, reason.Error()),
		),
	)
}

// handleOriginationFee handles the origination fee for the given loan
func (k Keeper) handleOriginationFee(ctx sdk.Context, loan *types.Loan) error {
	originationFee := sdk.NewCoin(loan.BorrowAmount.Denom, loan.OriginationFee)
	referralFee := sdk.NewCoin(loan.BorrowAmount.Denom, sdkmath.ZeroInt())

	if types.HasReferralFee(loan) {
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

		// emit referral event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReferral,
				sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
				sdk.NewAttribute(types.AttributeKeyReferralCode, loan.Referrer.ReferralCode),
				sdk.NewAttribute(types.AttributeKeyReferrerAddress, loan.Referrer.Address),
				sdk.NewAttribute(types.AttributeKeyReferralFeeFactor, loan.Referrer.ReferralFeeFactor.String()),
				sdk.NewAttribute(types.AttributeKeyReferralFee, referralFee.String()),
			),
		)
	}

	return nil
}
