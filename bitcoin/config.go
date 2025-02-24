package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

var (
	Network       = &chaincfg.TestNet3Params
	KeyringOption = func(options *keyring.Options) {
		options.SupportedAlgos = keyring.SigningAlgoList{hd.Secp256k1, SegWit, Taproot}
		options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.Secp256k1}
	}
)
