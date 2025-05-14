package taproot_test

import (
	"encoding/hex"
	"testing"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaproot(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := bip39.NewSeed(mnemonic, "")
	expectedAddress := "bc1p5cyxnuxmeuwuvkwfem96lqzszd02n6xdcjrs20cac6yqjjwudpxqkedrcr"
	// hrp, data, e := bech32.Decode(expectedAddress, 1024)
	// assert.NoError(t, e)
	// t.Log(hrp, data)
	t.Log("expectedAddress:", expectedAddress)

	sec, chainCode := hd.ComputeMastersFromSeed(seed)
	keyBytes, err := hd.DerivePrivateKeyForPath(sec, chainCode, "m/86'/0'/0'/0/0")

	require.NoError(t, err, "DerivePrivateKeyForPath should not fail")
	t.Logf("pr: %v", hex.EncodeToString(keyBytes))

	_, pubKey := secp256k1.PrivKeyFromBytes(keyBytes)
	t.Logf("pk: %v", hex.EncodeToString(pubKey.SerializeCompressed()))

	tp := txscript.ComputeTaprootKeyNoScript(pubKey)
	assert.NotNil(t, tp, "Taproot key should not be nil")
	t.Logf("pk: %v", hex.EncodeToString(tp.SerializeCompressed()))

	// comp := tp.SerializeCompressed()
	witnessProg := schnorr.SerializePubKey(tp)
	require.Equal(t, 32, len(witnessProg), "Witness program should be 32 bytes")
	tpaddress, err := btcutil.NewAddressTaproot(witnessProg, &chaincfg.MainNetParams)
	assert.NoError(t, err, "NewAddressTaproot should not fail")
	tpaddressStr := tpaddress.EncodeAddress()
	t.Log("tpaddressStr:", tpaddressStr)
	require.Equal(t, expectedAddress, tpaddressStr, "Address should match")

	// verify := pubKey.VerifySignature([]byte("1234"), sig)
	// assert.True(t, verify, "Verify should be true")

	// bech32Address, err := bech32.Encode("bc", pubKey.Address().Bytes())
	// // bech32Address, err := segwit.BitCoinAddr(pubKey.Bytes())
	// assert.NoError(t, err)
	// t.Logf("Generated SegWit Address: %s", bech32Address)
	// // Check if the Bech32 encoded address has the correct prefix and structure.
	// assert.True(t, strings.HasPrefix(bech32Address, "bc1q"), "Address should start with 'bc1q'")
	// t.Logf("Generated SegWit Address: %s", bech32Address)

	// // data, err := sdk.GetFromBech32(bech32Address, "bc")

	// hrp, version, data, err2 := bech32.DecodeUnsafe(bech32Address)
	// assert.NoError(t, err2)

	// println(hrp, version, data)
	// t.Log(hrp)

	// hrp, bz, err := bech32.Decode(bech32Address, 1000)
	// //hrp, bz, err := bech32.Decode("bc1qc2zm9xeje96yh6st7wmy60mmsteemsm3tfr2tn", 1000)
	// assert.NoError(t, err)
	// println(hrp, bz)
	// sdk.GetConfig().SetBech32PrefixForAccount("bc", "bc")
	// sdk.GetConfig().Seal()
	// acc, err := sdk.AccAddressFromBech32(bech32Address)
	// require.NoError(t, err)
	// t.Logf("Generated SegWit Address: %s", acc)
	// // addr := []byte{123, 95, 226, 43, 84, 70, 247, 198, 46, 162, 123, 139, 215, 28, 239, 148, 224, 63, 61, 242}
	// // _, err = sdkbech32.ConvertAndEncode("bc", addr)

	// //t.Logf("parsed address", dd)
	// require.NoError(t, err)

}
