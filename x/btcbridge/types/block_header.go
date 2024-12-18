package types

import (
	"math/big"
	time "time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates the block header
func (header *BlockHeader) Validate() error {
	wireHeader := header.ToWireHeader()

	if err := blockchain.CheckBlockHeaderSanity(
		wireHeader,
		sdk.GetConfig().GetBtcChainCfg().PowLimit,
		blockchain.NewMedianTime(),
		blockchain.BFNone,
	); err != nil {
		return errorsmod.Wrapf(ErrInvalidBlockHeader, "check failed: %v", err)
	}

	if header.Hash != wireHeader.BlockHash().String() {
		return errorsmod.Wrap(ErrInvalidBlockHeader, "incorrect block hash")
	}

	return nil
}

// ToWireHeader converts the block header to wire.BlockHeader
func (header *BlockHeader) ToWireHeader() *wire.BlockHeader {
	prevBlockHash, _ := chainhash.NewHashFromStr(header.PreviousBlockHash)
	merkleRoot, _ := chainhash.NewHashFromStr(header.MerkleRoot)

	bits := new(big.Int)
	bits.SetString(header.Bits, 16)

	return &wire.BlockHeader{
		Version:    int32(header.Version),
		PrevBlock:  *prevBlockHash,
		MerkleRoot: *merkleRoot,
		Timestamp:  time.Unix(int64(header.Time), 0),
		Bits:       uint32(bits.Uint64()),
		Nonce:      uint32(header.Nonce),
	}
}

// GetWork gets the work of the block header
func (header *BlockHeader) GetWork() *big.Int {
	return blockchain.CalcWork(BitsToTargetUint32(header.Bits))
}

func BitsToTarget(bits string) *big.Int {
	n := new(big.Int)
	n.SetString(bits, 16)

	return n
}

func BitsToTargetUint32(bits string) uint32 {
	return uint32(BitsToTarget(bits).Uint64())
}
