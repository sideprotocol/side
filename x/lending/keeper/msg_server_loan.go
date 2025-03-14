package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	btcschnorr "github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/schnorr"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
)

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(goCtx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.dlcKeeper.HasAgency(ctx, msg.AgencyId) {
		return nil, types.ErrInvalidAgency
	}

	agency := m.dlcKeeper.GetAgency(ctx, msg.AgencyId)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, agency.Pubkey, msg.LoanSecretHash, msg.MaturityTime, msg.MaturityTime+m.FinalTimeoutDuration(ctx))
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	poolConfig := m.GetPool(ctx, msg.PoolId).Config

	interests := msg.BorrowAmount.Amount.Mul(math.NewInt(int64(poolConfig.BorrowRate))).Quo(types.Permille)
	fees := msg.BorrowAmount.Amount.Mul(math.NewInt(int64(poolConfig.BorrowRate)).Sub(math.NewInt(int64(poolConfig.SupplyRate)))).Quo(types.Permille)

	if msg.BorrowAmount.Amount.LTE(poolConfig.OriginationFee) {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "borrowed amount must be greater than origination fee")
	}

	loan := &types.Loan{
		VaultAddress:   vault,
		Borrower:       msg.Borrower,
		BorrowerPubKey: msg.BorrowerPubkey,
		Agency:         agency.Pubkey,
		HashLoanSecret: msg.LoanSecretHash,
		MaturityTime:   msg.MaturityTime,
		FinalTimeout:   msg.MaturityTime + m.FinalTimeoutDuration(ctx),
		BorrowAmount:   msg.BorrowAmount,
		Interests:      interests,
		Fees:           fees,
		PoolId:         msg.PoolId,
		CreateAt:       ctx.BlockTime(),
		Status:         types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeApply,
			sdk.NewAttribute(types.AttributeKeyVault, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyBorrower, loan.Borrower),
			sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
			sdk.NewAttribute(types.AttributeKeyLoanSecretHash, loan.HashLoanSecret),
			sdk.NewAttribute(types.AttributeKeyMuturityTime, fmt.Sprint(loan.MaturityTime)),
			sdk.NewAttribute(types.AttributeKeyFinalTimeout, fmt.Sprint(loan.FinalTimeout)),
			sdk.NewAttribute(types.AttributeKeyBorrowAmount, loan.BorrowAmount.String()),
			sdk.NewAttribute(types.AttributeKeyPoolId, loan.PoolId),
		))

	return &types.MsgApplyResponse{}, nil

}

