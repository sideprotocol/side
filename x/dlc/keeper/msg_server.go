package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

type msgServer struct {
	Keeper
}

// SubmitNonce implements types.MsgServer.
func (m msgServer) SubmitNonce(goCtx context.Context, msg *types.MsgSubmitNonce) (*types.MsgSubmitNonceResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.HandleNonce(ctx, msg.Sender, msg.Nonce, msg.OraclePubkey, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgSubmitNonceResponse{}, nil
}

// SubmitAttestation implements types.MsgServer.
func (m msgServer) SubmitAttestation(goCtx context.Context, msg *types.MsgSubmitAttestation) (*types.MsgSubmitAttestationResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.HandleAttestation(ctx, msg.Sender, msg.EventId, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgSubmitAttestationResponse{}, nil
}

// SubmitOraclePubKey implements types.MsgServer.
func (m msgServer) SubmitOraclePubKey(goCtx context.Context, msg *types.MsgSubmitOraclePubKey) (*types.MsgSubmitOraclePubKeyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.SubmitOraclePubKey(ctx, msg.Sender, msg.PubKey, msg.OracleId, msg.OraclePubkey, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgSubmitOraclePubKeyResponse{}, nil
}

// SubmitAgencyPubKey implements types.MsgServer.
func (m msgServer) SubmitAgencyPubKey(goCtx context.Context, msg *types.MsgSubmitAgencyPubKey) (*types.MsgSubmitAgencyPubKeyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.SubmitAgencyPubKey(ctx, msg.Sender, msg.PubKey, msg.AgencyId, msg.AgencyPubkey, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgSubmitAgencyPubKeyResponse{}, nil
}

// CreateOracle implements types.MsgServer.
func (m msgServer) CreateOracle(goCtx context.Context, msg *types.MsgCreateOracle) (*types.MsgCreateOracleResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	oracle, err := m.Keeper.CreateOracle(ctx, msg.Participants, msg.Threshold)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateOracle,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", oracle.Id)),
			sdk.NewAttribute(types.AttributeKeyParticipants, fmt.Sprintf("%s", oracle.Participants)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, oracle.Time.Add(m.GetDKGTimeoutPeriod(ctx)).String()),
		),
	)

	return &types.MsgCreateOracleResponse{}, nil
}

// CreateAgency implements types.MsgServer.
func (m msgServer) CreateAgency(goCtx context.Context, msg *types.MsgCreateAgency) (*types.MsgCreateAgencyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	agency, err := m.Keeper.CreateAgency(ctx, msg.Participants, msg.Threshold)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeCreateAgency,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", agency.Id)),
			sdk.NewAttribute(types.AttributeKeyParticipants, fmt.Sprintf("%s", agency.Participants)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", agency.Threshold)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, agency.Time.Add(m.GetDKGTimeoutPeriod(ctx)).String()),
		),
	)

	return &types.MsgCreateAgencyResponse{}, nil
}

// UpdateParams updates the module params.
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
