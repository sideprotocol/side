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
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

	if types.HasRequestFee(pool) {
		if err := m.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(msg.Borrower), sdk.MustAccAddressFromBech32(m.RequestFeeCollector(ctx)), sdk.NewCoins(poolConfig.RequestFee)); err != nil {
			return nil, err
		}
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
		OriginationFee:     poolConfig.OriginationFee,
		Maturity:           trancheConfig.Maturity,
		BorrowAPR:          trancheConfig.BorrowAPR,
		MinMaturity:        trancheConfig.Maturity * int64(trancheConfig.MinMaturityFactor) / 1000,
		DlcEventId:         dlcEvent.Id,
		Referrer:           msg.Referrer,
		CreateAt:           ctx.BlockTime(),
		Status:             types.LoanStatus_Requested,
	}

	m.SetLoan(ctx, loan)
	m.SetLoanByAddress(ctx, loan)

	// set dlc meta
	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	// update dlc event
	m.UpdateDLCEvent(ctx, loan)

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
	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	// NOTE: only can be authorized once for now
	if loan.Status != types.LoanStatus_Requested {
		return nil, errorsmod.Wrap(types.ErrInvalidLoanStatus, "loan non requested")
	}

	vaultPkScript, _ := types.GetPkScriptFromAddress(loan.VaultAddress)

	depositTxs := []*psbt.Packet{}
	depositTxHashes := []string{}
	collateralAmount := sdkmath.ZeroInt()

	for _, depositTx := range msg.DepositTxs {
		p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(depositTx)), true)

		depositTxs = append(depositTxs, p)
		depositTxHashes = append(depositTxHashes, p.UnsignedTx.TxHash().String())

		for _, out := range p.UnsignedTx.TxOut {
			if bytes.Equal(out.PkScript, vaultPkScript) {
				collateralAmount = collateralAmount.Add(sdkmath.NewInt(out.Value))
			}
		}
	}

	if collateralAmount.IsZero() {
		return nil, errorsmod.Wrap(types.ErrInsufficientCollateral, "collateral amount can not be zero")
	}

	// calculate liquidation price
	liquidationPrice := m.GetLiquidationPrice(ctx, loan, collateralAmount)

	// update dlc event outcome
	dlcEvent := m.dlcKeeper.GetEvent(ctx, loan.DlcEventId)
	m.UpdateDLCEventLiquidatedOutcome(ctx, loan, dlcEvent, liquidationPrice)

	// verify cets
	if err := types.VerifyCets(depositTxs, vaultPkScript, loan.BorrowerPubKey, loan.BorrowerAuthPubKey, loan.DCM, dlcEvent, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures); err != nil {
		return nil, err
	}

	// update dlc meta
	if err := m.UpdateDLCMeta(ctx, msg.LoanId, depositTxs, msg.LiquidationCet, msg.LiquidationAdaptorSignatures, msg.DefaultLiquidationAdaptorSignatures, msg.RepaymentCet, msg.RepaymentSignatures); err != nil {
		return nil, err
	}

	// create authorization
	authorization := m.CreateAuthorization(ctx, msg.LoanId, depositTxHashes)

	for i, depositTx := range msg.DepositTxs {
		if !m.HasDepositLog(ctx, depositTxHashes[i]) {
			depositLog := &types.DepositLog{
				Txid:            depositTxHashes[i],
				VaultAddress:    loan.VaultAddress,
				AuthorizationId: authorization.Id,
				DepositTx:       depositTx,
			}

			m.SetDepositLog(ctx, depositLog)
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
	if loan.Status != types.LoanStatus_Requested && loan.Status != types.LoanStatus_Authorized && loan.Status != types.LoanStatus_Rejected {
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

	// check if the min maturity reached
	// commented out for now
	// if ctx.BlockTime().Unix()-loan.DisburseAt.Unix() < loan.MinMaturity {
	// 	return nil, types.ErrMinMaturityNotReached
	// }

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
