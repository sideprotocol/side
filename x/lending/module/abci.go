package lending

import (
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/adaptor"
	"github.com/sideprotocol/side/x/lending/keeper"
	"github.com/sideprotocol/side/x/lending/types"
	liquidationtypes "github.com/sideprotocol/side/x/liquidation/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// BeginBlocker called at the beginning of each block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	updatePools(ctx, k)
}

// EndBlocker called at the end of each block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	handlePendingLoans(ctx, k)
	handleActiveLoans(ctx, k)

	handleLiquidatedLoans(ctx, k)
	handleRepayments(ctx, k)
}

// handleActiveLoans handles pending loans
func handlePendingLoans(ctx sdk.Context, k keeper.Keeper) {
	// handler on loan rejected
	rejectHandler := func(loan *types.Loan, authorizationId uint64, reason error) {
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

	// get all pending loans
	loans := k.GetPendingLoans(ctx)

	for _, loan := range loans {
		pool := k.GetPool(ctx, loan.PoolId)
		authorizationId := k.GetAuthorizationId(ctx, loan.VaultAddress)

		// check if the maturity time already reached
		if ctx.BlockTime().Unix() >= loan.MaturityTime {
			rejectHandler(loan, authorizationId, types.ErrMaturityTimeReached)
			continue
		}

		if authorizationId > 0 {
			// get the current price
			currentPrice, err := k.GetPrice(ctx, types.GetPricePair(pool.Config))
			if err != nil {
				continue
			}

			// check if liquidation price reached
			if types.ToBeLiquidated(currentPrice, loan.LiquidationPrice, pool.Config.CollateralAsset.IsBasePriceAsset) {
				rejectHandler(loan, authorizationId, types.ErrLiquidationPriceReached)
				continue
			}

			// try to approve loan if all deposit txs verified
			if k.DepositsVerified(ctx, k.GetAuthorization(ctx, loan.VaultAddress, authorizationId)) {
				// check LTV
				if !types.CheckLTV(loan.CollateralAmount, int(pool.Config.CollateralAsset.Decimals), loan.BorrowAmount.Amount, int(pool.Config.LendingAsset.Decimals), pool.Config.MaxLtv, currentPrice, pool.Config.CollateralAsset.IsBasePriceAsset) {
					rejectHandler(loan, authorizationId, types.ErrInsufficientCollateral)
					continue
				}

				// check if the borrow cap already reached
				if err := types.CheckBorrowCap(pool, loan.BorrowAmount.Amount); err != nil {
					rejectHandler(loan, authorizationId, err)
					continue
				}

				// approve loan
				if err := k.HandleApproval(ctx, loan); err != nil {
					rejectHandler(loan, authorizationId, err)
					continue
				}

				// set the authorization status
				loan := k.GetLoan(ctx, loan.VaultAddress)
				loan.Authorizations[authorizationId-1].Status = types.AuthorizationStatus_AUTHORIZATION_STATUS_AUTHORIZED
				k.SetLoan(ctx, loan)
			}
		}
	}
}

// handleActiveLoans handles active loans
func handleActiveLoans(ctx sdk.Context, k keeper.Keeper) {
	// get all active loans
	loans := k.GetLoans(ctx, types.LoanStatus_Open)

	for _, loan := range loans {
		var liquidationCet string
		var sigHashes []string
		var signingIntent int32

		var outcomeIndex int

		var liquidationInterest sdkmath.Int

		pool := k.GetPool(ctx, loan.PoolId)
		pricePair := types.GetPricePair(pool.Config)

		currentPrice, err := k.GetPrice(ctx, pricePair)
		if err != nil {
			k.Logger(ctx).Info("failed to get price", "pair", pricePair, "err", err)
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

			// trigger dlc event if not triggered yet
			if !k.DLCKeeper().GetEvent(ctx, loan.DlcEventId).HasTriggered {
				k.DLCKeeper().TriggerDLCEvent(ctx, loan.DlcEventId, outcomeIndex)
			}

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
}

// handleLiquidatedLoans handles liquidated loans
func handleLiquidatedLoans(ctx sdk.Context, k keeper.Keeper) {
	// get liquidated loans
	loans := k.GetLiquidatedLoans(ctx)

	for _, loan := range loans {
		// get dlc meta
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)

		// get liquidation cet and type
		cet, cetType := types.GetLiquidationCetAndType(dlcMeta, loan.Status)

		// check if the borrower adapted signatures already exist
		if len(cet.BorrowerAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.DlcEventId)
			if attestation == nil {
				continue
			}

			eventSignature, _ := hex.DecodeString(attestation.Signature)
			adaptorSecret := eventSignature[32:]

			// decrypt the adaptor signatures
			for _, adaptorSignature := range cet.BorrowerAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				cet.BorrowerAdaptedSignatures = append(
					cet.BorrowerAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed liquidation cet if both borrower adapted signatures(obviously exist) and DCM signatures already exist
		if len(cet.DCMSignatures) != 0 {
			signedTx, txHash, err := types.BuildSignedCet(cet.Tx, loan.BorrowerAuthPubKey, cet.BorrowerAdaptedSignatures, loan.DCM, cet.DCMSignatures, cetType)
			if err != nil {
				k.Logger(ctx).Info("failed to build signed liquidation cet", "loan id", loan.VaultAddress, "err", err)
			} else {
				cet.SignedTxHex = hex.EncodeToString(signedTx)

				// remove from the liquidation queue
				k.RemoveFromLiquidationQueue(ctx, loan.VaultAddress)

				// emit event
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(types.EventTypeGenerateSignedCet,
						sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
						sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", cetType)),
						sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
					),
				)
			}
		}

		// update liquidation cet
		types.UpdateLiquidationCet(dlcMeta, cetType, cet)

		// update dlc meta
		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}

// handleRepayments handles repayments
func handleRepayments(ctx sdk.Context, k keeper.Keeper) {
	// get all repaid loans
	loans := k.GetLoans(ctx, types.LoanStatus_Repaid)

	for _, loan := range loans {
		// trigger dlc event if not triggered yet
		if !k.DLCKeeper().GetEvent(ctx, loan.DlcEventId).HasTriggered {
			k.DLCKeeper().TriggerDLCEvent(ctx, loan.DlcEventId, types.RepaidOutcomeIndex)
			continue
		}

		// check if the repayment cet has been signed
		dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
		if len(dlcMeta.RepaymentCet.SignedTxHex) != 0 {
			continue
		}

		// check if the DCM adaptor signatures have been submitted
		if len(dlcMeta.RepaymentCet.DCMAdaptorSignatures) == 0 {
			continue
		}

		if len(dlcMeta.RepaymentCet.DCMAdaptedSignatures) == 0 {
			// check if the event attestation has been submitted
			attestation := k.DLCKeeper().GetAttestationByEvent(ctx, loan.DlcEventId)
			if attestation == nil {
				continue
			}

			eventSignature, _ := hex.DecodeString(attestation.Signature)
			adaptorSecret := eventSignature[32:]

			// decrypt the DCM adaptor signatures
			for _, adaptorSignature := range dlcMeta.RepaymentCet.DCMAdaptorSignatures {
				adaptorSignature, _ := hex.DecodeString(adaptorSignature)
				adaptedSignature := adaptor.Adapt(adaptorSignature, adaptorSecret)

				// update the adapted signatures
				dlcMeta.RepaymentCet.DCMAdaptedSignatures = append(
					dlcMeta.RepaymentCet.DCMAdaptedSignatures,
					hex.EncodeToString(adaptedSignature))
			}
		}

		// build signed repayment cet
		signedTx, txHash, err := types.BuildSignedCet(dlcMeta.RepaymentCet.Tx, loan.BorrowerPubKey, dlcMeta.RepaymentCet.BorrowerSignatures, loan.DCM, dlcMeta.RepaymentCet.DCMAdaptedSignatures, types.CetType_REPAYMENT)
		if err != nil {
			k.Logger(ctx).Info("failed to build signed repayment cet", "loan id", loan.VaultAddress, "err", err)
		} else {
			dlcMeta.RepaymentCet.SignedTxHex = hex.EncodeToString(signedTx)

			if err := k.CompleteRepayment(ctx, loan); err != nil {
				k.Logger(ctx).Info("failed to complete repayment", "loan id", loan.VaultAddress, "err", err)
			}

			// emit event
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(types.EventTypeGenerateSignedCet,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyCetType, fmt.Sprintf("%d", types.CetType_REPAYMENT)),
					sdk.NewAttribute(types.AttributeKeyTxHash, txHash.String()),
				),
			)
		}

		k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)
	}
}

// updatePools updates all active pools at the beginning of each block
func updatePools(ctx sdk.Context, k keeper.Keeper) {
	// get all active pools
	pools := k.GetPools(ctx, types.PoolStatus_ACTIVE)

	for _, pool := range pools {
		k.UpdatePool(ctx, pool)

		k.SetPool(ctx, pool)
	}
}
