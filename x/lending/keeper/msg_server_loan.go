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

// Apply implements types.MsgServer.
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
		return nil, types.ErrPoolNotActive
	}

	poolConfig := m.GetPool(ctx, msg.PoolId).Config

	if msg.BorrowAmount.Denom != pool.Supply.Denom {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "mismatched denom")
	}

	if msg.BorrowAmount.Amount.LT(poolConfig.MinBorrowAmount) {
		return nil, errorsmod.Wrap(types.ErrInvalidAmount, "borrow amount can not be less than min borrow amount")
	}

	if msg.BorrowAmount.Amount.GT(pool.AvailableAmount) {
		return nil, types.ErrInsufficientLiquidity
	}

	if err := types.CheckBorrowCap(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, types.ErrBorrowCapExceeded
	}

	if err := types.CheckDebtCeiling(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, types.ErrDebtCeilingExceeded
	}

	duration := msg.MaturityTime - ctx.BlockTime().Unix()
	if duration < m.MinLoanDuration(ctx) || duration > m.MaxLoanDuration(ctx) {
		return nil, types.ErrInvalidLoanDuration
	}

	if !m.dlcKeeper.HasDCM(ctx, msg.DCMId) {
		return nil, errorsmod.Wrap(types.ErrInvalidDCM, "dcm does not exist")
	}

	dcm := m.dlcKeeper.GetDCM(ctx, msg.DCMId)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, dcm.Pubkey, msg.MaturityTime, msg.MaturityTime+m.FinalTimeoutDuration(ctx))
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

	// update repayment event
	repaymentEvent.Description = fmt.Sprintf("repayment event for loan %s", vault)
	repaymentEvent.Outcomes = []string{vault}
	m.dlcKeeper.SetEvent(ctx, repaymentEvent)

	interest := msg.BorrowAmount.Amount.Mul(sdkmath.NewInt(int64(poolConfig.BorrowAPR))).Mul(sdkmath.NewInt(int64(msg.MaturityTime - ctx.BlockTime().Unix()))).Quo(sdkmath.NewInt(int64(types.OneYear))).Quo(types.Permille)
	protocolFee := interest.Mul(sdkmath.NewInt(int64(poolConfig.ReserveFactor))).Quo(types.Permille)

	loan := &types.Loan{
		VaultAddress:              vault,
		Borrower:                  msg.Borrower,
		BorrowerPubKey:            msg.BorrowerPubkey,
		DCM:                       dcm.Pubkey,
		MaturityTime:              msg.MaturityTime,
		FinalTimeout:              msg.MaturityTime + m.FinalTimeoutDuration(ctx),
		PoolId:                    msg.PoolId,
		BorrowAmount:              msg.BorrowAmount,
		OriginationFee:            poolConfig.OriginationFee,
		Interest:                  interest,
		ProtocolFee:               protocolFee,
		Term:                      duration,
		DefaultLiquidationEventId: defaultLiquidationEvent.Id,
		RepaymentEventId:          repaymentEvent.Id,
		CreateAt:                  ctx.BlockTime(),
		Status:                    types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)
	m.SetLoanByAddress(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeApply,
			sdk.NewAttribute(types.AttributeKeyVault, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyBorrower, loan.Borrower),
			sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
			sdk.NewAttribute(types.AttributeKeyMuturityTime, fmt.Sprint(loan.MaturityTime)),
			sdk.NewAttribute(types.AttributeKeyFinalTimeout, fmt.Sprint(loan.FinalTimeout)),
			sdk.NewAttribute(types.AttributeKeyPoolId, loan.PoolId),
			sdk.NewAttribute(types.AttributeKeyBorrowAmount, loan.BorrowAmount.String()),
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
	pool := m.GetPool(ctx, loan.PoolId)
	poolConfig := pool.Config

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	fundTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.DepositTx)), true)
	depositTxHash := fundTx.UnsignedTx.TxHash().String()

	collateralAmount := sdkmath.ZeroInt()
	for _, out := range fundTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			collateralAmount = collateralAmount.Add(sdkmath.NewInt(out.Value))
		}
	}

	if collateralAmount.IsZero() {
		return nil, errorsmod.Wrap(types.ErrInsufficientCollateral, "collateral amount can not be zero")
	}

	dlcMeta, err := types.BuildDLCMeta(fundTx, vaultPkScript, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures, loan.BorrowerPubKey, loan.DCM, loan.MaturityTime, loan.FinalTimeout)
	if err != nil {
		return nil, err
	}

	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	if !m.HasDepositLog(ctx, depositTxHash) {
		depositLog := &types.DepositLog{
			Txid:         depositTxHash,
			VaultAddress: loan.VaultAddress,
			DepositTx:    msg.DepositTx,
		}

		m.SetDepositLog(ctx, depositLog)
	}

	loan.CollateralAmount = collateralAmount
	loan.DepositTxs = append(loan.DepositTxs, depositTxHash)

	var errRejected error

	defer func() {
		if errRejected != nil {
			m.Logger(ctx).Info("loan rejected", "reason", errRejected)

			loan.Status = types.LoanStatus_Rejected
			m.SetLoan(ctx, loan)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeReject,
					sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
					sdk.NewAttribute(types.AttributeKeyDepositTxHash, depositTxHash),
				),
			)
		}
	}()

	if ctx.BlockTime().Unix() >= loan.MaturityTime {
		errRejected = types.ErrMaturityTimeReached
		return nil, nil
	}

	if pool.AvailableAmount.LT(loan.BorrowAmount.Amount) {
		errRejected = types.ErrInsufficientLiquidity
		return nil, nil
	}

	if err := types.CheckBorrowCap(pool, loan.BorrowAmount.Amount); err != nil {
		errRejected = types.ErrBorrowCapExceeded
		return nil, nil
	}

	if err := types.CheckDebtCeiling(pool, loan.BorrowAmount.Amount); err != nil {
		errRejected = types.ErrDebtCeilingExceeded
		return nil, nil
	}

	liquidationPrice := types.GetLiquidationPrice(collateralAmount, loan.BorrowAmount.Amount, loan.MaturityTime-loan.CreateAt.Unix(), poolConfig.BorrowAPR, poolConfig.LiquidationThreshold)
	if !m.dlcKeeper.HasEventByPrice(ctx, liquidationPrice) {
		errRejected = errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event does not exist")
		return nil, nil
	}

	liquidationEvent := m.dlcKeeper.GetEventByPrice(ctx, liquidationPrice)
	if liquidationEvent.HasTriggered {
		errRejected = errorsmod.Wrap(types.ErrInvalidEvent, "liquidation event has triggered")
		return nil, nil
	}

	defaultLiquidationEvent := m.dlcKeeper.GetEvent(ctx, loan.DefaultLiquidationEventId)

	if err := types.VerifyCets(fundTx, loan.BorrowerPubKey, loan.DCM, liquidationEvent, defaultLiquidationEvent, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures); err != nil {
		return nil, err
	}

	currentPrice, err := m.GetPrice(ctx, "BTCUSD")
	if err != nil {
		return nil, err
	}

	// check LTV
	if collateralAmount.Mul(sdkmath.NewIntWithDecimal(1, 6)).Mul(sdkmath.NewInt(int64(poolConfig.MaxLtv))).ToLegacyDec().Mul(currentPrice).Quo(sdkmath.NewIntWithDecimal(1, 8).Mul(types.Percent).ToLegacyDec()).TruncateInt().LT(loan.BorrowAmount.Amount) {
		errRejected = types.ErrInsufficientCollateral
		return nil, nil
	}

	// if deposit tx already verified, approve the loan
	if m.GetDepositLog(ctx, depositTxHash).Verified {
		if err := m.HandleApproval(ctx, msg.Borrower, depositTxHash, loan); err != nil {
			errRejected = err
			return nil, nil
		}
	}

	loan.LiquidationPrice = liquidationPrice
	loan.LiquidationEventId = liquidationEvent.Id

	m.SetLoan(ctx, loan)

	return &types.MsgSubmitCetsResponse{}, nil
}

