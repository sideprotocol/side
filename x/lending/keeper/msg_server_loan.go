package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/bitcoin/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
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

	if err := types.CheckBorrowAmountLimit(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, err
	}

	if err := types.CheckBorrowCap(pool, msg.BorrowAmount.Amount); err != nil {
		return nil, err
	}

	if msg.BorrowAmount.Amount.GT(pool.AvailableAmount) {
		return nil, types.ErrInsufficientLiquidity
	}

	trancheConfig, found := types.GetTrancheConfig(poolConfig.Tranches, msg.Maturity)
	if !found {
		return nil, errorsmod.Wrap(types.ErrInvalidMaturity, "maturity does not exist")
	}

	if err := m.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), sdk.MustAccAddressFromBech32(m.RequestFeeCollector(ctx)), sdk.NewCoins(poolConfig.RequestFee)); err != nil {
		return nil, err
	}

	if !m.dlcKeeper.HasDCM(ctx, msg.DCMId) {
		return nil, errorsmod.Wrap(types.ErrInvalidDCM, "dcm does not exist")
	}

	dcm := m.dlcKeeper.GetDCM(ctx, msg.DCMId)

	maturityTime := ctx.BlockTime().Add(time.Duration(trancheConfig.Maturity) * time.Second).Unix()
	finalTimeout := maturityTime + m.FinalTimeoutDuration(ctx)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, msg.BorrowerAuthPubkey, dcm.Pubkey, finalTimeout)
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	dlcMeta, err := types.BuildDLCMeta(msg.BorrowerPubkey, msg.BorrowerAuthPubkey, dcm.Pubkey, finalTimeout)
	if err != nil {
		return nil, err
	}

	dlcEvent := m.dlcKeeper.GetAvailableLendingEvent(ctx)
	if dlcEvent == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidEvent, "no available dlc lending event")
	}

	loan := &types.Loan{
		VaultAddress:       vault,
		Borrower:           msg.Borrower,
		BorrowerPubKey:     msg.BorrowerPubkey,
		BorrowerAuthPubKey: msg.BorrowerAuthPubkey,
		DCM:                dcm.Pubkey,
		MaturityTime:       maturityTime,
		FinalTimeout:       finalTimeout,
		PoolId:             msg.PoolId,
		BorrowAmount:       msg.BorrowAmount,
		RequestFee:         poolConfig.RequestFee,
		OriginationFee:     msg.BorrowAmount.Amount.ToLegacyDec().Mul(poolConfig.OriginationFeeFactor).TruncateInt(),
		Maturity:           trancheConfig.Maturity,
		BorrowAPR:          trancheConfig.BorrowAPR,
		DlcEventId:         dlcEvent.Id,
		Referrer:           m.GetReferrer(ctx, msg.ReferralCode),
		CreateAt:           ctx.BlockTime(),
		Status:             types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)
	m.SetLoanByAddress(ctx, loan)
	m.SetLoanByOracle(ctx, loan.VaultAddress, dlcEvent.Pubkey)

	// set dlc meta
	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	// update dlc event
	m.UpdateDLCEvent(ctx, loan)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeApply,
			sdk.NewAttribute(types.AttributeKeyVault, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyBorrower, loan.Borrower),
			sdk.NewAttribute(types.AttributeKeyBorrowerPubKey, loan.BorrowerPubKey),
			sdk.NewAttribute(types.AttributeKeyBorrowerAuthPubKey, loan.BorrowerAuthPubKey),
			sdk.NewAttribute(types.AttributeKeyDCMPubKey, loan.DCM),
			sdk.NewAttribute(types.AttributeKeyMaturityTime, fmt.Sprint(loan.MaturityTime)),
			sdk.NewAttribute(types.AttributeKeyFinalTimeout, fmt.Sprint(loan.FinalTimeout)),
			sdk.NewAttribute(types.AttributeKeyPoolId, loan.PoolId),
			sdk.NewAttribute(types.AttributeKeyBorrowAmount, loan.BorrowAmount.String()),
			sdk.NewAttribute(types.AttributeKeyDLCEventId, fmt.Sprintf("%d", loan.DlcEventId)),
			sdk.NewAttribute(types.AttributeKeyOraclePubKey, dlcEvent.Pubkey),
			sdk.NewAttribute(types.AttributeKeyReferralCode, msg.ReferralCode),
		),
	)

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
	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	// NOTE: only can be authorized once for now
	if loan.Status != types.LoanStatus_Requested {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan non requested")
	}

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	// parse deposit txs
	depositTxs, depositTxHashes, collateralAmount, err := types.ParseDepositTxs(msg.DepositTxs, vaultPkScript)
	if err != nil {
		return nil, err
	}

	if collateralAmount.IsZero() {
		return nil, errorsmod.Wrap(types.ErrInsufficientCollateral, "collateral amount cannot be zero")
	}

	// calculate liquidation price
	liquidationPrice := m.GetLiquidationPrice(ctx, loan, collateralAmount)

	// update dlc event outcome
	dlcEvent := m.dlcKeeper.GetEvent(ctx, loan.DlcEventId)
	m.UpdateDLCEventLiquidatedOutcome(ctx, loan, dlcEvent, liquidationPrice)

	// get fee rate
	feeRate := m.btcbridgeKeeper.GetFeeRate(ctx)
	if err := m.btcbridgeKeeper.CheckFeeRate(ctx, feeRate); err != nil {
		return nil, err
	}

	// verify cets
	if err := types.VerifyCets(m.GetDLCMeta(ctx, msg.LoanId), depositTxs, vaultPkScript, loan.BorrowerPubKey, loan.BorrowerAuthPubKey, loan.DCM, dlcEvent, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures, feeRate.Value, m.MaxLiquidationFeeRateMultiplier(ctx)); err != nil {
		return nil, err
	}

	// update dlc meta
	if err := m.UpdateDLCMeta(ctx, msg.LoanId, depositTxs, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures, feeRate.Value); err != nil {
		return nil, err
	}

	// initiate signing request for repayment cet adaptor signatures from DCM
	if err := m.InitiateRepaymentCetSigningRequest(ctx, loan.VaultAddress); err != nil {
		return nil, err
	}

	// create authorization
	authorization := m.CreateAuthorization(ctx, msg.LoanId, depositTxHashes)

	// set deposit logs
	for i, depositTx := range msg.DepositTxs {
		if !m.HasDepositLog(ctx, depositTxHashes[i]) {
			m.NewDepositLog(ctx, depositTxHashes[i], msg.LoanId, authorization.Id, depositTx)
		}
	}

	// update loan
	loan.Authorizations = append(loan.Authorizations, *authorization)
	loan.CollateralAmount = collateralAmount
	loan.LiquidationPrice = liquidationPrice
	loan.Status = types.LoanStatus_Authorized
	m.SetLoan(ctx, loan)

	// update dlc event
	m.dlcKeeper.SetEvent(ctx, dlcEvent)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuthorize,
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyCollateralAmount, loan.CollateralAmount.String()),
			sdk.NewAttribute(types.AttributeKeyLiquidationPrice, types.FormatPrice(loan.LiquidationPrice)),
		),
	)

	return &types.MsgSubmitCetsResponse{}, nil
}

