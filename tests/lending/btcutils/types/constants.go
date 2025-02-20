package types

const (
	// default tx version
	TxVersion = 2

	// minimum relay fee for transactions in sat/kvB
	MinRelayTxFee = int64(1000)
)

const (
	// witness size for P2TR in bytes
	P2TRWitnessSize = 64

	// witness size for P2WPKH in bytes
	P2WPKHWitnessSize = 72 + 33

	// signature script size for P2SH-P2WPKH in bytes
	NestedSegWitSigScriptSize = 1 + 1 + 1 + 20

	// signature script size for P2PKH in bytes
	P2PKHSigScriptSize = 1 + 72 + 1 + 33
)
