package keeper

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
)

// HandleLiquidationSignatures handles the liquidation signatures
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

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return types.ErrInvalidSignature
		}
	}

	dlcMeta.LiquidationCet.DCMSignatures = signatures
	k.SetDLCMeta(ctx, loanId, dlcMeta)

	return nil
}

// handleDefaultLiquidationSignatures handles the default liquidation signatures
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

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.DefaultLiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return types.ErrInvalidSignature
		}
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
	protocolFee := types.GetProtocolFee(interest, pool.Config.ReserveFactor)

	referralFee := sdkmath.ZeroInt()
	actualProtocolFee := protocolFee
	if protocolFee.IsPositive() && types.HasReferralFee(loan, pool) {
		referralFee = protocolFee.Mul(sdkmath.NewInt(int64(pool.Config.ReferralFeeFactor))).Quo(types.Permille)
		actualProtocolFee = protocolFee.Sub(referralFee)
	}

	if debtAmount.Amount.LTE(interest) {
		// TODO
		return nil
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
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(loan.Referrer), sdk.NewCoins(sdk.NewCoin(debtAmount.Denom, referralFee))); err != nil {
			return err
		}
	}

	k.AfterPoolRepaid(ctx, loan.PoolId, loan.Maturity, debtAmount.SubAmount(interest), interest, protocolFee, actualProtocolFee)

	k.DeductLiquidationAccruedInterest(ctx, loan)

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
