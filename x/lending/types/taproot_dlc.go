package types

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/psbt"
)

func CreateDLCTransactions(borrower string, agency string, muturityTime int64, finalTimeout int64, gas_fee_per_vb uint64) ([]psbt.Packet, error) {

	return nil, nil
}

// VerifyCETs verifies the given CETs
func VerifyCETs(depositTx *psbt.Packet, cets *Cets) error {
	if err := depositTx.SanityCheck(); err != nil {
		return ErrInvalidFunding
	}

	liquidationCET := cets.Liquidate
	forceRepayCET := cets.ForceRepay
	refundCET := cets.Refund

	liquidationCETPacket, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	forceRepayCETPacket, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(forceRepayCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	refundCETPacket, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(refundCET)), true)
	if err != nil {
		return ErrInvalidCET
	}

	fundtxHash := depositTx.UnsignedTx.TxHash()

	for _, input := range liquidationCETPacket.UnsignedTx.TxIn {
		if input.PreviousOutPoint.Hash != fundtxHash {
			return ErrInvalidCET
		}
	}

	for _, input := range forceRepayCETPacket.UnsignedTx.TxIn {
		if input.PreviousOutPoint.Hash != fundtxHash {
			return ErrInvalidCET
		}
	}

	for _, input := range refundCETPacket.UnsignedTx.TxIn {
		if input.PreviousOutPoint.Hash != fundtxHash {
			return ErrInvalidCET
		}
	}

	return nil
}

func CreateRepaymentTransaction(depositTx []string) (*psbt.Packet, error) {

	return nil, nil
}
