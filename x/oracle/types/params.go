package types

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		KeepBitcoinBlocks: 10,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}
