package keeper

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

type (
	Keeper struct {
		BaseUTXOKeeper

		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		bankKeeper    types.BankKeeper
		stakingKeeper types.StakingKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		bankKeeper:     bankKeeper,
		stakingKeeper:  stakingKeeper,
		BaseUTXOKeeper: *NewBaseUTXOKeeper(cdc, storeKey),
		authority:      authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsStoreKey, bz)
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	var params types.Params
	bz := store.Get(types.ParamsStoreKey)
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

func (k Keeper) GetBestBlockHeader(ctx sdk.Context) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	var blockHeader types.BlockHeader
	bz := store.Get(types.BtcBestBlockHeaderKey)
	k.cdc.MustUnmarshal(bz, &blockHeader)
	return &blockHeader
}

func (k Keeper) SetBestBlockHeader(ctx sdk.Context, header *types.BlockHeader) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(header)
	store.Set(types.BtcBestBlockHeaderKey, bz)
}

func (k Keeper) SetBlockHeader(ctx sdk.Context, header *types.BlockHeader) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(header)

	store.Set(types.BtcBlockHeaderHashKey(header.Hash), bz)
	store.Set(types.BtcBlockHeaderHeightKey(header.Height), []byte(header.Hash))
}

func (k Keeper) SetBlockHeaders(ctx sdk.Context, headers []*types.BlockHeader) {
	for _, h := range headers {
		k.SetBlockHeader(ctx, h)
	}
}

func (k Keeper) InsertBlockHeaders(ctx sdk.Context, blockHeaders []*types.BlockHeader) error {
	store := ctx.KVStore(k.storeKey)

	startBlockHeader := blockHeaders[0]
	newBestBlockHeader := blockHeaders[len(blockHeaders)-1]

	// check if the starting block header already exists
	if store.Has(types.BtcBlockHeaderHashKey(startBlockHeader.Hash)) {
		// return no error
		return nil
	}

	params := k.GetParams(ctx)

	// get the best block header
	best := k.GetBestBlockHeader(ctx)

	if startBlockHeader.PreviousBlockHash == best.Hash {
		if startBlockHeader.Height != best.Height+1 {
			return errorsmod.Wrap(types.ErrInvalidBlockHeaders, "invalid block height")
		}
	} else {
		// reorg detected
		// check if the reorg depth exceeds the safe confirmations
		if best.Height-startBlockHeader.Height+1 > uint64(params.Confirmations) {
			return types.ErrInvalidReorgDepth
		}

		// check if the previous block exists
		if !store.Has(types.BtcBlockHeaderHashKey(startBlockHeader.PreviousBlockHash)) {
			return errorsmod.Wrap(types.ErrInvalidBlockHeaders, "previous block does not exist")
		}

		// check the block height
		prevBlock := k.GetBlockHeader(ctx, startBlockHeader.PreviousBlockHash)
		if startBlockHeader.Height != prevBlock.Height+1 {
			return errorsmod.Wrap(types.ErrInvalidBlockHeaders, "invalid block height")
		}

		// check if the new block headers has more work than the work accumulated from the forked block to the current tip
		totalWorkOldToTip := k.CalcTotalWork(ctx, startBlockHeader.Height, best.Height)
		totalWorkNew := types.BlockHeaders(blockHeaders).GetTotalWork()
		if sdk.GetConfig().GetBtcChainCfg().Net == wire.MainNet && totalWorkNew.Cmp(totalWorkOldToTip) <= 0 || totalWorkNew.Cmp(totalWorkOldToTip) < 0 {
			return errorsmod.Wrap(types.ErrInvalidBlockHeaders, "invalid forking block headers")
		}

		// remove the block headers starting from the forked block height
		for i := startBlockHeader.Height; i <= best.Height; i++ {
			ctx.Logger().Info("Removing block header: ", i)
			thash := k.GetBlockHashByHeight(ctx, i)
			store.Delete(types.BtcBlockHeaderHashKey(thash))
			store.Delete(types.BtcBlockHeaderHeightKey(i))
		}
	}

	// set block headers
	k.SetBlockHeaders(ctx, blockHeaders)

	// set the best block header
	k.SetBestBlockHeader(ctx, newBestBlockHeader)

	return nil
}

func (k Keeper) GetBlockHeader(ctx sdk.Context, hash string) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	var blockHeader types.BlockHeader
	bz := store.Get(types.BtcBlockHeaderHashKey(hash))
	k.cdc.MustUnmarshal(bz, &blockHeader)
	return &blockHeader
}

