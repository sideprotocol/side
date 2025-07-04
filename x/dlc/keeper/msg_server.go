package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/dlc/types"
)

type msgServer struct {
	Keeper
}

// CreateDCM implements types.MsgServer.
func (m msgServer) CreateDCM(goCtx context.Context, msg *types.MsgCreateDCM) (*types.MsgCreateDCMResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	baseParticipants := m.tssKeeper.AllowedDKGParticipants(ctx)
	if len(baseParticipants) != 0 {
		for _, p := range msg.Participants {
			if !slices.Contains(baseParticipants, p) {
				return nil, errorsmod.Wrap(types.ErrInvalidParticipants, "dcm participant not authorized")
			}
		}
	}

	m.tssKeeper.InitiateDKG(ctx, types.ModuleName, types.DKG_TYPE_DCM, int32(types.DKGIntent_DKG_INTENT_DEFAULT), msg.Participants, msg.Threshold, 1)

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

	// validate oracle participant allowlist
	if err := m.ValidateOracleParticipantAllowlist(ctx, msg.Params.AllowedOracleParticipants); err != nil {
		return nil, err
	}

	// update params
	m.SetParams(ctx, msg.Params)

	// update oracle participants liveness on params changed
	m.UpdateOracleParticipantsLiveness(ctx, msg.Params.AllowedOracleParticipants)

	return &types.MsgUpdateParamsResponse{}, nil
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
