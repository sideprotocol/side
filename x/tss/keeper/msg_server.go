package keeper

import (
	"context"
	"fmt"
	"slices"

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

	dkgRequest := m.GetDKGRequest(ctx, msg.Id)

	// callback to the module handler
	if err := m.GetDKGCompletionReceivedHandler(dkgRequest.Module)(ctx, dkgRequest.Id, dkgRequest.Type, dkgRequest.Intent, msg.ConsensusPubkey); err != nil {
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

	// verify signatures
	if err := m.VerifySignatures(ctx, req, msg.Signatures); err != nil {
		return nil, err
	}

	// callback to the module handler
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

// Refresh refreshes the key shares
func (m msgServer) Refresh(goCtx context.Context, msg *types.MsgRefresh) (*types.MsgRefreshResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	for i, dkgId := range msg.DkgIds {
		if !m.HasDKGRequest(ctx, dkgId) {
			return nil, errorsmod.Wrapf(types.ErrDKGRequestDoesNotExist, "dkg %d", dkgId)
		}

		dkgRequest := m.GetDKGRequest(ctx, dkgId)
		if dkgRequest.Status != types.DKGStatus_DKG_STATUS_COMPLETED {
			return nil, errorsmod.Wrapf(types.ErrInvalidDKGStatus, "dkg %d not completed", dkgId)
		}

		remainingParticipantNum := len(dkgRequest.Participants) - len(msg.RemovedParticipants)
		if remainingParticipantNum < types.MinDKGParticipantNum {
			return nil, errorsmod.Wrapf(types.ErrInvalidParticipants, "remaining participants %d cannot be less than min participants %d", remainingParticipantNum, types.MinDKGParticipantNum)
		}

		for _, p := range msg.RemovedParticipants {
			if !slices.Contains(dkgRequest.Participants, p) {
				return nil, errorsmod.Wrapf(types.ErrInvalidParticipants, "participant %s does not exist for dkg %d", p, dkgId)
			}
		}

		if msg.Thresholds[i] > uint32(remainingParticipantNum) {
			return nil, errorsmod.Wrapf(types.ErrInvalidThresholds, "threshold %d cannot be greater than participants %d for dkg %d", msg.Thresholds[i], remainingParticipantNum, dkgId)
		}

		m.InitiateRefreshingRequest(ctx, dkgId, msg.RemovedParticipants, msg.Thresholds[i], msg.TimeoutDuration)
	}

	return &types.MsgRefreshResponse{}, nil
}

// CompleteRefreshing completes the refreshing request by the participant
func (m msgServer) CompleteRefreshing(goCtx context.Context, msg *types.MsgCompleteRefreshing) (*types.MsgCompleteRefreshingResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CompleteRefreshing(ctx, msg.Sender, msg.Id, msg.ConsensusPubkey, msg.Signature); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCompleteRefreshing,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", msg.Id)),
			sdk.NewAttribute(types.AttributeKeyParticipant, msg.ConsensusPubkey),
		),
	)

	return &types.MsgCompleteRefreshingResponse{}, nil
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