func (k Keeper) GetBlockHashByHeight(ctx sdk.Context, height uint64) string {
	store := ctx.KVStore(k.storeKey)
	hash := store.Get(types.BtcBlockHeaderHeightKey(height))
	return string(hash)
}

func (k Keeper) GetBlockHeaderByHeight(ctx sdk.Context, height uint64) *types.BlockHeader {
	store := ctx.KVStore(k.storeKey)
	hash := store.Get(types.BtcBlockHeaderHeightKey(height))
	return k.GetBlockHeader(ctx, string(hash))
}

// GetAllBlockHeaders returns all block headers
func (k Keeper) GetAllBlockHeaders(ctx sdk.Context) []*types.BlockHeader {
	var headers []*types.BlockHeader
	k.IterateBlockHeaders(ctx, func(header types.BlockHeader) (stop bool) {
		headers = append(headers, &header)
		return false
	})
	return headers
}

// IterateBlockHeaders iterates through all block headers
func (k Keeper) IterateBlockHeaders(ctx sdk.Context, process func(header types.BlockHeader) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.BtcBlockHeaderHashPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.BlockHeader
		k.cdc.MustUnmarshal(iterator.Value(), &header)
		if process(header) {
			break
		}
	}
}

// CalcTotalWork calculates the total work of the given range of block headers
func (k Keeper) CalcTotalWork(ctx sdk.Context, startHeight uint64, endHeight uint64) *big.Int {
	totalWork := new(big.Int)

	for i := startHeight; i <= endHeight; i++ {
		work := k.GetBlockHeaderByHeight(ctx, i).GetWork()
		totalWork = new(big.Int).Add(totalWork, work)
	}

	return totalWork
}

// ValidateTransaction validates the given transaction
func (k Keeper) ValidateTransaction(ctx sdk.Context, txBytes string, prevTxBytes string, blockHash string, proof []string) (*btcutil.Tx, *btcutil.Tx, error) {
	params := k.GetParams(ctx)

	header := k.GetBlockHeader(ctx, blockHash)
	// Check if block confirmed
	if header == nil || header.Height == 0 {
		return nil, nil, types.ErrBlockNotFound
	}

	best := k.GetBestBlockHeader(ctx)
	// Check if the block is confirmed
	if best.Height-header.Height+1 < uint64(params.Confirmations) {
		return nil, nil, types.ErrNotConfirmed
	}
	// Check if the block is within the acceptable depth
	// if best.Height-header.Height > param.MaxAcceptableBlockDepth {
	//  return types.ErrExceedMaxAcceptanceDepth
	// }

	// Decode the base64 transaction
	rawTx, err := base64.StdEncoding.DecodeString(txBytes)
	if err != nil {
		fmt.Println("Error decoding transaction from base64:", err)
		return nil, nil, err
	}

	// Create a new transaction
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(rawTx))
	if err != nil {
		fmt.Println("Error deserializing transaction:", err)
		return nil, nil, err
	}

	tx := btcutil.NewTx(&msgTx)

	// Validate the transaction
	if err := blockchain.CheckTransactionSanity(tx); err != nil {
		fmt.Println("Transaction is not valid:", err)
		return nil, nil, err
	}

	var prevTx *btcutil.Tx

	// Check the previous tx if given
	if len(prevTxBytes) > 0 {
		// Decode the previous transaction
		rawPrevTx, err := base64.StdEncoding.DecodeString(prevTxBytes)
		if err != nil {
			fmt.Println("Error decoding transaction from base64:", err)
			return nil, nil, err
		}

		// Create a new transaction
		var prevMsgTx wire.MsgTx
		err = prevMsgTx.Deserialize(bytes.NewReader(rawPrevTx))
		if err != nil {
			fmt.Println("Error deserializing transaction:", err)
			return nil, nil, err
		}

		prevTx = btcutil.NewTx(&prevMsgTx)

		// Validate the transaction
		if err := blockchain.CheckTransactionSanity(prevTx); err != nil {
			fmt.Println("Transaction is not valid:", err)
			return nil, nil, err
		}

		if tx.MsgTx().TxIn[0].PreviousOutPoint.Hash.String() != prevTx.Hash().String() {
			return nil, nil, types.ErrInvalidBtcTransaction
		}
	}

	// check if the proof is valid
	root, err := chainhash.NewHashFromStr(header.MerkleRoot)
	if err != nil {
		return nil, nil, err
	}

	if !types.VerifyMerkleProof(proof, tx.Hash(), root) {
		k.Logger(ctx).Error("Invalid merkle proof", "txhash", tx, "root", root, "proof", proof)
		return nil, nil, types.ErrTransactionNotIncluded
	}

	return tx, prevTx, nil
}
