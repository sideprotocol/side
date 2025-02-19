package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/schnorr"
	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
)

type msgServer struct {
	Keeper
}

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(goCtx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.dlcKeeper.HasEvent(ctx, msg.EventId) {
		return nil, types.ErrInvalidPriceEvent
	}

	event := m.dlcKeeper.GetEvent(ctx, msg.EventId)
	if event.HasTriggered {
		return nil, types.ErrInvalidPriceEvent
	}

	if !m.dlcKeeper.HasAgency(ctx, msg.AgencyId) {
		return nil, types.ErrInvalidAgency
	}

	agency := m.dlcKeeper.GetAgency(ctx, msg.AgencyId)

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, agency.Pubkey, msg.LoanSecretHash, msg.MaturityTime, msg.FinalTimeout)
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	fundTx, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.DepositTx)), true)
	if err != nil {
		return nil, types.ErrInvalidFunding
	}
	depositTxid := fundTx.UnsignedTx.TxHash().String()

	adaptorPoint, err := dlctypes.GetSignaturePointFromEvent(event)
	if err != nil {
		return nil, err
	}

	if err := types.VerifyLiquidationCET(fundTx, msg.LiquidationCet, msg.BorrowerPubkey, msg.LiquidationAdaptorSignature, hex.EncodeToString(adaptorPoint)); err != nil {
		return nil, err
	}

	vaultPkScript, _ := types.GetPkScriptFromAddress(vault)

	dlcMeta, err := types.BuildDLCMeta(fundTx, vaultPkScript, msg.LiquidationCet, msg.LiquidationAdaptorSignature, msg.BorrowerPubkey, agency.Pubkey, msg.LoanSecretHash, msg.MaturityTime, msg.FinalTimeout)
	if err != nil {
		return nil, err
	}

	params := m.GetParams(ctx)

	collateralAmount := math.NewInt(0)
	for _, out := range fundTx.UnsignedTx.TxOut {
		if bytes.Equal(out.PkScript, vaultPkScript) {
			collateralAmount.Add(math.NewInt(out.Value))
		}
	}

	currentPrice := math.NewInt(1) // read it from price oracle later
	decimal := math.NewInt(1)

	// verify LTV (Loan-to-Value Ratio)
	// collateral value * min_ltv > borrow amount
	if collateralAmount.Mul(currentPrice).Quo(decimal).Mul(params.MinInitialLtvPercent).Quo(types.Percent).LT(msg.BorrowAmount.Amount) {
		return nil, types.ErrInsufficientCollateral
	}

	// verify liquidation events. TODO improve price interval
	if collateralAmount.Mul(event.TriggerPrice).Quo(event.PriceDecimal).Mul(params.LiquidationThresholdPercent.Quo(types.Percent)).LT(msg.BorrowAmount.Amount) {
		return nil, types.ErrInvalidPriceEvent
	}

	interests := msg.BorrowAmount.Amount.Mul(params.BorrowRatePermille).Quo(types.Permille)
	fees := msg.BorrowAmount.Amount.Mul(params.BorrowRatePermille.Sub(params.SupplyRatePermille)).Quo(types.Permille)

	loan := types.Loan{
		VaultAddress:     vault,
		Borrower:         msg.Borrower,
		Agency:           agency.Pubkey,
		HashLoanSecret:   msg.LoanSecretHash,
		MaturityTime:     msg.MaturityTime,
		FinalTimeout:     msg.FinalTimeout,
		BorrowAmount:     msg.BorrowAmount,
		CollateralAmount: collateralAmount,
		Interests:        interests,
		Fees:             fees,
		EventId:          msg.EventId,
		DepositTxs:       []string{depositTxid},
		CreateAt:         ctx.BlockTime(),
		PoolId:           msg.PoolId,
		Status:           types.LoanStatus_Apply,
	}

	m.SetLoan(ctx, loan)

	depositLog := types.DepositLog{
		Txid:         depositTxid,
		VaultAddress: vault,
		DepositTx:    msg.DepositTx,
	}

	m.SetDepositLog(ctx, depositLog)

	// set dlc meta
	m.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	m.EmitEvent(ctx, msg.Borrower,
		sdk.NewAttribute("vault", loan.VaultAddress),
		sdk.NewAttribute("borrower", loan.Borrower),
		sdk.NewAttribute("agency", loan.Agency),
		sdk.NewAttribute("loan_secret_hash", loan.HashLoanSecret),
		sdk.NewAttribute("muturity_time", fmt.Sprint(loan.MaturityTime)),
		sdk.NewAttribute("final_timeout", fmt.Sprint(loan.FinalTimeout)),
		sdk.NewAttribute("borrow_amount", loan.BorrowAmount.String()),
		sdk.NewAttribute("collateral", loan.CollateralAmount.String()),
		sdk.NewAttribute("pool_id", loan.PoolId),
		sdk.NewAttribute("event_id", fmt.Sprintf("%d", loan.EventId)),
	)

	return &types.MsgApplyResponse{}, nil

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

	if _, _, err := m.btcbridgeKeeper.ValidateTransaction(ctx, log.DepositTx, "", msg.BlockHash, msg.Proof); err != nil {
		return nil, types.ErrInvalidProof
	}

	loan := m.GetLoan(ctx, log.VaultAddress)

	loan.Status = types.LoanStatus_Approve
	m.SetLoan(ctx, loan)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("vault", loan.VaultAddress),
		sdk.NewAttribute("deposit_tx", msg.DepositTxId),
		sdk.NewAttribute("proof", fmt.Sprintf("%s", msg.Proof)),
		sdk.NewAttribute("block_hash", msg.BlockHash),
	)

	return &types.MsgApproveResponse{}, nil
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

	m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, sdk.NewCoins(*loan.BorrowAmount))

	loan.Status = types.LoanStatus_Disburse
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

	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, types.ErrInvalidSender
	}

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	amount := loan.BorrowAmount.Amount.Add(loan.Interests).Add(loan.Fees)

	// send repayment to escrow account
	if e := m.bankKeeper.SendCoinsFromAccountToModule(ctx, borrower, types.RepaymentEscrowAccount, sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, amount))); e != nil {
		return nil, e
	}

	loan.Status = types.LoanStatus_Repay
	m.SetLoan(ctx, loan)

	dls := []string{}
	for _, txid := range loan.DepositTxs {
		dl := m.GetDepositLog(ctx, txid)
		dls = append(dls, dl.DepositTx)
	}

	depositTx, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dls[0])), true)

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)

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
		[]byte(dlcMeta.InternalKey),
		[][]byte{
			[]byte(dlcMeta.LiquidationCetScript),
			[]byte(dlcMeta.ForcedRepaymentScript),
			[]byte(dlcMeta.TimeoutRefundScript),
		},
		feeRate,
	)
	if err != nil {
		return nil, err
	}

	repaymentTxPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentTx)), true)
	if err != nil {
		return nil, err
	}

	repayment := types.Repayment{
		LoanId:            msg.LoanId,
		Txid:              repaymentTxPsbt.UnsignedTx.TxHash().String(),
		Tx:                repaymentTx,
		RepayAdaptorPoint: msg.AdaptorPoint,
		CreateAt:          ctx.BlockTime(),
	}

	m.SetRepayment(ctx, repayment)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRepay,
			sdk.NewAttribute(types.AttributeKeyLoanId, loan.VaultAddress),
			sdk.NewAttribute(types.AttributeKeyAdaptorPoint, msg.AdaptorPoint),
		),
	)

	return &types.MsgRepayResponse{}, nil
}

