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
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
)

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(goCtx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasPool(ctx, msg.PoolId) {
		return nil, types.ErrPoolDoesNotExist
	}

	pool := m.GetPool(ctx, msg.PoolId)
	if pool.Status != types.PoolStatus_ACTIVE {
		return nil, types.ErrInactivePool
	}

	if msg.BorrowAmount.Denom != pool.Supply.Denom {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "mismatched denom")
	}

	if msg.BorrowAmount.Amount.GT(pool.AvailableAmount) {
		return nil, types.ErrInsufficientLiquidity
	}

	if !m.dlcKeeper.HasAgency(ctx, msg.AgencyId) {
		return nil, errorsmod.Wrap(types.ErrInvalidAgency, "agency does not exist")
	}

	agency := m.dlcKeeper.GetAgency(ctx, msg.AgencyId)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, agency.Pubkey, msg.MaturityTime, msg.MaturityTime+m.FinalTimeoutDuration(ctx))
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	defaultLiquidationDate := types.GetDefaultLiquidationDate(msg.MaturityTime)
	if !m.dlcKeeper.HasEventByDate(ctx, defaultLiquidationDate) {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "default liquidation event does not exist")
	}

	defaultLiquidationEvent := m.dlcKeeper.GetEventByDate(ctx, defaultLiquidationDate)

	repaymentEvent := m.dlcKeeper.GetAvailableLendingEvent(ctx)
	if repaymentEvent == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "no available event for repayment")
	}

	poolConfig := m.GetPool(ctx, msg.PoolId).Config

	interest := msg.BorrowAmount.Amount.Mul(sdkmath.NewInt(int64(poolConfig.BorrowRate))).Quo(types.Permille)
	protocolFee := interest.Mul(sdkmath.NewInt(int64(poolConfig.ReserveFactor))).Quo(types.Permille)

	if msg.BorrowAmount.Amount.LTE(poolConfig.OriginationFee) {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "borrowed amount must be greater than origination fee")
	}

	loan := &types.Loan{
		VaultAddress:              vault,
		Borrower:                  msg.Borrower,
		BorrowerPubKey:            msg.BorrowerPubkey,
		Agency:                    agency.Pubkey,
		MaturityTime:              msg.MaturityTime,
		FinalTimeout:              msg.MaturityTime + m.FinalTimeoutDuration(ctx),
		BorrowAmount:              msg.BorrowAmount,
		OriginationFee:            poolConfig.OriginationFee,
		Interest:                  interest,
		ProtocolFee:               protocolFee,
		DefaultLiquidationEventId: defaultLiquidationEvent.Id,
		RepaymentEventId:          repaymentEvent.Id,
		PoolId:                    msg.PoolId,
		CreateAt:                  ctx.BlockTime(),
		Status:                    types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeApply,
			sdk.NewAttribute(types.AttributeKeyVault, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyBorrower, loan.Borrower),
			sdk.NewAttribute(types.AttributeKeyAgencyPubKey, loan.Agency),
			sdk.NewAttribute(types.AttributeKeyMuturityTime, fmt.Sprint(loan.MaturityTime)),
			sdk.NewAttribute(types.AttributeKeyFinalTimeout, fmt.Sprint(loan.FinalTimeout)),
			sdk.NewAttribute(types.AttributeKeyBorrowAmount, loan.BorrowAmount.String()),
			sdk.NewAttribute(types.AttributeKeyPoolId, loan.PoolId),
		))

	return &types.MsgApplyResponse{}, nil
}

