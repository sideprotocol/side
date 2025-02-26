package codec

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sideprotocol/side/crypto/keys/segwit"
	"github.com/sideprotocol/side/crypto/keys/taproot"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	var pk *cryptotypes.PubKey
	// registry.RegisterInterface("cosmos.crypto.PubKey", pk)
	registry.RegisterImplementations(pk, &segwit.PubKey{})
	registry.RegisterImplementations(pk, &taproot.PubKey{})

	var priv *cryptotypes.PrivKey
	// registry.RegisterInterface("cosmos.crypto.PrivKey", priv)
	registry.RegisterImplementations(priv, &segwit.PrivKey{})
	registry.RegisterImplementations(priv, &taproot.PrivKey{})
}