// Approve implements types.MsgServer.
func (m msgServer) Approve(goCtx context.Context, msg *types.MsgApprove) (*types.MsgApproveResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.Vault) {
		return nil, errorsmod.Wrap(types.ErrInvalidVault, "vault does not match any loan")
	}

	loan := m.GetLoan(ctx, msg.Vault)
	if loan.Status != types.LoanStatus_Requested && loan.Status != types.LoanStatus_Rejected {
		return nil, types.ErrInvalidLoanStatus
	}

	// Do not validate tx for now
	// if _, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, depositLog.DepositTx, "", msg.BlockHash, msg.Proof); err != nil {
	// 	return nil, types.ErrInvalidProof
	// }

	depositTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.DepositTx)), true)
	depositTxHash := depositTx.UnsignedTx.TxHash().String()

	depositTxAlreadyExists := true
	if !m.HasDepositLog(ctx, depositTxHash) {
		depositTxAlreadyExists = false

		depositLog := &types.DepositLog{
			Txid:         depositTxHash,
			VaultAddress: msg.Vault,
			DepositTx:    msg.DepositTx,
		}

		m.SetDepositLog(ctx, depositLog)
	}

	depositLog := m.GetDepositLog(ctx, depositTxHash)
	if depositLog.Verified {
		return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx already verified")
	}

	depositLog.Verified = true
	m.SetDepositLog(ctx, depositLog)

	// cets submission rejected
	if loan.Status == types.LoanStatus_Rejected {
		return nil, nil
	}

	var errRejected error

	defer func() {
		if errRejected != nil {
			m.Logger(ctx).Info("loan rejected", "reason", errRejected)

			loan.Status = types.LoanStatus_Rejected
			m.SetLoan(ctx, loan)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeReject,
					sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
					sdk.NewAttribute(types.AttributeKeyDepositTxHash, depositTxHash),
				),
			)
		}
	}()

	if ctx.BlockTime().Unix() >= loan.MaturityTime {
		errRejected = types.ErrMaturityTimeReached
		return nil, nil
	}

	pool := m.GetPool(ctx, loan.PoolId)

	if err := types.CheckBorrowCap(pool, loan.BorrowAmount.Amount); err != nil {
		errRejected = types.ErrBorrowCapExceeded
		return nil, nil
	}

	if err := types.CheckDebtCeiling(pool, loan.BorrowAmount.Amount); err != nil {
		errRejected = types.ErrDebtCeilingExceeded
		return nil, nil
	}

	liquidationPrice := sdkmath.ZeroInt()

	// loan requested
	if !depositTxAlreadyExists {
		vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

		collateralAmount := sdkmath.ZeroInt()
		for _, out := range depositTx.UnsignedTx.TxOut {
			if bytes.Equal(out.PkScript, vaultPkScript) {
				collateralAmount = collateralAmount.Add(sdkmath.NewInt(out.Value))
			}

			poolConfig := m.GetPool(ctx, loan.PoolId).Config
			liquidationPrice = types.GetLiquidationPrice(collateralAmount, loan.BorrowAmount.Amount, loan.MaturityTime-loan.CreateAt.Unix(), poolConfig.BorrowAPR, poolConfig.LiquidationThreshold)
		}
	} else {
		liquidationPrice = loan.LiquidationPrice
	}

	currentPrice, err := m.GetPrice(ctx, "BTCUSD")
	if err != nil {
		return nil, err
	}

	// check if liquidation price reached
	if currentPrice.LTE(liquidationPrice.ToLegacyDec()) {
		errRejected = types.ErrLiquidationPriceReached
		return nil, nil
	}

	// cets submitted
	if depositTxAlreadyExists {
		if err := m.HandleApproval(ctx, msg.Relayer, depositTxHash, loan); err != nil {
			errRejected = err
			return nil, nil
		}
	}

	return &types.MsgApproveResponse{}, nil
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
	if loan.Status != types.LoanStatus_Open && loan.Status != types.LoanStatus_Repaid {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan neither open nor repaid")
	}

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)

	repaymentCet := dlcMeta.RepaymentCet
	if len(repaymentCet.DCMAdaptorSignatures) != 0 {
		return nil, types.ErrRepaymentAdaptorSigsAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet.Tx)), true)
	if len(msg.AdaptorSignatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidAdaptorSignatures, "mismatched adaptor signature number")
	}

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
	adaptorPoint, _ := m.GetRepaymentCetAdaptorPoint(ctx, msg.LoanId)
	dcmPubKey, _ := hex.DecodeString(m.GetLoan(ctx, msg.LoanId).DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return nil, err
		}

		adaptorSigBytes, _ := hex.DecodeString(msg.AdaptorSignatures[i])

		if !adaptor.Verify(adaptorSigBytes, sigHash, dcmPubKey, adaptorPoint) {
			return nil, types.ErrInvalidAdaptorSignature
		}
	}

	dlcMeta.RepaymentCet.DCMAdaptorSignatures = msg.AdaptorSignatures
	m.SetDLCMeta(ctx, msg.LoanId, dlcMeta)

	return &types.MsgSubmitRepaymentAdaptorSignaturesResponse{}, nil
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

	for i, ti := range p.UnsignedTx.TxIn {
		prevTxHash := ti.PreviousOutPoint.Hash.String()
		if !m.HasDepositLog(ctx, prevTxHash) {
			return nil, types.ErrDepositTxDoesNotExist
		}

		if !m.GetDepositLog(ctx, prevTxHash).Verified {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx not verified")
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

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
			sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
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
	if len(cancellation.DCMSignatures) != 0 {
		return nil, types.ErrDCMSignaturesAlreadyExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(cancellation.Tx)), true)
	if len(msg.Signatures) != len(p.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	script, _ := hex.DecodeString(m.GetDLCMeta(ctx, msg.LoanId).MultisigScript)
	leafHash := txscript.NewBaseTapLeaf(script).TapHash()

	for i := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, types.DefaultSigHashType, script)
		if err != nil {
			return nil, err
		}

		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return nil, types.ErrInvalidSignature
		}

		borrowerSig, _ := hex.DecodeString(cancellation.Signatures[i])

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
			{
				XOnlyPubKey: dcmPubKey,
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
	cancellation.DCMSignatures = msg.Signatures
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

	interest := m.GetCurrentInterest(ctx, loan)
	amount := loan.BorrowAmount.Add(interest)

	// escrow repaid amount
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), types.RepaymentEscrowAccount, sdk.NewCoins(amount)); err != nil {
		return nil, err
	}

	loan.Status = types.LoanStatus_Repaid
	m.SetLoan(ctx, loan)

	repayment := &types.Repayment{
		LoanId: msg.LoanId,
		Amount: amount,
	}
	m.SetRepayment(ctx, repayment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRepay,
			sdk.NewAttribute(types.AttributeKeyBorrower, msg.Borrower),
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyAmount, amount.String()),
		),
	)

	return &types.MsgRepayResponse{}, nil
}

// SubmitLiquidationSignatures implements types.MsgServer.
func (m msgServer) SubmitLiquidationSignatures(goCtx context.Context, msg *types.MsgSubmitLiquidationSignatures) (*types.MsgSubmitLiquidationSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	switch loan.Status {
	case types.LoanStatus_Liquidated:
		if err := m.handleLiquidationSignatures(ctx, loan, msg.Signatures); err != nil {
			return nil, err
		}

	case types.LoanStatus_Defaulted:
		if err := m.handleDefaultLiquidationSignatures(ctx, loan, msg.Signatures); err != nil {
			return nil, err
		}

	default:
		return nil, types.ErrLoanNotLiquidated
	}

	return &types.MsgSubmitLiquidationSignaturesResponse{}, nil
}
