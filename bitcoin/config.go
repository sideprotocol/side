package bitcoin

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/sideprotocol/side/crypto/keys/segwit"
	"github.com/sideprotocol/side/crypto/keys/taproot"
)

var (
	Network       = &chaincfg.TestNet3Params
	KeyringOption = func(options *keyring.Options) {
		options.SupportedAlgos = keyring.SigningAlgoList{hd.Secp256k1, SegWit, Taproot}
		options.SupportedAlgosLedger = keyring.SigningAlgoList{hd.Secp256k1}
	}
	DefaultSigVerificationGasConsumer = func(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error {
		pubkey := sig.PubKey

		switch pubkey.(type) {
		case *segwit.PubKey:
			meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: Segwit")
			return nil

		case *taproot.PubKey:
			meter.ConsumeGas(params.SigVerifyCostSecp256k1, "ante verify: Taproot")
			return nil
		default:
			return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
		}
	}
)
