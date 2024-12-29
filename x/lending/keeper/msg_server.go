package keeper

import (
	"bytes"
	"context"
	"encoding/base64"

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

	loan := types.Loan{
		VaultAddress:   vault,
		Borrower:       msg.Borrower,
		Agency:         event.Pubkey,
		HashLoanSecret: msg.LoanSecretHash,
		MaturityTime:   msg.MaturityTime,
		FinalTimeout:   msg.FinalTimeout,
		BorrowAmount:   msg.BorrowAmount,
		// Fees:             sdk.NewCoin("xx", math.NewInt(0)),
		CollateralAmount: math.NewInt(0),
		InterestRate:     math.NewInt(int64(params.LendingRate)),
		EventId:          msg.EventId,
		Cets:             msg.Cets,
		DepositTx:        msg.DepositTx,
		CreateAt:         ctx.BlockTime(),
		PoolId:           msg.PoolId,
		Status:           types.LoanStatus_Apply,
	}

	m.SetLoan(ctx, loan)

	return &types.MsgApplyResponse{}, nil

}

// Repay implements types.MsgServer.
func (m msgServer) Repay(ctx context.Context, msg *types.MsgRepay) (*types.MsgRepayResponse, error) {
	panic("unimplemented")
}

// RequestVaultAddress implements types.MsgServer.
func (m msgServer) Redeem(ctx context.Context, msg *types.MsgRedeem) (*types.MsgRedeemResponse, error) {
	panic("unimplemented")
}

// SubmitFundingTx implements types.MsgServer.
func (m msgServer) Deposit(ctx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	panic("unimplemented")
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
