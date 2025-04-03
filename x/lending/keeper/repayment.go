package keeper

import (
	"bytes"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil/psbt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// InitiateRepaymentCetSigningRequest initiates the signing request for the repayment cet
// Assume that both the loan and repayment cet exist
func (k Keeper) InitiateRepaymentCetSigningRequest(ctx sdk.Context, loanId string) error {
	adaptorPoint, err := k.GetRepaymentCetAdaptorPoint(ctx, loanId)
	if err != nil {
		return err
	}

	sigHashes, err := types.GetRepaymentCetSigHashes(k.GetDLCMeta(ctx, loanId))
	if err != nil {
		return err
	}

	k.tssKeeper.InitiateSigningRequest(
		ctx,
		types.ModuleName,
		loanId,
		tsstypes.SigningType_SIGNING_TYPE_SCHNORR_ADAPTOR,
		int32(types.SigningIntent_SIGNING_INTENT_REPAYMENT),
		k.GetLoan(ctx, loanId).DCM,
		sigHashes,
		&tsstypes.SigningOptions{AdaptorPoint: hex.EncodeToString(adaptorPoint)},
	)

	return nil
}

// HandleRepaymentAdaptorSignatures handles repayment adaptor signatures
func (k Keeper) HandleRepaymentAdaptorSignatures(ctx sdk.Context, loanId string, adaptorSignatures []string) error {
	if !k.HasLoan(ctx, loanId) {
		return types.ErrLoanDoesNotExist
	}

	loan := k.GetLoan(ctx, loanId)
	if loan.Status != types.LoanStatus_Open && loan.Status != types.LoanStatus_Repaid {
		return errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan neither open nor repaid")
	}

	dlcMeta := k.GetDLCMeta(ctx, loanId)

	repaymentCet := dlcMeta.RepaymentCet
	if len(repaymentCet.DCMAdaptorSignatures) != 0 {
		return types.ErrRepaymentAdaptorSigsAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet.Tx)), true)
	if len(adaptorSignatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidAdaptorSignatures, "mismatched adaptor signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	adaptorPoint, _ := k.GetRepaymentCetAdaptorPoint(ctx, loanId)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		adaptorSigBytes, _ := hex.DecodeString(adaptorSignatures[i])

		if !adaptor.Verify(adaptorSigBytes, sigHash, dcmPubKey, adaptorPoint) {
			return types.ErrInvalidAdaptorSignature
		}
	}

	dlcMeta.RepaymentCet.DCMAdaptorSignatures = adaptorSignatures
	k.SetDLCMeta(ctx, loanId, dlcMeta)

	return nil
}

// CompleteRepayment completes the repayment of the given loan
func (k Keeper) CompleteRepayment(ctx sdk.Context, loan *types.Loan) error {
	pool := k.GetPool(ctx, loan.PoolId)
	repayment := k.GetRepayment(ctx, loan.VaultAddress)

	interest := repayment.Amount.Sub(loan.BorrowAmount)
	protocolFee := sdk.NewCoin(loan.BorrowAmount.Denom, interest.Amount.Mul(sdkmath.NewInt(int64(pool.Config.ReserveFactor))).Quo(types.Permille))

	amount := repayment.Amount.Sub(protocolFee)
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RepaymentEscrowAccount, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RepaymentEscrowAccount, sdk.MustAccAddressFromBech32(k.ProtocolFeeCollector(ctx)), sdk.NewCoins(protocolFee)); err != nil {
		return err
	}

	// update pool
	k.AfterPoolRepaid(ctx, loan.PoolId, loan.BorrowAmount, interest.Amount, protocolFee.Amount)

	loan.Status = types.LoanStatus_Closed
	k.SetLoan(ctx, loan)

	return nil
}

// GetRepaymentCetAdaptorPoint gets the adaptor point of the repayment cet
// Assume that the loan exists
func (k Keeper) GetRepaymentCetAdaptorPoint(ctx sdk.Context, loanId string) ([]byte, error) {
	loan := k.GetLoan(ctx, loanId)
	repaymentEvent := k.dlcKeeper.GetEvent(ctx, loan.RepaymentEventId)

	return dlctypes.GetSignaturePointFromEvent(repaymentEvent, 0)
}