// SubmitLiquidationCet implements types.MsgServer.
func (m msgServer) SubmitLiquidationCet(goCtx context.Context, msg *types.MsgSubmitLiquidationCet) (*types.MsgSubmitLiquidationCetResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	if !m.dlcKeeper.HasEvent(ctx, msg.EventId) {
		return nil, types.ErrInvalidPriceEvent
	}

	event := m.dlcKeeper.GetEvent(ctx, msg.EventId)
	if event.HasTriggered {
		return nil, errorsmod.Wrap(types.ErrInvalidPriceEvent, "event has triggered")
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	fundTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.DepositTx)), true)
	depositTxid := fundTx.UnsignedTx.TxHash().String()

	adaptorPoint, err := dlctypes.GetSignaturePointFromEvent(event)
	if err != nil {
		return nil, err
	}

	if err := types.VerifyLiquidationCET(fundTx, msg.LiquidationCet, loan.BorrowerPubKey, loan.Agency, msg.LiquidationAdaptorSignature, hex.EncodeToString(adaptorPoint)); err != nil {
		return nil, err
	}

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	dlcMeta, err := types.BuildDLCMeta(fundTx, vaultPkScript, msg.LiquidationCet, msg.LiquidationAdaptorSignature, loan.BorrowerPubKey, loan.Agency, loan.HashLoanSecret, loan.MaturityTime, loan.FinalTimeout)
	if err != nil {
		return nil, err
	}

	collateralAmount := math.NewInt(0)
	for _, out := range fundTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			collateralAmount = collateralAmount.Add(math.NewInt(out.Value))
		}
	}

	currentPrice, err := m.GetPrice(ctx, "")
	if err != nil {
		return nil, err
	}

	// TODO: retrieve from params
	collateralDecimal := math.NewInt(100000000)
	borrowedDecimal := math.NewInt(1000000)

	poolConfig := m.GetPool(ctx, loan.PoolId).Config

	// verify LTV
	// collateral value * ltv > borrow amount
	if collateralAmount.Mul(currentPrice).Mul(borrowedDecimal).Quo(collateralDecimal).Mul(math.NewInt(int64(poolConfig.Ltv))).Quo(types.Percent).LT(loan.BorrowAmount.Amount) {
		return nil, types.ErrInsufficientCollateral
	}

	// verify liquidation events. TODO improve price interval
	// if collateralAmount.Mul(event.TriggerPrice).Mul(borrowedDecimal).Quo(event.PriceDecimal).Mul(params.LiquidationThresholdPercent).Quo(types.Percent).LT(msg.BorrowAmount.Amount) {
	// 	return nil, types.ErrInvalidPriceEvent
	// }

	loan.CollateralAmount = collateralAmount
	loan.EventId = msg.EventId
	loan.DepositTxs = append(loan.DepositTxs, depositTxid)

	m.SetLoan(ctx, loan)

	depositLog := &types.DepositLog{
		Txid:         depositTxid,
		VaultAddress: loan.VaultAddress,
		DepositTx:    msg.DepositTx,
	}

	m.SetDepositLog(ctx, depositLog)

	// set dlc meta
	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return &types.MsgSubmitLiquidationCetResponse{}, nil
}

// Approve implements types.MsgServer.
func (m msgServer) Approve(goCtx context.Context, msg *types.MsgApprove) (*types.MsgApproveResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasDepositLog(ctx, msg.DepositTxId) {
		return nil, types.ErrDepositTxDoesNotExist
	}

	log := m.GetDepositLog(ctx, msg.DepositTxId)
	if !m.HasLoan(ctx, log.VaultAddress) {
		return nil, types.ErrLoanDoesNotExist
	}

	// Do not validate tx for now
	// if _, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, log.DepositTx, "", msg.BlockHash, msg.Proof); err != nil {
	// 	return nil, types.ErrInvalidProof
	// }

	loan := m.GetLoan(ctx, log.VaultAddress)

	loan.Status = types.LoanStatus_Approved
	m.SetLoan(ctx, loan)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("vault", loan.VaultAddress),
		sdk.NewAttribute("deposit_tx", msg.DepositTxId),
		sdk.NewAttribute("proof", fmt.Sprintf("%s", msg.Proof)),
		sdk.NewAttribute("block_hash", msg.BlockHash),
	)

	return &types.MsgApproveResponse{}, nil
}

