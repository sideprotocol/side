package bitcoin

import (
	"strings"

	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/sideprotocol/side/crypto/keys/segwit"
	"github.com/sideprotocol/side/crypto/keys/taproot"
)

const (
	SegWitType  = hd.PubKeyType("segwit")
	TaprootType = hd.PubKeyType("taproot")
)

var SegWit = segWigAlgo{}
var Taproot = taprootAlgo{}

type segWigAlgo struct{}

func (s segWigAlgo) Name() hd.PubKeyType {
	return SegWitType
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s segWigAlgo) Derive() hd.DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		if !strings.HasPrefix(hdPath, "m/84'") {
			sps := strings.Split(hdPath, "/")
			sps[1] = "84'" // replace purpose
			sps[2] = "0'"
			hdPath = strings.Join(sps, "/")
		}
		println("hdPath", hdPath)
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath)

		return derivedKey, err
	}
}

// Generate generates a secp256k1 private key from the given bytes.
func (s segWigAlgo) Generate() hd.GenerateFn {
	return func(bz []byte) types.PrivKey {
		bzArr := make([]byte, segwit.PrivKeySize)
		copy(bzArr, bz)

		return &segwit.PrivKey{Key: bzArr}
	}
}

type taprootAlgo struct{}

func (s taprootAlgo) Name() hd.PubKeyType {
	return TaprootType
}

// Derive derives and returns the secp256k1 private key for the given seed and HD path.
func (s taprootAlgo) Derive() hd.DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		if !strings.HasPrefix(hdPath, "m/86'") {
			sps := strings.Split(hdPath, "/")
			sps[1] = "86'" // replace purpose
			sps[2] = "0'"
			hdPath = strings.Join(sps, "/")
			// panic("Invalid HD path for Taproot")
		}
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := hd.ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := hd.DerivePrivateKeyForPath(masterPriv, ch, hdPath)

		return derivedKey, err
	}
}

// Generate generates a secp256k1 private key from the given bytes.
func (s taprootAlgo) Generate() hd.GenerateFn {
	return func(bz []byte) types.PrivKey {
		bzArr := make([]byte, taproot.PrivKeySize)
		copy(bzArr, bz)

		return &taproot.PrivKey{Key: bzArr}
	}
}
