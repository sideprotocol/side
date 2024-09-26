package keeper

import (
	"bytes"
	"context"
	"fmt"

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

// UpdateTrustedNonBtcRelayers implements types.MsgServer.
func (m msgServer) UpdateTrustedNonBtcRelayers(goCtx context.Context, msg *types.MsgUpdateTrustedNonBtcRelayers) (*types.MsgUpdateTrustedNonBtcRelayersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.IsTrustedNonBtcRelayer(ctx, msg.Sender) {
		return nil, types.ErrUntrustedNonBtcRelayer
	}

	// update non-btc relayers
	params := m.GetParams(ctx)
	params.TrustedNonBtcRelayers = msg.Relayers
	m.SetParams(ctx, params)

	return &types.MsgUpdateTrustedNonBtcRelayersResponse{}, nil
}

// UpdateTrustedOracles implements types.MsgServer.
func (m msgServer) UpdateTrustedOracles(goCtx context.Context, msg *types.MsgUpdateTrustedOracles) (*types.MsgUpdateTrustedOraclesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.IsTrustedOracle(ctx, msg.Sender) {
		return nil, types.ErruntrustedOracle
	}

	// update oracles
	params := m.GetParams(ctx)
	params.TrustedOracles = msg.Oracles
	m.SetParams(ctx, params)

	return &types.MsgUpdateTrustedOraclesResponse{}, nil
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

// SubmitFeeRate submits the bitcoin network fee rate
func (m msgServer) SubmitFeeRate(goCtx context.Context, msg *types.MsgSubmitFeeRate) (*types.MsgSubmitFeeRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if !m.IsTrustedOracle(ctx, msg.Sender) {
		return nil, types.ErruntrustedOracle
	}

	m.SetFeeRate(ctx, msg.FeeRate)

	// Emit Events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("fee_rate", fmt.Sprintf("%d", msg.FeeRate)),
	)

	return &types.MsgSubmitFeeRateResponse{}, nil
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

	withdrawRequest, err := m.HandleWithdrawal(ctx, msg.Sender, amount)
	if err != nil {
		return nil, err
	}

	// Emit events
	m.EmitEvent(ctx, msg.Sender,
		sdk.NewAttribute("amount", amount.String()),
		sdk.NewAttribute("sequence", fmt.Sprintf("%d", withdrawRequest.Sequence)),
		sdk.NewAttribute("txid", withdrawRequest.Txid),
	)

	return &types.MsgWithdrawToBitcoinResponse{}, nil
}

// SubmitSignatures submits the signatures of the signing request.
func (m msgServer) SubmitSignatures(goCtx context.Context, msg *types.MsgSubmitSignatures) (*types.MsgSubmitSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasSigningRequestByTxHash(ctx, msg.Txid) {
		return nil, types.ErrSigningRequestNotExist
	}

	signingRequest := m.GetSigningRequestByTxHash(ctx, msg.Txid)
	if signingRequest.Status != types.SigningStatus_SIGNING_STATUS_PENDING {
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

	// set the signing request status to broadcasted
	signingRequest.Psbt = msg.Psbt
	signingRequest.Status = types.SigningStatus_SIGNING_STATUS_BROADCASTED

	m.SetSigningRequest(ctx, signingRequest)

	return &types.MsgSubmitSignaturesResponse{}, nil
}

// ConsolidateVaults performs the UTXO consolidation for the given vaults.
func (m msgServer) ConsolidateVaults(goCtx context.Context, msg *types.MsgConsolidateVaults) (*types.MsgConsolidateVaultsResponse, error) {
	if m.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.ConsolidateVaults(ctx, msg.VaultVersion, msg.BtcConsolidation, msg.RunesConsolidations, msg.FeeRate); err != nil {
		return nil, err
	}

	return &types.MsgConsolidateVaultsResponse{}, nil
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

	req, err := m.Keeper.InitiateDKG(ctx, msg.Participants, msg.Threshold, msg.VaultTypes, msg.DisableBridge, msg.EnableTransfer, msg.TargetUtxoNum, msg.FeeRate)
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

// TransferVault performs the vault asset transfer from the source version to the destination version
func (m msgServer) TransferVault(goCtx context.Context, msg *types.MsgTransferVault) (*types.MsgTransferVaultResponse, error) {
	if m.authority != msg.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.TransferVault(ctx, msg.SourceVersion, msg.DestVersion, msg.AssetType, msg.Psbts, msg.TargetUtxoNum, msg.FeeRate); err != nil {
		return nil, err
	}

	// Emit events
	m.EmitEvent(ctx, msg.Authority,
		sdk.NewAttribute("source_version", fmt.Sprintf("%d", msg.SourceVersion)),
		sdk.NewAttribute("dest_version", fmt.Sprintf("%d", msg.DestVersion)),
		sdk.NewAttribute("asset_type", msg.AssetType.String()),
	)

	return &types.MsgTransferVaultResponse{}, nil
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