// Cancel implements types.MsgServer.
func (m msgServer) Cancel(goCtx context.Context, msg *types.MsgCancel) (*types.MsgCancelResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	if loan.Status != types.LoanStatus_Rejected {
		return nil, types.ErrInvalidLoanStatus
	}

	if len(loan.DepositTxs) == 0 {
		return nil, types.ErrDepositTxDoesNotExist
	}

	depositTxId := loan.DepositTxs[0]

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.Tx)), true)

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).RepaymentScript)
	sigHashes := []string{}

	merkleTree := types.GetTapscriptTree(types.GetDLCTapscripts(m.GetDLCMeta(ctx, msg.LoanId)))
	scriptProof := merkleTree.LeafMerkleProofs[0]

	controlBlock, err := types.GetControlBlock(types.GetInternalKey(), scriptProof)
	if err != nil {
		return nil, err
	}

	for i, signature := range msg.Signatures {
		if p.UnsignedTx.TxIn[i].PreviousOutPoint.Hash.String() != depositTxId {
			return nil, errorsmod.Wrap(types.ErrDepositTxDoesNotExist, "mismatched deposit tx hash")
		}

		sigBytes, _ := hex.DecodeString(signature)

		sigHash, err := types.CalcTapscriptSigHash(p, i, types.DefaultSigHashType, script)
		if err != nil {
			return nil, err
		}

		if !schnorr.Verify(sigBytes, sigHash, borrowerPubKey) {
			return nil, types.ErrInvalidSignature
		}

		sigHashes = append(sigHashes, base64.StdEncoding.EncodeToString(sigHash))

		p.Inputs[i].TaprootInternalKey = btcschnorr.SerializePubKey(types.GetInternalKey())
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       script,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return nil, err
	}

	loan.Status = types.LoanStatus_Cancelled
	m.SetLoan(ctx, loan)

	cancellation := &types.Cancellation{
		LoanId:     msg.LoanId,
		Txid:       p.UnsignedTx.TxHash().String(),
		Tx:         serializedTx,
		Signatures: msg.Signatures,
		CreateAt:   ctx.BlockTime(),
	}
	m.SetCancellation(ctx, cancellation)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCancel,
			sdk.NewAttribute(types.AttributeKeyBorrower, msg.Borrower),
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
			sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(sigHashes, types.AttributeValueSeparator)),
		),
	)

	return &types.MsgCancelResponse{}, nil
}

// SubmitCancellationSignatures implements types.MsgServer.
func (m msgServer) SubmitCancellationSignatures(goCtx context.Context, msg *types.MsgSubmitCancellationSignatures) (*types.MsgSubmitCancellationSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasCancellation(ctx, msg.LoanId) {
		return nil, types.ErrCancellationDoesNotExist
	}

	cancellation := m.GetCancellation(ctx, msg.LoanId)
	if len(cancellation.DcaSignatures) != 0 {
		return nil, types.ErrDcaSignaturesAlreadyExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(cancellation.Tx)), true)
	if err != nil {
		return nil, err
	}

	if len(msg.Signatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).RepaymentScript)
	leafHash := txscript.NewBaseTapLeaf(script).TapHash()

	for i := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, types.DefaultSigHashType, script)
		if err != nil {
			return nil, err
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return nil, types.ErrInvalidSignature
		}

		borrowerSig, _ := hex.DecodeString(cancellation.Signatures[i])

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
			{
				XOnlyPubKey: agencyPubKey,
				LeafHash:    leafHash[:],
				Signature:   sigBytes,
				SigHash:     txscript.SigHashDefault,
			},
			{
				XOnlyPubKey: borrowerPubKey,
				LeafHash:    leafHash[:],
				Signature:   borrowerSig,
				SigHash:     txscript.SigHashDefault,
			},
		}
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return nil, err
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return nil, err
	}

	cancellation.Tx = serializedTx
	cancellation.DcaSignatures = msg.Signatures
	m.SetCancellation(ctx, cancellation)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateSignedCancellationTransaction,
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyTxHash, cancellation.Txid),
		),
	)

	return &types.MsgSubmitCancellationSignaturesResponse{}, nil
}

// Redeem implements types.MsgServer.
func (m msgServer) Redeem(goCtx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, types.ErrInvalidSender
	}

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	if types.HashLoanSecret(msg.LoanSecret) != loan.HashLoanSecret {
		return nil, types.ErrMismatchedLoanSecret
	}

	poolConfig := m.GetPool(ctx, loan.PoolId).Config

	redeemedAmount := sdk.NewInt64Coin(loan.BorrowAmount.Denom, loan.BorrowAmount.Amount.Int64()-poolConfig.OriginationFee.Int64())
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, sdk.NewCoins(redeemedAmount)); err != nil {
		return nil, err
	}

	originationFee := sdk.NewInt64Coin(loan.BorrowAmount.Denom, poolConfig.OriginationFee.Int64())
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(m.GetParams(ctx).OriginationFeeCollector), sdk.NewCoins(originationFee)); err != nil {
		return nil, err
	}

	// update pool
	m.AfterPoolBorrowed(ctx, loan.PoolId, *loan.BorrowAmount)

	loan.Status = types.LoanStatus_Open
	loan.LoanSecret = msg.LoanSecret

	m.SetLoan(ctx, loan)

	m.EmitEvent(ctx, msg.Borrower,
		sdk.NewAttribute("vault", loan.VaultAddress),
		sdk.NewAttribute("loan_secret", msg.LoanSecret),
	)

	return &types.MsgRedeemResponse{}, nil
}

