package taproot_test

import (
	"testing"

	"github.com/cosmos/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/sideprotocol/side/bitcoin/keys/taproot"
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

	SIDE_HRP := "bc"
	sdk.GetConfig().SetBech32PrefixForAccount(SIDE_HRP, SIDE_HRP)
	sdk.GetConfig().Seal()
	bech32.SIDE_HRP = SIDE_HRP

	sec, chainCode := hd.ComputeMastersFromSeed(seed)
	derivedPrivKey, err := hd.DerivePrivateKeyForPath(sec, chainCode, "m/86'/0'/0'/0/0")
	require.NoError(t, err, "Private key derivation should not fail")

	priv := taproot.PrivKey{Key: derivedPrivKey}
	pubkey := priv.PubKey()

	msg := []byte("1234")

	sig, err := priv.Sign(msg)
	require.NoError(t, err, "Sign should not fail")
	v := pubkey.VerifySignature(msg, sig)
	require.True(t, v, "Signature should be valid")

	// bech32Address, err := bech32.Encode("bc", pubkey.Address().Bytes())
	// bech32Address, err := segwit.BitCoinAddr(pubKey.Bytes())

	require.Equal(t, 32, len(pubkey.Address()), "Address should be 32 bytes")
	// require.Equal(t, expectedAddress, sdk.AccAddress(pubkey.Address()).String(), "Public key should be 33 bytes")
	bech32Address, err := bech32.Encode("bc", pubkey.Address())
	require.NoError(t, err, "Bech32 encoding should not fail")
	require.Equal(t, expectedAddress, bech32Address, "Bech32 address should match")
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

	// // addr := []byte{123, 95, 226, 43, 84, 70, 247, 198, 46, 162, 123, 139, 215, 28, 239, 148, 224, 63, 61, 242}
	// // _, err = sdkbech32.ConvertAndEncode("bc", addr)

	// //t.Logf("parsed address", dd)
	// require.NoError(t, err)

}
