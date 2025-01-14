package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil/psbt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/x/lending/types"
)

type msgServer struct {
	Keeper
}

// CreateLoan implements types.MsgServer.
func (m msgServer) Apply(goCtx context.Context, msg *types.MsgApply) (*types.MsgApplyResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	// _borrower, errb := sdk.AccAddressFromBech32(msg.Borrower)
	// if errb != nil {
	// 	return nil, errb
	// }

	if !m.dlcKeeper.HasEvent(ctx, msg.EventId) {
		return nil, types.ErrInvalidPriceEvent
	}

	event := m.dlcKeeper.GetEvent(ctx, msg.EventId)
	if event.HasTriggered {
		return nil, types.ErrInvalidPriceEvent
	}

	vault, err := types.CreateVaultAddress(msg.BorrowerPubkey, event.Pubkey, msg.LoanSecretHash, msg.MaturityTime, msg.FinalTimeout)
	if err != nil {
		return nil, err
	}

	if m.HasLoan(ctx, vault) {
		return nil, types.ErrDuplicatedVault
	}

	fundBytes, err := base64.StdEncoding.DecodeString(msg.DepositTx)
	if err != nil {
		return nil, types.ErrInvalidFunding
	}

	fundTx, err := psbt.NewFromRawBytes(bytes.NewReader(fundBytes), true)
	if err != nil {
		return nil, types.ErrInvalidFunding
	}
	depositTxid := fundTx.UnsignedTx.TxHash().String()

	if e := types.VerifyCETs(fundTx, msg.Cets); e != nil {
		return nil, e
	}

	params := m.GetParams(ctx)

	collateralAmount := math.NewInt(0)
	for _, o := range fundTx.UnsignedTx.TxOut {
		address, e := types.GetTaprootAddress(o.PkScript)
		if e != nil {
			return nil, e
		}
		if address.EncodeAddress() == vault {
			collateralAmount.Add(math.NewInt(o.Value))
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
		Agency:           event.Pubkey,
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

	// set CETs
	m.SetCETs(ctx, loan.VaultAddress, msg.Cets)

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

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasDepositLog(ctx, msg.DepositTxId) {
		return nil, types.ErrDepositTxNotExists
	}

	log := m.GetDepositLog(ctx, msg.DepositTxId)
	if !m.HasLoan(ctx, log.VaultAddress) {
		return nil, types.ErrLoanNotExists
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, types.ErrInvalidSender
	}

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanNotExists
	}

	loan := m.GetLoan(ctx, msg.LoanId)

	if msg.Borrower != loan.Borrower {
		return nil, types.ErrMismatchedBorrower
	}

	if types.HashLoanSecret(msg.LoanSecret) != loan.HashLoanSecret {
		return nil, types.ErrMismatchLoanSecret
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, types.ErrInvalidSender
	}

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanNotExists
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

	claimTx, err := types.CreateRepaymentTransaction(dls)
	if err != nil {
		return nil, err
	}
	tx, err := claimTx.B64Encode()
	if err != nil {
		return nil, err
	}

	repayment := types.Repayment{
		LoanId:            msg.LoanId,
		Txid:              claimTx.UnsignedTx.TxHash().String(),
		Tx:                tx,
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
		return nil, types.ErrLoanNotExists
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

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("loan_id", msg.LoanId),
		sdk.NewAttribute("adaptor_signature", msg.AdaptorSignature),
	)

	return &types.MsgSubmitRepaymentAdaptorSignatureResponse{}, nil
}

// Close implements types.MsgServer.
func (m msgServer) Close(goCtx context.Context, msg *types.MsgClose) (*types.MsgCloseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasLoan(ctx, msg.LoanId) {
		return nil, types.ErrLoanNotExists
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