// Repay implements types.MsgServer.
func (m msgServer) Repay(goCtx context.Context, msg *types.MsgRepay) (*types.MsgRepayResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if loan.Status != types.LoanStatus_Open {
		return nil, types.ErrInvalidLoanStatus
	}

	amount := loan.BorrowAmount.Amount.Add(loan.Interests)

	// send repayment to escrow account
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), types.RepaymentEscrowAccount, sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, amount))); err != nil {
		return nil, err
	}

	loan.Status = types.LoanStatus_Repaid
	m.SetLoan(ctx, loan)

	dls := []string{}
	for _, txid := range loan.DepositTxs {
		dl := m.GetDepositLog(ctx, txid)
		dls = append(dls, dl.DepositTx)
	}

	depositTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dls[0])), true)

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)

	internalKey, _ := hex.DecodeString(dlcMeta.InternalKey)
	tapscripts := types.GetDLCTapscripts(dlcMeta)

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)
	borrowerPkScript, _ := types.GetPkScriptFromAddress(loan.Borrower)

	feeRate := m.btcbridgeKeeper.GetFeeRate(ctx).Value
	if feeRate == 0 {
		// use default fee rate for now
		feeRate = 10
	}

	repaymentTx, err := types.CreateRepaymentTransaction(
		depositTx,
		vaultPkScript,
		borrowerPkScript,
		internalKey,
		tapscripts,
		feeRate,
	)
	if err != nil {
		return nil, err
	}

	repaymentTxPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentTx)), true)
	if err != nil {
		return nil, err
	}

	repayment := &types.Repayment{
		LoanId:            msg.LoanId,
		Txid:              repaymentTxPsbt.UnsignedTx.TxHash().String(),
		Tx:                repaymentTx,
		RepayAdaptorPoint: msg.AdaptorPoint,
		CreateAt:          ctx.BlockTime(),
	}

	m.SetRepayment(ctx, repayment)

	// get sig hashes
	sigHashes, err := types.GetRepaymentTxSigHashes(repaymentTx, dlcMeta.RepaymentScript)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRepay,
			sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyAdaptorPoint, msg.AdaptorPoint),
			sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
			sdk.NewAttribute(types.AttributeKeySigHashes, strings.Join(sigHashes, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AtrtibuteKeyRepaymentTxHash, repaymentTxPsbt.UnsignedTx.TxHash().String()),
		),
	)

	return &types.MsgRepayResponse{}, nil
}

// SubmitRepaymentAdaptorSignatures implements types.MsgServer.
func (m msgServer) SubmitRepaymentAdaptorSignatures(goCtx context.Context, msg *types.MsgSubmitRepaymentAdaptorSignatures) (*types.MsgSubmitRepaymentAdaptorSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasRepayment(ctx, msg.LoanId) {
		return nil, types.ErrInvalidRepayment
	}

	repayment := m.GetRepayment(ctx, msg.LoanId)
	if len(repayment.DcaAdaptorSignatures) != 0 {
		return nil, types.ErrRepaymentAdaptorSigsAlreadyExist
	}

	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repayment.Tx)), true)
	if err != nil {
		return nil, err
	}

	if len(msg.AdaptorSignatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidAdaptorSignatures, "mismatched adaptor signature number")
	}

	repaymentScript, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).RepaymentScript)

	adaptorPointBytes, _ := hex.DecodeString(repayment.RepayAdaptorPoint)
	agencyPubKeyBytes, _ := hex.DecodeString(m.GetLoan(ctx, msg.LoanId).Agency)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, repaymentScript)
		if err != nil {
			return nil, err
		}

		adaptorSigBytes, _ := hex.DecodeString(msg.AdaptorSignatures[i])

		if !adaptor.Verify(adaptorSigBytes, sigHash, agencyPubKeyBytes, adaptorPointBytes) {
			return nil, types.ErrInvalidAdaptorSignature
		}
	}

	repayment.DcaAdaptorSignatures = msg.AdaptorSignatures
	m.SetRepayment(ctx, repayment)

	return &types.MsgSubmitRepaymentAdaptorSignaturesResponse{}, nil
}

