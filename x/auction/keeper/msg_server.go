package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/sideprotocol/side/x/auction/types"
)

type msgServer struct {
	Keeper
}

// Bid implements types.MsgServer.
func (m msgServer) Bid(goCtx context.Context, msg *types.MsgBid) (*types.MsgBidResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	bid, err := m.Keeper.HandleBid(ctx, msg.Sender, msg.AuctionId, msg.Price, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBid,
			sdk.NewAttribute(types.AttributeKeyBidId, fmt.Sprintf("%d", bid.Id)),
			sdk.NewAttribute(types.AttributeKeyBidder, bid.Bidder),
			sdk.NewAttribute(types.AttributeKeyAuctionId, fmt.Sprintf("%d", bid.AuctionId)),
			sdk.NewAttribute(types.AttributeKeyBidPrice, fmt.Sprintf("%d", bid.BidPrice)),
			sdk.NewAttribute(types.AttributeKeyBidAmount, bid.BidAmount.String()),
		),
	)

	return &types.MsgBidResponse{}, nil
}

// CancelBid implements types.MsgServer.
func (m msgServer) CancelBid(goCtx context.Context, msg *types.MsgCancelBid) (*types.MsgCancelBidResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.Keeper.CancelBid(ctx, msg.Sender, msg.Id); err != nil {
		return nil, err
	}

	return &types.MsgCancelBidResponse{}, nil
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
