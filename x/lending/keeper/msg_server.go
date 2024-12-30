package keeper

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcutil/psbt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dlc "github.com/sideprotocol/side/x/dlc/types"
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

	event := dlc.DLCPriceEvent{} // need to integrate with dlc module
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

	cetBytes, err := base64.StdEncoding.DecodeString(msg.Cets)
	if err != nil {
		return nil, types.ErrInvalidCET
	}
	cet, err := psbt.NewFromRawBytes(bytes.NewReader(cetBytes), true)
	if err != nil {
		return nil, types.ErrInvalidCET
	}

	if e := types.VerifyCET(fundTx, cet); e != nil {
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
		Cets:             msg.Cets,
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
		sdk.NewAttribute("event_id", loan.EventId),
	)

	return &types.MsgApplyResponse{}, nil

}

// Approve implements types.MsgServer.
func (m msgServer) Approve(goCtx context.Context, msg *types.MsgApprove) (*types.MsgApproveResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasDepositLog(ctx, msg.DepositTxId) {
		return nil, types.ErrDepositTxNotExists
	}

	// verify merkle proof
	// verify(msg.Height, msg.DepositTxId, msg.Poof)

	log := m.GetDepositLog(ctx, msg.DepositTxId)
	if !m.HasLoan(ctx, log.VaultAddress) {
		return nil, types.ErrLoanNotExists
	}
	loan := m.GetLoan(ctx, log.VaultAddress)

	loan.Status = types.LoanStatus_Approve
	m.SetLoan(ctx, loan)

	m.EmitEvent(ctx, msg.Relayer,
		sdk.NewAttribute("vault", loan.VaultAddress),
		sdk.NewAttribute("deposit_tx", msg.DepositTxId),
		sdk.NewAttribute("proof", msg.Poof),
		sdk.NewAttribute("height", string(msg.Height)),
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

	if types.HashLoanSecret(msg.LoanSecret) != loan.HashLoanSecret {
		return nil, types.ErrMismatchLoanSecret
	}

	m.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, borrower, *loan.BorrowAmount)

	loan.Status = types.LoanStatus_Disburse

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

	m.EmitEvent(ctx, msg.Borrower,
		sdk.NewAttribute("loan_id", loan.VaultAddress),
		sdk.NewAttribute("adaptor_point", msg.AdaptorPoint),
		sdk.NewAttribute("txid", loan.BorrowAmount.String()),
	)

	return &types.MsgRepayResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
