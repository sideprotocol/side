package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// HandleLiquidation performs possible liquidation for the given loan
func (k Keeper) HandleLiquidation(ctx sdk.Context, loan *types.Loan) {
	var liquidationCet string
	var sigHashes []string
	var signingIntent int32

	var outcomeIndex int

	var liquidationInterest sdkmath.Int

	pool := k.GetPool(ctx, loan.PoolId)
	pricePair := types.GetPricePair(pool.Config)

	currentPrice, err := k.GetPrice(ctx, pricePair)
	if err != nil {
		k.Logger(ctx).Warn("failed to get price", "pair", pricePair, "err", err)
	}

	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)

	// check if the loan has defaulted
	if ctx.BlockTime().Unix() >= loan.MaturityTime {
		liquidationInterest = loan.Interest
		loan.Status = types.LoanStatus_Defaulted

		liquidationCet = dlcMeta.DefaultLiquidationCet.Tx
		signingIntent = int32(types.SigningIntent_SIGNING_INTENT_DEFAULT_LIQUIDATION)
		outcomeIndex = types.DefaultLiquidatedOutcomeIndex

		// get default liquidation cet sig hashes; no error
		sigHashes, _ = types.GetCetSigHashes(dlcMeta, types.CetType_DEFAULT_LIQUIDATION)

		// emit default event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeDefault,
				sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			),
		)
	} else if !currentPrice.IsZero() {
		// check if the loan is to be liquidated
		if types.ToBeLiquidated(currentPrice, loan.LiquidationPrice, pool.Config.CollateralAsset.IsBasePriceAsset) {
			liquidationInterest = k.GetCurrentInterest(ctx, loan).Amount
			loan.Status = types.LoanStatus_Liquidated

			liquidationCet = dlcMeta.LiquidationCet.Tx
			signingIntent = int32(types.SigningIntent_SIGNING_INTENT_LIQUIDATION)
			outcomeIndex = types.LiquidatedOutcomeIndex

			// get liquidation cet sig hashes; no error
			sigHashes, _ = types.GetCetSigHashes(dlcMeta, types.CetType_LIQUIDATION)

			// emit liquidation event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeLiquidate,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
				),
			)
		}
	}

	// create liquidation if defaulted or liquidated
	if loan.Status == types.LoanStatus_Defaulted || loan.Status == types.LoanStatus_Liquidated {
		collateralDenom := pool.Config.CollateralAsset.Denom
		debtDenom := pool.Config.LendingAsset.Denom

		liquidation := k.LiquidationKeeper().CreateLiquidation(ctx, &liquidationtypes.Liquidation{
			LoanId:                     loan.VaultAddress,
			Debtor:                     loan.Borrower,
			DCM:                        loan.DCM,
			CollateralAmount:           sdk.NewCoin(collateralDenom, loan.CollateralAmount),
			ActualCollateralAmount:     sdk.NewCoin(collateralDenom, sdkmath.NewInt(types.GetLiquidationCetOutput(liquidationCet))),
			DebtAmount:                 sdk.NewCoin(debtDenom, loan.BorrowAmount.Amount.Add(liquidationInterest)),
			CollateralAsset:            types.ToLiquidationAssetMeta(pool.Config.CollateralAsset),
			DebtAsset:                  types.ToLiquidationAssetMeta(pool.Config.LendingAsset),
			LiquidationPrice:           currentPrice,
			LiquidationTime:            ctx.BlockTime(),
			LiquidatedCollateralAmount: sdk.NewCoin(collateralDenom, sdkmath.ZeroInt()),
			LiquidatedDebtAmount:       sdk.NewCoin(debtDenom, sdkmath.ZeroInt()),
			LiquidationBonusAmount:     sdk.NewCoin(collateralDenom, sdkmath.ZeroInt()),
			ProtocolLiquidationFee:     sdk.NewCoin(collateralDenom, sdkmath.ZeroInt()),
			LiquidationCet:             liquidationCet,
		})

		// update loan
		loan.LiquidationId = liquidation.Id
		k.SetLoan(ctx, loan)

		// add to liquidation queue
		k.AddToLiquidationQueue(ctx, loan.VaultAddress)

		// trigger dlc event
		k.DLCKeeper().TriggerDLCEvent(ctx, loan.DlcEventId, outcomeIndex)

		// initiate signing request
		k.TSSKeeper().InitiateSigningRequest(
			ctx,
			types.ModuleName,
			loan.VaultAddress,
			tsstypes.SigningType_SIGNING_TYPE_SCHNORR,
			signingIntent,
			loan.DCM,
			sigHashes,
			nil,
		)
	}
}

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
	if protocolFee.IsPositive() && types.HasReferralFee(loan) {
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

		// emit referral event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeReferral,
				sdk.NewAttribute(types.AttributeKeyLoanId, loanId),
				sdk.NewAttribute(types.AttributeKeyReferralCode, loan.Referrer.ReferralCode),
				sdk.NewAttribute(types.AttributeKeyReferrerAddress, loan.Referrer.Address),
				sdk.NewAttribute(types.AttributeKeyReferralFeeFactor, loan.Referrer.ReferralFeeFactor.String()),
				sdk.NewAttribute(types.AttributeKeyReferralFee, sdk.NewCoin(debtAmount.Denom, referralFee).String()),
			),
		)
	}

	k.AfterPoolRepaid(ctx, loan.PoolId, loan.Maturity, principal, interest, protocolFee, actualProtocolFee)

	k.DeductLiquidationAccruedInterest(ctx, loan, pool)

	return nil
}

// DeductLiquidationAccruedInterest deducts the interest accrued during the loan liquidation from the pool
func (k Keeper) DeductLiquidationAccruedInterest(ctx sdk.Context, loan *types.Loan, pool *types.LendingPool) {
	interest := k.GetLiquidationAccruedInterest(ctx, loan)
	protocolFee := types.GetProtocolFee(interest, pool.Config.ReserveFactor)

	k.RebalancePool(ctx, loan.PoolId, loan.Maturity, interest, protocolFee)
}

// GetLiquidationAccruedInterest gets the current accrued interest during the loan liquidation
func (k Keeper) GetLiquidationAccruedInterest(ctx sdk.Context, loan *types.Loan) sdkmath.Int {
	currentTotalInterest := types.GetInterest(loan.BorrowAmount.Amount, loan.StartBorrowIndex, k.GetCurrentBorrowIndex(ctx, loan))

	return currentTotalInterest.Sub(k.GetCurrentInterest(ctx, loan).Amount)
}