// SubmitDepositTransaction implements types.MsgServer.
func (m msgServer) SubmitDepositTransaction(goCtx context.Context, msg *types.MsgSubmitDepositTransaction) (*types.MsgSubmitDepositTransactionResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.Vault) {
		return nil, errorsmod.Wrap(types.ErrInvalidVault, "vault does not match any loan")
	}

	loan := m.GetLoan(ctx, msg.Vault)
	if loan.Status != types.LoanStatus_Requested && loan.Status != types.LoanStatus_Cancelled && loan.Status != types.LoanStatus_Authorized && loan.Status != types.LoanStatus_Rejected {
		return nil, types.ErrInvalidLoanStatus
	}

	// validate deposit tx
	tx, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, msg.DepositTx, "", msg.BlockHash, msg.Proof, m.btcbridgeKeeper.DepositConfirmationDepth(ctx))
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrInvalidDepositTx, "failed to validate tx: %v", err)
	}

	depositTxHash := tx.Hash().String()

	var depositLog *types.DepositLog
	if m.HasDepositLog(ctx, depositTxHash) {
		depositLog = m.GetDepositLog(ctx, depositTxHash)
		if depositLog.Status != types.DepositStatus_DEPOSIT_STATUS_PENDING {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx not pending")
		}
	} else {
		depositLog = &types.DepositLog{
			Txid:            depositTxHash,
			VaultAddress:    msg.Vault,
			DepositTx:       msg.DepositTx,
			AuthorizationId: m.GetAuthorizationId(ctx, msg.Vault) + 1,
		}
	}

	depositLog.Status = types.DepositStatus_DEPOSIT_STATUS_VERIFIED
	m.SetDepositLog(ctx, depositLog)

	return &types.MsgSubmitDepositTransactionResponse{}, nil
}

