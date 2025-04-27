package keeper

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

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
		oracleKeeper  types.OracleKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	oracleKeeper types.OracleKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		memKey:         memKey,
		bankKeeper:     bankKeeper,
		stakingKeeper:  stakingKeeper,
		oracleKeeper:   oracleKeeper,
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

// ValidateTransaction validates the given transaction
func (k Keeper) ValidateTransaction(ctx sdk.Context, txBytes string, prevTxBytes string, blockHash string, proof []string) (*btcutil.Tx, *btcutil.Tx, error) {
	params := k.GetParams(ctx)

	if !k.oracleKeeper.HasBlockHeader(ctx, blockHash) {
		return nil, nil, types.ErrBlockNotFound
	}

	header := k.oracleKeeper.GetBlockHeader(ctx, blockHash)
	bestHeader := k.oracleKeeper.GetBestBlockHeader(ctx)

	// Check if the block is confirmed
	if bestHeader.Height-header.Height+1 < params.Confirmations {
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
