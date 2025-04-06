package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/tss/types"
)

type msgServer struct {
	Keeper
}

// CompleteDKG completes the DKG request by the DKG participant
func (m msgServer) CompleteDKG(goCtx context.Context, msg *types.MsgCompleteDKG) (*types.MsgCompleteDKGResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CompleteDKG(ctx, msg.Sender, msg.Id, msg.PubKeys, msg.ConsensusPubkey, msg.Signature); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompleteDKG,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", msg.Id)),
			sdk.NewAttribute(types.AttributeKeyParticipant, msg.ConsensusPubkey),
		),
	)

	return &types.MsgCompleteDKGResponse{}, nil
}

// SubmitSignatures submits the signatures for the specified signing request
func (m msgServer) SubmitSignatures(goCtx context.Context, msg *types.MsgSubmitSignatures) (*types.MsgSubmitSignaturesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !m.HasSigningRequest(ctx, msg.Id) {
		return nil, types.ErrSigningRequestDoesNotExist
	}

	req := m.GetSigningRequest(ctx, msg.Id)
	if req.Status != types.SigningStatus_SIGNING_STATUS_PENDING {
		return nil, errorsmod.Wrap(types.ErrInvalidSigningStatus, "signing request non pending")
	}

	if err := m.GetSigningRequestCompletedHandler(req.Module)(ctx, msg.Sender, req.Id, req.ScopedId, req.Type, req.Intent, req.PubKey, msg.Signatures); err != nil {
		return nil, err
	}

	req.Status = types.SigningStatus_SIGNING_STATUS_SIGNED
	m.SetSigningRequest(ctx, req)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompleteSigning,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", msg.Id)),
		),
	)

	return &types.MsgSubmitSignaturesResponse{}, nil
}

// UpdateParams updates the module params
func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
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