// Redeem implements types.MsgServer.
func (m msgServer) Redeem(goCtx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
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

	// check if the collateral is redeemable
	if !types.CollateralRedeemable(loan) {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan collateral not redeemable")
	}

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.Tx)), true)

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)

	internalKey, _ := hex.DecodeString(dlcMeta.InternalKey)
	script, controlBlock, _ := types.UnwrapLeafScript(dlcMeta.RepaymentScript)

	sigHashes := []string{}

	for i, ti := range p.UnsignedTx.TxIn {
		prevTxHash := ti.PreviousOutPoint.Hash.String()
		if !m.HasDepositLog(ctx, prevTxHash) {
			return nil, types.ErrDepositTxDoesNotExist
		}

		depositLog := m.GetDepositLog(ctx, prevTxHash)
		if depositLog.VaultAddress != msg.LoanId {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx does not match the loan id")
		}

		// check deposit status
		if depositLog.Status != types.DepositStatus_DEPOSIT_STATUS_VERIFIED {
			return nil, errorsmod.Wrap(types.ErrInvalidDepositTx, "deposit tx non verified")
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

		p.Inputs[i].TaprootInternalKey = internalKey
		p.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: controlBlock,
				Script:       script,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}

		// update deposit status
		depositLog.Status = types.DepositStatus_DEPOSIT_STATUS_REDEEMING
		m.SetDepositLog(ctx, depositLog)
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return nil, err
	}

	redemption := &types.Redemption{
		Id:         m.IncrementRedemptionId(ctx),
		LoanId:     msg.LoanId,
		Txid:       p.UnsignedTx.TxHash().String(),
		Tx:         serializedTx,
		Signatures: msg.Signatures,
		CreateAt:   ctx.BlockTime(),
	}
	m.SetRedemption(ctx, redemption)

	m.tssKeeper.InitiateSigningRequest(
		ctx,
		types.ModuleName,
		types.ToScopedId(redemption.Id),
		tsstypes.SigningType_SIGNING_TYPE_SCHNORR,
		int32(types.SigningIntent_SIGNING_INTENT_REDEMPTION),
		loan.DCM,
		sigHashes,
		nil,
	)

	// update loan status to cancelled if the current status is requested
	if loan.Status == types.LoanStatus_Requested {
		loan.Status = types.LoanStatus_Cancelled
		m.SetLoan(ctx, loan)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRedeem,
			sdk.NewAttribute(types.AttributeKeyBorrower, msg.Borrower),
			sdk.NewAttribute(types.AttributeKeyLoanId, msg.LoanId),
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", redemption.Id)),
		),
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
		LoanId:   msg.LoanId,
		Amount:   amount,
		CreateAt: ctx.BlockTime(),
	}
	m.SetRepayment(ctx, repayment)

	// trigger dlc event
	m.DLCKeeper().TriggerDLCEvent(ctx, loan.DlcEventId, types.RepaidOutcomeIndex)

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

// RegisterReferrer implements types.MsgServer.
func (m msgServer) RegisterReferrer(goCtx context.Context, msg *types.MsgRegisterReferrer) (*types.MsgRegisterReferrerResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the referrer already exists
	if m.HasReferrer(ctx, msg.ReferralCode) {
		return nil, types.ErrReferrerAlreadyExists
	}

	// create new referrer
	referrer := &types.Referrer{
		Name:              msg.Name,
		ReferralCode:      msg.ReferralCode,
		Address:           msg.Address,
		ReferralFeeFactor: msg.ReferralFeeFactor,
	}
	m.SetReferrer(ctx, referrer)

	// emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterReferrer,
			sdk.NewAttribute(types.AttributeKeyReferrerName, referrer.Name),
			sdk.NewAttribute(types.AttributeKeyReferralCode, referrer.ReferralCode),
			sdk.NewAttribute(types.AttributeKeyReferrerAddress, referrer.Address),
			sdk.NewAttribute(types.AttributeKeyReferralFeeFactor, referrer.ReferralFeeFactor.String()),
		),
	)

	return &types.MsgRegisterReferrerResponse{}, nil
}

// UpdateReferrer implements types.MsgServer.
func (m msgServer) UpdateReferrer(goCtx context.Context, msg *types.MsgUpdateReferrer) (*types.MsgUpdateReferrerResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// check if the referrer already exists
	if !m.HasReferrer(ctx, msg.ReferralCode) {
		return nil, types.ErrReferrerDoesNotExist
	}

	referrer := m.GetReferrer(ctx, msg.ReferralCode)

	// update referrer
	referrer.Name = msg.Name
	referrer.Address = msg.Address
	referrer.ReferralFeeFactor = msg.ReferralFeeFactor
	m.SetReferrer(ctx, referrer)

	// emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateReferrer,
			sdk.NewAttribute(types.AttributeKeyReferralCode, referrer.ReferralCode),
			sdk.NewAttribute(types.AttributeKeyReferrerName, referrer.Name),
			sdk.NewAttribute(types.AttributeKeyReferrerAddress, referrer.Address),
			sdk.NewAttribute(types.AttributeKeyReferralFeeFactor, referrer.ReferralFeeFactor.String()),
		),
	)

	return &types.MsgUpdateReferrerResponse{}, nil
}
