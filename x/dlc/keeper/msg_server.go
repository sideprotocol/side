package keeper

import (
	"context"
	"fmt"
	"strings"

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

	if err := m.Keeper.HandleNonce(ctx, msg.Sender, msg.EventType, msg.Nonce, msg.OraclePubkey, msg.Signature); err != nil {
		return nil, err
	}

	return &types.MsgSubmitNonceResponse{}, nil
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
			sdk.NewAttribute(types.AttributeKeyParticipants, strings.Join(oracle.Participants, types.AttributeValueSeparator)),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", oracle.Threshold)),
			sdk.NewAttribute(types.AttributeKeyExpirationTime, oracle.Time.Add(m.DKGTimeoutPeriod(ctx)).String()),
		),
	)

	return &types.MsgCreateOracleResponse{}, nil
}

// CreateDCM implements types.MsgServer.
func (m msgServer) CreateDCM(goCtx context.Context, msg *types.MsgCreateDCM) (*types.MsgCreateDCMResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	m.tssKeeper.InitiateDKG(ctx, types.ModuleName, types.DCM_TYPE, int32(types.DKGIntent_DKG_INTENT_DCM), msg.Participants, msg.Threshold, 1)

	return &types.MsgCreateDCMResponse{}, nil
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
