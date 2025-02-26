package codec

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/sideprotocol/side/crypto/keys/segwit"
	"github.com/sideprotocol/side/crypto/keys/taproot"
)

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
