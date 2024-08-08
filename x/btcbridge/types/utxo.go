package types

// UTXOIterator defines the iterator over the utxos
type UTXOIterator interface {
	Valid() bool
	Next()
	Close() error

	GetUTXO() *UTXO
}
