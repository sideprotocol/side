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

// BlockHeaders defines a set of block headers which form a chain
type BlockHeaders []*BlockHeader

// Validate validates if each block header is valid and if the block headers form a chain
func (headers BlockHeaders) Validate() error {
	if len(headers) == 0 {
		return errorsmod.Wrap(ErrInvalidBlockHeaders, "block headers can not be empty")
	}

	var lastHeight uint64
	var lastHash string

	for i, h := range headers {
		if err := h.Validate(); err != nil {
			return err
		}

		if i > 0 && h.Height != lastHeight+1 && h.PreviousBlockHash != lastHash {
			return errorsmod.Wrap(ErrInvalidBlockHeaders, "block headers can not form a chain")
		}

		lastHeight = h.Height
		lastHash = h.Hash
	}

	return nil
}

// GetTotalWork gets the total work of the block headers
func (headers BlockHeaders) GetTotalWork() *big.Int {
	totalWork := new(big.Int)

	for _, h := range headers {
		work := h.GetWork()
		totalWork = new(big.Int).Add(totalWork, work)
	}

	return totalWork
}

func BitsToTarget(bits string) *big.Int {
	n := new(big.Int)
	n.SetString(bits, 16)

	return n
}

func BitsToTargetUint32(bits string) uint32 {
	return uint32(BitsToTarget(bits).Uint64())
}