// SubmitCets implements types.MsgServer.
func (m msgServer) SubmitCets(goCtx context.Context, msg *types.MsgSubmitCets) (*types.MsgSubmitCetsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	poolConfig := m.GetPool(ctx, msg.LoanId).Config

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	fundTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.DepositTx)), true)
	depositTxid := fundTx.UnsignedTx.TxHash().String()

	collateralAmount := sdkmath.ZeroInt()
	for _, out := range fundTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			collateralAmount = collateralAmount.Add(sdkmath.NewInt(out.Value))
		}
	}

	liquidationPrice := types.GetLiquidationPrice(collateralAmount, loan.BorrowAmount.Amount, sdkmath.NewInt(int64(poolConfig.LiquidationThreshold)))
	if !m.dlcKeeper.HasEventByPrice(ctx, liquidationPrice) {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event does not exist")
	}

	liquidationEvent := m.dlcKeeper.GetEventByPrice(ctx, liquidationPrice)
	if liquidationEvent.HasTriggered {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event has triggered")
	}

	defaultLiquidationEvent := m.dlcKeeper.GetEvent(ctx, loan.DefaultLiquidationEventId)

	if err := types.VerifyCets(fundTx, loan.BorrowerPubKey, loan.Agency, liquidationEvent, defaultLiquidationEvent, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures); err != nil {
		return nil, err
	}

	dlcMeta, err := types.BuildDLCMeta(fundTx, vaultPkScript, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures, loan.BorrowerPubKey, loan.Agency, loan.MaturityTime, loan.FinalTimeout)
	if err != nil {
		return nil, err
	}

	currentPrice, err := m.GetPrice(ctx, "")
	if err != nil {
		return nil, err
	}

	// TODO: retrieve from params
	collateralDecimal := sdkmath.NewInt(100000000)
	borrowedDecimal := sdkmath.NewInt(1000000)

	// verify LTV
	// collateral value * ltv > borrow amount
	if collateralAmount.Mul(currentPrice).Mul(borrowedDecimal).Quo(collateralDecimal).Mul(sdkmath.NewInt(int64(poolConfig.Ltv))).Quo(types.Percent).LT(loan.BorrowAmount.Amount) {
		return nil, types.ErrInsufficientCollateral
	}

	// verify liquidation events. TODO improve price interval
	// if collateralAmount.Mul(event.TriggerPrice).Mul(borrowedDecimal).Quo(event.PriceDecimal).Mul(params.LiquidationThresholdPercent).Quo(types.Percent).LT(msg.BorrowAmount.Amount) {
	// 	return nil, types.ErrInvalidPriceEvent
	// }

	loan.CollateralAmount = collateralAmount
	loan.LiquidationEventId = liquidationEvent.Id
	loan.DepositTxs = append(loan.DepositTxs, depositTxid)

	m.SetLoan(ctx, loan)

	depositLog := &types.DepositLog{
		Txid:         depositTxid,
		VaultAddress: loan.VaultAddress,
		DepositTx:    msg.DepositTx,
	}

	m.SetDepositLog(ctx, depositLog)

	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return &types.MsgSubmitCetsResponse{}, nil
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

	depositLog := m.GetDepositLog(ctx, msg.DepositTxId)
	loan := m.GetLoan(ctx, depositLog.VaultAddress)

	// Do not validate tx for now
	// if _, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, depositLog.DepositTx, "", msg.BlockHash, msg.Proof); err != nil {
	// 	return nil, types.ErrInvalidProof
	// }

	amount := sdk.NewInt64Coin(loan.BorrowAmount.Denom, loan.BorrowAmount.Amount.Int64()-loan.OriginationFee.Int64())
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(loan.Borrower), sdk.NewCoins(amount)); err != nil {
		return nil, err
	}

	originationFee := sdk.NewInt64Coin(loan.BorrowAmount.Denom, loan.OriginationFee.Int64())
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromBech32(m.GetParams(ctx).OriginationFeeCollector), sdk.NewCoins(originationFee)); err != nil {
		return nil, err
	}

	// update pool
	m.AfterPoolBorrowed(ctx, loan.PoolId, loan.BorrowAmount)

	loan.Status = types.LoanStatus_Open
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
	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
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

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(cancellation.Tx)), true)
	if len(msg.Signatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
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
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not open")
	}

	// escrow repaid amount
	amount := loan.BorrowAmount.Amount.Add(loan.Interest)
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

	repaymentTx, err := types.CreateRepaymentCet(
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
		LoanId:       msg.LoanId,
		Txid:         repaymentTxPsbt.UnsignedTx.TxHash().String(),
		Tx:           repaymentTx,
		AdaptorPoint: msg.AdaptorPoint,
		CreateAt:     ctx.BlockTime(),
	}

	m.SetRepayment(ctx, repayment)

	// get sig hashes
	sigHashes, err := types.GetRepaymentCetSigHashes(dlcMeta)
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

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)
	if loan.Status != types.LoanStatus_Repaid {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not repaid")
	}

	repayment := m.GetRepayment(ctx, msg.LoanId)
	if len(repayment.DcaAdaptorSignatures) != 0 {
		return nil, types.ErrRepaymentAdaptorSigsAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(repayment.Tx)), true)
	if len(msg.AdaptorSignatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidAdaptorSignatures, "mismatched adaptor signature number")
	}

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
	adaptorPointBytes, _ := hex.DecodeString(repayment.AdaptorPoint)
	agencyPubKeyBytes, _ := hex.DecodeString(m.GetLoan(ctx, msg.LoanId).Agency)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
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
	if len(dlcMeta.LiquidationCet.AgencySignatures) > 0 {
		return nil, types.ErrLiquidationSignaturesAlreadyExist
	}

	liquidationCet, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet.Tx)), true)
	if len(msg.Signatures) != len(liquidationCet.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	for i, input := range liquidationCet.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(liquidationCet, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return nil, types.ErrInvalidSignature
		}
	}

	dlcMeta.LiquidationCet.AgencySignatures = msg.Signatures
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
	if loan.Status != types.LoanStatus_Repaid {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan not repaid")
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
		return nil, errorsmod.Wrap(types.ErrInvalidSignature, "failed to reveal adaptor secret")
	}

	if types.AdaptorPointFromSecret(secret) != repayment.AdaptorPoint {
		return nil, errorsmod.Wrap(types.ErrInvalidSignature, "revealed adaptor secret does not match the adaptor point")
	}

	amount := loan.BorrowAmount.Amount.Add(loan.Interest).Sub(loan.ProtocolFee)
	if err := m.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RepaymentEscrowAccount, types.ModuleName, sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, amount))); err != nil {
		return nil, err
	}

	protocolFee := sdk.NewCoin(loan.BorrowAmount.Denom, loan.ProtocolFee)
	if err := m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RepaymentEscrowAccount, sdk.MustAccAddressFromBech32(m.GetParams(ctx).ProtocolFeeCollector), sdk.NewCoins(protocolFee)); err != nil {
		return nil, err
	}

	// update pool
	m.AfterPoolRepaid(ctx, loan.PoolId, loan.BorrowAmount, loan.Interest.Sub(loan.ProtocolFee))

	loan.Status = types.LoanStatus_Closed
	m.SetLoan(ctx, loan)

	repayment.DcaAdaptedSignature = msg.Signature
	m.SetRepayment(ctx, repayment)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("loan_id", loan.VaultAddress),
		sdk.NewAttribute("payment_secret", hex.EncodeToString(secret)),
	)

	return &types.MsgCloseResponse{}, nil
}
