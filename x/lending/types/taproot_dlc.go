package types

import (
	"github.com/btcsuite/btcd/btcutil/psbt"
)

func CreateDLCTransactions(borrower string, agency string, muturityTime int64, finalTimeout int64, gas_fee_per_vb uint64) ([]psbt.Packet, error) {

	return nil, nil
}

func VerifyCET(depositTx *psbt.Packet, cet *psbt.Packet) error {
	if err := depositTx.SanityCheck(); err != nil {
		return ErrInvalidFunding
	}
	if err := cet.SanityCheck(); err != nil {
		return ErrInvalidCET
	}

	fundtxHash := depositTx.UnsignedTx.TxHash()

	if len(depositTx.Outputs) != len(cet.Inputs) {
		return ErrInvalidCET
	}

	for _, input := range cet.UnsignedTx.TxIn {
		if input.PreviousOutPoint.Hash != fundtxHash {
			return ErrInvalidCET
		}
	}

	return nil
}

func CreateRepaymentTransaction(depositTx []string) (*psbt.Packet, error) {

	return nil, nil
}
