package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"

	"github.com/sideprotocol/side/bitcoin/keys/segwit"
	"github.com/sideprotocol/side/bitcoin/keys/taproot"
)

func init() {
	RegisterCrypto(legacy.Cdc)
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&segwit.PubKey{},
		segwit.PubKeyName, nil)
	cdc.RegisterConcrete(&taproot.PubKey{},
		taproot.PubKeyName, nil)
	cdc.RegisterConcrete(&segwit.PrivKey{},
		segwit.PrivKeyName, nil)
	cdc.RegisterConcrete(&taproot.PrivKey{},
		taproot.PrivKeyName, nil)
}
