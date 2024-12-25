package types

// this line is used by starport scaffolding # genesis/types/import

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
		// BestBlockHeader: DefaultBestBlockHeader(),
		// BlockHeaders:    []*BlockHeader{},
		// Utxos:           []*UTXO{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// this line is used by starport scaffolding # genesis/types/validate

	// validate the best block header
	// if err := gs.BestBlockHeader.Validate(); err != nil {
	// 	return err
	// }

	// // validate params
	// return gs.Params.Validate()
	return nil
}
