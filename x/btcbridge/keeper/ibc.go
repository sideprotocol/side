package keeper

import (
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// IBCTransfer transfers the specified token via IBC
func (k Keeper) IBCTransfer(ctx sdk.Context, sender string, recipient string, token sdk.Coin, channelId string) error {
	msg := &transfertypes.MsgTransfer{
		SourcePort:       types.DefaultPortId,
		SourceChannel:    channelId,
		Token:            token,
		Sender:           sender,
		Receiver:         recipient,
		TimeoutHeight:    clienttypes.NewHeight(0, 0),
		TimeoutTimestamp: 0,
		Memo:             types.DefaultMemo,
	}

	if _, err := k.ibctransferKeeper.Transfer(ctx, msg); err != nil {
		return err
	}

	return nil
}

// AddToIBCWithdrawRequestQueue adds the given packet to IBC withdrawal queue for sBTC
func (k Keeper) AddToIBCWithdrawRequestQueue(ctx sdk.Context, sequence uint64, recipient string, amount int64) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&types.IBCWithdrawRequest{
		Sequence: sequence,
		Address:  recipient,
		Amount:   sdk.NewInt64Coin(k.GetParams(ctx).BtcVoucherDenom, amount).String(),
	})

	store.Set(types.IBCWithdrawRequestQueueKey(sequence), bz)
}

// RemoveFromIBCWithdrawRequestQueue removes the given IBC withdrawal request from the IBC withdrawal request queue
func (k Keeper) RemoveFromIBCWithdrawRequestQueue(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.IBCWithdrawRequestQueueKey(sequence))
}

// GetPendingIBCWithdrawRequests gets the pending IBC withdrawal requests up to the given maximum number
func (k Keeper) GetPendingIBCWithdrawRequests(ctx sdk.Context, maxNum uint32) []*types.IBCWithdrawRequest {
	requests := make([]*types.IBCWithdrawRequest, 0)

	k.IterateIBCWithdrawRequestQueue(ctx, func(req *types.IBCWithdrawRequest) (stop bool) {
		requests = append(requests, req)

		return maxNum != 0 && len(requests) >= int(maxNum)
	})

	return requests
}

// IterateIBCWithdrawRequestQueue iterates through the IBC withdrawal request queue
func (k Keeper) IterateIBCWithdrawRequestQueue(ctx sdk.Context, cb func(req *types.IBCWithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.IBCWithdrawRequestQueueKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var request types.IBCWithdrawRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)

		if cb(&request) {
			break
		}
	}
}

// CheckSBTCAutoPegOut returns true if the given packet is sBTC transfer and auto-pegout enabled, false otherwise
func (k Keeper) CheckSBTCAutoPegOut(ctx sdk.Context, packet transfertypes.FungibleTokenPacketData) bool {
	return packet.Denom == k.BtcDenom(ctx) && packet.Memo == types.FlagAutoPegOut
}

// IBCSendPacketCallback implements IBC callbacks
func (k Keeper) IBCSendPacketCallback(
	ctx sdk.Context,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	packetData []byte,
	contractAddress,
	packetSenderAddress string,
) error {
	// no-op
	return nil
}

// IBCOnAcknowledgementPacketCallback implements IBC callbacks
func (k Keeper) IBCOnAcknowledgementPacketCallback(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
	contractAddress,
	packetSenderAddress string,
) error {
	// no-op
	return nil
}

// IBCOnTimeoutPacketCallback implements IBC callbacks
func (k Keeper) IBCOnTimeoutPacketCallback(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	contractAddress,
	packetSenderAddress string,
) error {
	// no-op
	return nil
}

// IBCReceivePacketCallback implements IBC callbacks
func (k Keeper) IBCReceivePacketCallback(
	ctx sdk.Context,
	packet ibcexported.PacketI,
	ack ibcexported.Acknowledgement,
	contractAddress string,
) error {
	// check if the bridge withdrawal enabled
	if !k.WithdrawEnabled(ctx) {
		return nil
	}

	// parse the transfer packet
	tranferPacket, ok := tryGetTransferPacket(packet)
	if !ok || !k.CheckSBTCAutoPegOut(ctx, tranferPacket) {
		return nil
	}

	// check amount
	amount, ok := sdkmath.NewIntFromString(tranferPacket.Amount)
	if !ok || !amount.IsInt64() {
		return nil
	}

	// check if the recipient address is valid btc address
	if !types.IsValidBtcAddress(tranferPacket.Receiver) {
		return nil
	}

	// add to IBC withdrawal request queue
	k.AddToIBCWithdrawRequestQueue(ctx, packet.GetSequence(), tranferPacket.Receiver, amount.Int64())

	return nil
}

// tryGetTransferPacket attempts to parse the IBC transfer packet from the given packet
func tryGetTransferPacket(packet ibcexported.PacketI) (transfertypes.FungibleTokenPacketData, bool) {
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return data, false
	}

	return data, true
}
