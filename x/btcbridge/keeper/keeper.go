package keeper

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

type (
	Keeper struct {
		BaseUTXOKeeper

		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		authority string

		bankKeeper    types.BankKeeper
		stakingKeeper types.StakingKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	authority string,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		authority:      authority,
		bankKeeper:     bankKeeper,
		stakingKeeper:  stakingKeeper,
		BaseUTXOKeeper: *NewBaseUTXOKeeper(cdc, storeKey),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
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

func (k Keeper) SetBlockHeaders(ctx sdk.Context, blockHeader []*types.BlockHeader) error {
	store := ctx.KVStore(k.storeKey)
	// check if the previous block header exists
	best := k.GetBestBlockHeader(ctx)
	for _, header := range blockHeader {

		// check the block header sanity
		err := blockchain.CheckBlockHeaderSanity(
			HeaderConvert(header),
			sdk.GetConfig().GetBtcChainCfg().PowLimit,
			blockchain.NewMedianTime(),
			blockchain.BFNone,
		)
		if err != nil {
			return err
		}

		// check whether it's next block header or not
		if best.Hash != header.PreviousBlockHash {
			// check if the block header already exists
			// if exists, then it is a forked block header
			if !store.Has(types.BtcBlockHeaderHeightKey(header.Height)) {
				return types.ErrInvalidHeader
			}

			// a forked block header is detected
			// check if the new block header has more work than the old one
			oldNode := k.GetBlockHeaderByHeight(ctx, header.Height)
			worksOld := blockchain.CalcWork(BitsToTargetUint32(oldNode.Bits))
			worksNew := blockchain.CalcWork(BitsToTargetUint32(header.Bits))
			if worksNew.Cmp(worksOld) <= 0 {
				return types.ErrForkedBlockHeader
			}

			// remove the block headers after the forked block header
			// and consider the forked block header as the best block header
			for i := header.Height; i <= best.Height; i++ {
				ctx.Logger().Info("Removing block header: ", i)
				thash := k.GetBlockHashByHeight(ctx, i)
				store.Delete(types.BtcBlockHeaderHashKey(thash))
				store.Delete(types.BtcBlockHeaderHeightKey(i))
			}
		}

		// store the block header
		bz := k.cdc.MustMarshal(header)
		store.Set(types.BtcBlockHeaderHashKey(header.Hash), bz)
		// store the height to hash mapping
		store.Set(types.BtcBlockHeaderHeightKey(header.Height), []byte(header.Hash))
		// update the best block header
		best = header
	}

	if len(blockHeader) > 0 {
		// set the best block header
		k.SetBestBlockHeader(ctx, best)
	}

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
	iterator := sdk.KVStorePrefixIterator(store, types.BtcBlockHeaderHashPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var header types.BlockHeader
		k.cdc.MustUnmarshal(iterator.Value(), &header)
		if process(header) {
			break
		}
	}
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
	if best.Height-header.Height < uint64(params.Confirmations) {
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
	var tx wire.MsgTx
	err = tx.Deserialize(bytes.NewReader(rawTx))
	if err != nil {
		fmt.Println("Error deserializing transaction:", err)
		return nil, nil, err
	}

	uTx := btcutil.NewTx(&tx)

	// Validate the transaction
	if err := blockchain.CheckTransactionSanity(uTx); err != nil {
		fmt.Println("Transaction is not valid:", err)
		return nil, nil, err
	}

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

	prevTx := btcutil.NewTx(&prevMsgTx)

	// Validate the transaction
	if err := blockchain.CheckTransactionSanity(prevTx); err != nil {
		fmt.Println("Transaction is not valid:", err)
		return nil, nil, err
	}

	if uTx.MsgTx().TxIn[0].PreviousOutPoint.Hash.String() != prevTx.Hash().String() {
		return nil, nil, types.ErrInvalidBtcTransaction
	}

	// check if the proof is valid
	root, err := chainhash.NewHashFromStr(header.MerkleRoot)
	if err != nil {
		return nil, nil, err
	}

	if !types.VerifyMerkleProof(proof, uTx.Hash(), root) {
		k.Logger(ctx).Error("Invalid merkle proof", "txhash", tx, "root", root, "proof", proof)
		return nil, nil, types.ErrTransactionNotIncluded
	}

	return uTx, prevTx, nil
}