// SubmitLiquidationCetSignatures implements types.MsgServer.
func (m msgServer) SubmitLiquidationCetSignatures(goCtx context.Context, msg *types.MsgSubmitLiquidationCetSignatures) (*types.MsgSubmitLiquidationCetSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if loan.Status != types.LoanStatus_Liquidated {
		return nil, types.ErrLoanNotLiquidated
	}

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)
	if len(dlcMeta.LiquidationAgencySignatures) > 0 {
		return nil, types.ErrLiquidationSignaturesAlreadyExist
	}

	liquidationCet, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet)), true)

	if len(msg.Signatures) != len(liquidationCet.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidLiquidationSignatures, "mismatched signature number")
	}

	liquidationCetScript, _ := hex.DecodeString(dlcMeta.LiquidationCetScript)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	for i, input := range liquidationCet.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(liquidationCet, i, input.SighashType, liquidationCetScript)
		if err != nil {
			return nil, err
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return nil, types.ErrInvalidSignature
		}
	}

	dlcMeta.LiquidationAgencySignatures = msg.Signatures
	m.SetDLCMeta(ctx, msg.LoanId, dlcMeta)

	return &types.MsgSubmitLiquidationCetSignaturesResponse{}, nil
}

// Close implements types.MsgServer.
func (m msgServer) Close(goCtx context.Context, msg *types.MsgClose) (*types.MsgCloseResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if loan.Status == types.LoanStatus_Closed {
		return nil, types.ErrInvalidLoanStatus
	}

	if !m.HasRepayment(ctx, msg.LoanId) {
		return nil, types.ErrInvalidRepayment
	}

	repayment := m.GetRepayment(ctx, msg.LoanId)
	if len(repayment.DcaAdaptorSignatures) == 0 {
		return nil, types.ErrRepaymentAdaptorSigsDoNotExist
	}

	sigBytes, _ := hex.DecodeString(msg.Signature)
	adaptorSigBytes, _ := hex.DecodeString(repayment.DcaAdaptorSignatures[0])

	// extract secret from signatures
	secret := adaptor.Extract(adaptorSigBytes, sigBytes)
	if len(secret) == 0 {
		return nil, types.ErrInvalidSignature
	}

	if types.AdaptorPoint(secret) != repayment.RepayAdaptorPoint {
		return nil, types.ErrInvalidRepaymentSecret
	}

	amount := loan.BorrowAmount.Amount.Add(loan.Interests).Sub(loan.Fees)
	if err := m.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RepaymentEscrowAccount, types.ModuleName, sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, amount))); err != nil {
		return nil, err
	}

	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RepaymentEscrowAccount, sdk.MustAccAddressFromBech32(m.GetParams(ctx).ProtocolFeeCollector), sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, loan.Fees))); err != nil {
		return nil, err
	}

	// update pool
	m.AfterPoolRepaid(ctx, loan.PoolId, *loan.BorrowAmount, loan.Interests.Sub(loan.Fees))

	loan.Status = types.LoanStatus_Closed
	m.SetLoan(ctx, loan)

	repayment.BorrowerSignature = msg.Signature
	m.SetRepayment(ctx, repayment)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("loan_id", loan.VaultAddress),
		sdk.NewAttribute("payment_secret", hex.EncodeToString(secret)),
	)

	return &types.MsgCloseResponse{}, nil
}