// SubmitRepaymentAdaptorSignature implements types.MsgServer.
func (m msgServer) SubmitRepaymentAdaptorSignature(goCtx context.Context, msg *types.MsgSubmitRepaymentAdaptorSignature) (*types.MsgSubmitRepaymentAdaptorSignatureResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanDoesNotExist
	}

	if !m.HasRepayment(ctx, msg.LoanId) {
		return nil, types.ErrInvalidRepayment
	}

	repayment := m.GetRepayment(ctx, msg.LoanId)
	if len(repayment.DcaAdaptorSignature) != 0 {
		return nil, types.ErrRepaymentAdaptorSigAlreadyExists
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	adaptorSigBytes, _ := hex.DecodeString(msg.AdaptorSignature)
	adaptorPointBytes, _ := hex.DecodeString(repayment.RepayAdaptorPoint)
	pubKeyBytes, _ := hex.DecodeString(loan.Agency)

	// TODO: calculate sig hash
	sigHash := []byte{}

	if !adaptor.Verify(adaptorSigBytes, sigHash, pubKeyBytes, adaptorPointBytes) {
		return nil, types.ErrInvalidAdaptorSignature
	}

	repayment.DcaAdaptorSignature = msg.AdaptorSignature
	m.SetRepayment(ctx, repayment)

	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("loan_id", msg.LoanId),
		sdk.NewAttribute("adaptor_signature", msg.AdaptorSignature),
	)

	return &types.MsgSubmitRepaymentAdaptorSignatureResponse{}, nil
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
	if loan.Status != types.LoanStatus_Liquidate {
		return nil, types.ErrLoanNotLiquidated
	}

	dlcMeta := m.GetDLCMeta(ctx, msg.LoanId)
	if len(dlcMeta.LiquidationAgencySignatures) > 0 {
		return nil, types.ErrLiquidationSignaturesAlreadyExist
	}

	liquidationCet, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet)), true)

	if len(msg.Signatures) != len(liquidationCet.Inputs) {
		return nil, errorsmod.Wrap(types.ErrInvalidLiquidationSignatures, "incorrect signature number")
	}

	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	for i := range liquidationCet.Inputs {
		sigBytes, _ := hex.DecodeString(msg.Signatures[i])

		// TODO: calculate sig hash
		sigHash := []byte{}

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

	if !m.HasRepayment(ctx, msg.LoanId) {
		return nil, types.ErrInvalidRepayment
	}

	repayment := m.GetRepayment(ctx, msg.LoanId)
	if len(repayment.DcaAdaptorSignature) == 0 {
		return nil, types.ErrRepaymentAdaptorSigDoesNotExist
	}

	sigBytes, _ := hex.DecodeString(msg.Signature)
	adaptorSigBytes, _ := hex.DecodeString(repayment.DcaAdaptorSignature)

	// extract secret from signature
	secret := adaptor.Extract(sigBytes, adaptorSigBytes)
	if len(secret) == 0 {
		return nil, types.ErrInvalidSignature
	}

	if types.AdaptorPoint(secret) != repayment.RepayAdaptorPoint {
		return nil, types.ErrInvalidRepaymentSecret
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	amount := loan.BorrowAmount.Amount.Add(loan.Interests).Add(loan.Fees)
	if err := m.bankKeeper.SendCoinsFromModuleToModule(ctx, types.RepaymentEscrowAccount, types.ModuleName, sdk.NewCoins(sdk.NewCoin(loan.BorrowAmount.Denom, amount))); err != nil {
		return nil, err
	}

	loan.Status = types.LoanStatus_Close
	m.SetLoan(ctx, loan)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("loan_id", loan.VaultAddress),
		sdk.NewAttribute("payment_secret", hex.EncodeToString(secret)),
	)

	return &types.MsgCloseResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
