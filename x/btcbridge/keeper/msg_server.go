package keeper

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcutil/psbt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

type msgServer struct {
	Keeper
}

// SubmitBlockHeaders implements types.MsgServer.
func (m msgServer) SubmitBlockHeaders(goCtx context.Context, msg *types.MsgSubmitBlockHeaders) (*types.MsgSubmitBlockHeadersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// Set block headers
	err := m.SetBlockHeaders(ctx, msg.BlockHeaders)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitBlockHeadersResponse{}, nil
}

// UpdateNonBtcRelayers implements types.MsgServer.
func (m msgServer) UpdateNonBtcRelayers(goCtx context.Context, msg *types.MsgUpdateNonBtcRelayers) (*types.MsgUpdateNonBtcRelayersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.IsTrustedNonBtcRelayer(ctx, msg.Sender) {
		return nil, types.ErrUntrustedNonBtcRelayer
	}

	// update non-btc relayers
	params := m.GetParams(ctx)
	params.NonBtcRelayers = msg.Relayers
	m.SetParams(ctx, params)

	return &types.MsgUpdateNonBtcRelayersResponse{}, nil
}

// SubmitDepositTransaction implements types.MsgServer.
// No Permission check required for this message
// Since everyone can submit a transaction to mint voucher tokens
// This message is usually sent by relayers
func (m msgServer) SubmitDepositTransaction(goCtx context.Context, msg *types.MsgSubmitDepositTransaction) (*types.MsgSubmitDepositTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("Error validating basic", "error", err)
		return nil, err
	}

	if !m.DepositEnabled(ctx) {
		return nil, types.ErrDepositNotEnabled
	}

	txHash, recipient, err := m.ProcessBitcoinDepositTransaction(ctx, msg)
	if err != nil {
		ctx.Logger().Error("Error processing bitcoin deposit transaction", "error", err)
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("blockhash", msg.Blockhash),
		sdk.NewAttribute("txid", txHash.String()),
		sdk.NewAttribute("recipient", recipient.EncodeAddress()),
	)

	return &types.MsgSubmitDepositTransactionResponse{}, nil
}

// SubmitWithdrawTransaction implements types.MsgServer.
// No Permission check required for this message
// This message is usually sent by relayers
func (m msgServer) SubmitWithdrawTransaction(goCtx context.Context, msg *types.MsgSubmitWithdrawTransaction) (*types.MsgSubmitWithdrawTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		ctx.Logger().Error("Error validating basic", "error", err)
		return nil, err
	}

	txHash, err := m.ProcessBitcoinWithdrawTransaction(ctx, msg)
	if err != nil {
		ctx.Logger().Error("Error processing bitcoin withdraw transaction", "error", err)
		return nil, err
	}

	// Emit Events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("blockhash", msg.Blockhash),
		sdk.NewAttribute("txid", txHash.String()),
	)

	return &types.MsgSubmitWithdrawTransactionResponse{}, nil
}

// WithdrawToBitcoin withdraws the asset to the bitcoin chain.
func (m msgServer) WithdrawToBitcoin(goCtx context.Context, msg *types.MsgWithdrawToBitcoin) (*types.MsgWithdrawToBitcoinResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.WithdrawEnabled(ctx) {
		return nil, types.ErrWithdrawNotEnabled
	}

	sender := sdk.MustAccAddressFromBech32(msg.Sender)

	amount, err := sdk.ParseCoinNormalized(msg.Amount)
	if err != nil {
		return nil, err
	}

	if m.ProtocolWithdrawFeeEnabled(ctx) {
		// deduct the protocol fee and get the actual withdrawal amount
		amount, err = m.handleWithdrawProtocolFee(ctx, sender, amount)
		if err != nil {
			return nil, err
		}
	}

	feeRate, _ := strconv.ParseInt(msg.FeeRate, 10, 64)

	req, err := m.Keeper.NewWithdrawRequest(ctx, msg.Sender, amount, feeRate)
	if err != nil {
		return nil, err
	}

	// lock assets
	if err := m.LockAssets(ctx, req, amount); err != nil {
		return nil, err
	}

	// Emit events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("amount", msg.Amount),
		sdk.NewAttribute("txid", req.Txid),
	)

	return &types.MsgWithdrawToBitcoinResponse{}, nil
}

// SubmitWithdrawSignatures submits the signatures of the withdrawal transaction.
func (m msgServer) SubmitWithdrawSignatures(goCtx context.Context, msg *types.MsgSubmitWithdrawSignatures) (*types.MsgSubmitWithdrawSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasWithdrawRequestByTxHash(ctx, msg.Txid) {
		return nil, types.ErrWithdrawRequestNotExist
	}

	withdrawRequest := m.GetWithdrawRequestByTxHash(ctx, msg.Txid)
	if withdrawRequest.Status != types.WithdrawStatus_WITHDRAW_STATUS_CREATED {
		return nil, types.ErrInvalidSignatures
	}

	packet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(msg.Psbt)), true)
	if err != nil {
		return nil, types.ErrInvalidSignatures
	}

	if packet.UnsignedTx.TxHash().String() != msg.Txid {
		return nil, types.ErrInvalidSignatures
	}

	if err = packet.SanityCheck(); err != nil {
		return nil, err
	}

	if !packet.IsComplete() {
		return nil, types.ErrInvalidSignatures
	}

	// verify the signatures
	if !types.VerifyPsbtSignatures(packet) {
		return nil, types.ErrInvalidSignatures
	}

	// set the withdraw status to broadcasted
	withdrawRequest.Psbt = msg.Psbt
	withdrawRequest.Status = types.WithdrawStatus_WITHDRAW_STATUS_BROADCASTED

	m.SetWithdrawRequest(ctx, withdrawRequest)

	return &types.MsgSubmitWithdrawSignaturesResponse{}, nil
}

// InitiateDKG initiates the DKG request.
func (m msgServer) InitiateDKG(goCtx context.Context, msg *types.MsgInitiateDKG) (*types.MsgInitiateDKGResponse, error) {
	if m.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	req, err := m.Keeper.InitiateDKG(ctx, msg.Participants, msg.Threshold, msg.VaultTypes)
	if err != nil {
		return nil, err
	}

	// Emit events
	m.EmitEvent(ctx, msg.Authority,
		sdk.NewAttribute("id", fmt.Sprintf("%d", req.Id)),
		sdk.NewAttribute("expiration", req.Expiration.String()),
	)

	return &types.MsgInitiateDKGResponse{}, nil
}

// CompleteDKG completes the DKG request by the DKG participant
func (m msgServer) CompleteDKG(goCtx context.Context, msg *types.MsgCompleteDKG) (*types.MsgCompleteDKGResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	req := &types.DKGCompletionRequest{
		Id:               msg.Id,
		Sender:           msg.Sender,
		Vaults:           msg.Vaults,
		ConsensusAddress: msg.ConsensusAddress,
		Signature:        msg.Signature,
	}

	if err := m.Keeper.CompleteDKG(ctx, req); err != nil {
		return nil, err
	}

	// Emit events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("id", fmt.Sprintf("%d", msg.Id)),
	)

	return &types.MsgCompleteDKGResponse{}, nil
}

// UpdateParams updates the module params.
func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	m.SetParams(ctx, msg.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
