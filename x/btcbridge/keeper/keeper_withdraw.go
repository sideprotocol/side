package keeper

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetRequestSeqence returns the request sequence
func (k Keeper) GetRequestSeqence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SequenceKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// IncrementRequestSequence increments the request sequence and returns the new sequence
func (k Keeper) IncrementRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	seq := k.GetRequestSeqence(ctx) + 1
	store.Set(types.SequenceKey, sdk.Uint64ToBigEndian(seq))
	return seq
}

// NewWithdrawRequest creates a new withdrawal request
func (k Keeper) NewWithdrawRequest(ctx sdk.Context, sender string, amount sdk.Coin) (*types.BitcoinWithdrawRequest, error) {
	switch types.AssetTypeFromDenom(amount.Denom, k.GetParams(ctx)) {
	case types.AssetType_ASSET_TYPE_BTC, types.AssetType_ASSET_TYPE_RUNE:
		withdrawRequest := &types.BitcoinWithdrawRequest{
			Address:  sender,
			Amount:   amount,
			Sequence: k.IncrementRequestSequence(ctx),
			Status:   types.WithdrawStatus_WITHDRAW_STATUS_CREATED,
		}

		k.SetWithdrawRequest(ctx, withdrawRequest)

		return withdrawRequest, nil

	default:
		return nil, types.ErrAssetNotSupported
	}
}

// HasWithdrawRequest returns true if the given withdrawal request exists, false otherwise
func (k Keeper) HasWithdrawRequest(ctx sdk.Context, sequence uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcWithdrawRequestKey(sequence))
}

// HasWithdrawRequestByTxHash returns true if the given withdrawal request exists, false otherwise
func (k Keeper) HasWithdrawRequestByTxHash(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcWithdrawRequestByTxHashKey(hash))
}

// GetWithdrawRequest returns the withdraw request by the given sequence
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, sequence uint64) *types.BitcoinWithdrawRequest {
	store := ctx.KVStore(k.storeKey)

	var withdrawRequest types.BitcoinWithdrawRequest
	bz := store.Get(types.BtcWithdrawRequestKey(sequence))
	k.cdc.MustUnmarshal(bz, &withdrawRequest)

	return &withdrawRequest
}

// GetWithdrawRequestByTxHash returns the withdraw request by the given hash
func (k Keeper) GetWithdrawRequestByTxHash(ctx sdk.Context, hash string) *types.BitcoinWithdrawRequest {
	store := ctx.KVStore(k.storeKey)

	var withdrawRequest types.BitcoinWithdrawRequest
	bz := store.Get(types.BtcWithdrawRequestByTxHashKey(hash))
	k.cdc.MustUnmarshal(bz, &withdrawRequest)

	return &withdrawRequest
}

// SetWithdrawRequest sets the withdrawal request
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, withdrawRequest *types.BitcoinWithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(withdrawRequest)
	store.Set(types.BtcWithdrawRequestKey(withdrawRequest.Sequence), bz)

	if len(withdrawRequest.Txid) != 0 {
		store.Set(types.BtcWithdrawRequestByTxHashKey(withdrawRequest.Txid), types.Int64ToBytes(withdrawRequest.Sequence))
	}
}

// IterateWithdrawRequests iterates through all withdrawal requests
func (k Keeper) IterateWithdrawRequests(ctx sdk.Context, process func(withdrawRequest types.BitcoinWithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcWithdrawRequestPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawRequest types.BitcoinWithdrawRequest
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawRequest)

		if process(withdrawRequest) {
			break
		}
	}
}

// filter WithdrawRequest by status with pagination
func (k Keeper) FilterWithdrawRequestsByStatus(ctx sdk.Context, req *types.QueryWithdrawRequestsRequest) []*types.BitcoinWithdrawRequest {
	var withdrawRequests []*types.BitcoinWithdrawRequest

	k.IterateWithdrawRequests(ctx, func(withdrawRequest types.BitcoinWithdrawRequest) (stop bool) {
		if withdrawRequest.Status == req.Status {
			withdrawRequests = append(withdrawRequests, &withdrawRequest)
		}

		// pagination TODO: limit the number of withdrawal requests
		if len(withdrawRequests) >= 100 {
			return true
		}

		return false
	})

	return withdrawRequests
}

// filter WithdrawRequest by address with pagination
func (k Keeper) FilterWithdrawRequestsByAddr(ctx sdk.Context, req *types.QueryWithdrawRequestsByAddressRequest) []*types.BitcoinWithdrawRequest {
	var withdrawRequests []*types.BitcoinWithdrawRequest
	k.IterateWithdrawRequests(ctx, func(withdrawRequest types.BitcoinWithdrawRequest) (stop bool) {
		if withdrawRequest.Address == req.Address {
			withdrawRequests = append(withdrawRequests, &withdrawRequest)
		}

		// pagination TODO: limit the number of withdrawal requests
		if len(withdrawRequests) >= 100 {
			return true
		}

		return false
	})

	return withdrawRequests
}

// Process Bitcoin Withdraw Transaction
func (k Keeper) ProcessBitcoinWithdrawTransaction(ctx sdk.Context, msg *types.MsgSubmitWithdrawTransaction) (*chainhash.Hash, error) {
	ctx.Logger().Info("accept bitcoin withdraw tx", "blockhash", msg.Blockhash)

	param := k.GetParams(ctx)

	header := k.GetBlockHeader(ctx, msg.Blockhash)
	// Check if block confirmed
	if header == nil {
		return nil, types.ErrBlockNotFound
	}

	best := k.GetBestBlockHeader(ctx)
	// Check if the block is confirmed
	if best.Height-header.Height < uint64(param.Confirmations) {
		return nil, types.ErrNotConfirmed
	}
	// Check if the block is within the acceptable depth
	if best.Height-header.Height > param.MaxAcceptableBlockDepth {
		return nil, types.ErrExceedMaxAcceptanceDepth
	}

	// Decode the base64 transaction
	txBytes, err := base64.StdEncoding.DecodeString(msg.TxBytes)
	if err != nil {
		fmt.Println("Error decoding transaction from base64:", err)
		return nil, err
	}

	// Create a new transaction
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		fmt.Println("Error deserializing transaction:", err)
		return nil, err
	}

	uTx := btcutil.NewTx(&tx)
	if len(uTx.MsgTx().TxIn) < 1 {
		return nil, types.ErrInvalidBtcTransaction
	}

	txHash := uTx.MsgTx().TxHash()

	if !k.HasWithdrawRequestByTxHash(ctx, txHash.String()) {
		return nil, types.ErrWithdrawRequestNotExist
	}

	withdrawRequest := k.GetWithdrawRequestByTxHash(ctx, txHash.String())
	// if withdrawRequest.Status != types.WithdrawStatus_WITHDRAW_STATUS_BROADCASTED {
	// 	return types.ErrInvalidStatus
	// }
	withdrawRequest.Status = types.WithdrawStatus_WITHDRAW_STATUS_CONFIRMED
	k.SetWithdrawRequest(ctx, withdrawRequest)

	// Validate the transaction
	if err := blockchain.CheckTransactionSanity(uTx); err != nil {
		fmt.Println("Transaction is not valid:", err)
		return nil, err
	}

	return &txHash, nil
}
