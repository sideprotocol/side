package bitcoin_test

import (
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/bitcoin"
	"github.com/sideprotocol/side/bitcoin/keys/segwit"
	"github.com/sideprotocol/side/bitcoin/keys/taproot"
	"github.com/stretchr/testify/assert"
)

func TestAddressEncodeDecode(t *testing.T) {

	conf := sdk.GetConfig()
	conf.SetBech32PrefixForAccount("side", "side")
	conf.Seal()

	adds := []string{
		"side10d07y265gmmuvt4z0w9aw880jnsr700jwrwlg5",
		"bc1qqs4cyfvr6fwlca38hvyrgwl08k7cxme6jw3rr6",
		"bc1q73ssvy27zd8kjhrzjalzjkfdya0kd9na8pz00n",
		"bc1q3v4fcnzdtduepkxhuq4cwehsw3pgtn4gakpc9t",
		"bc1pln2mzgrk689xfuacgmwpym95karxf8283qh9k7ze5ucc7crl6qrq4w30es",
		"bc1p93svdel208e2wva9gmnqsm3hd5p0k768a9pyg0ptd7r4lzl0sxvqeaw5gv",
	}

	for _, a := range adds {

		addr, err := sdk.AccAddressFromBech32(a)
		assert.NoError(t, err, "invalid address "+a)
		if strings.HasPrefix(a, "side") {
			assert.Equal(t, 20, len(addr.Bytes()), a)
		} else if strings.HasPrefix(a, "bc1q") {
			assert.Equal(t, 33, len(addr.Bytes()), a)
		} else {
			assert.Equal(t, 32, len(addr.Bytes()), a)
		}

		text_addr := addr.String()
		assert.EqualValues(t, a, text_addr, "address should equals")

	}

}

func TestGenKeys(t *testing.T) {

	conf := sdk.GetConfig()
	conf.SetBech32PrefixForAccount("side", "side")
	conf.Seal()
	bitcoin.Network = &chaincfg.MainNetParams

	// hash := btcutil.Hash160([]byte{0, 3, 3, 3, 3, 3})
	hash := make([]byte, 32, 32)
	assert.Equal(t, 32, len(hash))

	// sh, err := btcutil.NewAddressScriptHashFromHash(hash, &chaincfg.MainNetParams)
	// assert.NoError(t, err)
	std, err := btcutil.NewAddressTaproot(hash, bitcoin.Network)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(std.AddressSegWit.ScriptAddress()))
	// println(std.ScriptAddress())
	// text := std.AddressSegWit.EncodeAddress()
	// _, bte, err := bech32.Decode(text)
	// assert.NoError(t, err)
	// assert.Equal(t, 53, len(bte), text)
	// a_str := sdk.MustAccAddressFromBech32(text)
	// assert.Equal(t, 53, len(a_str.Bytes()), text)
	// assert.Equal(t, bte, a_str.Bytes())

	// btcutil.DecodeAddress()
	// println("taproot:", len(taproot.GenPrivKey().PubKey().Address()))
	a := segwit.GenPrivKey().PubKey().Address()
	t.Log("segwit:", len(a), sdk.AccAddress(a).String())

	// aa, _ := btcutil.NewAddressWitnessPubKeyHash(a.Bytes(), keys.Network)
	// t.Log("aa", aa.EncodeAddress())

	b := taproot.GenPrivKey().PubKey().Address()
	t.Log("taproot:", len(b), sdk.AccAddress(b).String())
	bb, _ := btcutil.NewAddressTaproot(b.Bytes(), bitcoin.Network)
	t.Log("bb", bb.EncodeAddress())

	// println("bech32:", text)
	assert.Equal(t, sdk.AccAddress(b).String(), bb.EncodeAddress(), "bech32 address should be equal")

	// addrs := []sdk.Address{sdk.AccAddress(taproot.GenPrivKey().PubKey().Address()), sdk.AccAddress(segwit.GenPrivKey().PubKey().Address())}

	// for _, a := range addrs {
	// 	println(len(a.Bytes()), a.String())
	// 	assert.Equal(t, true, strings.HasPrefix(a.String(), "bc"), a.String())
	// 	if strings.HasPrefix(a.String(), "bc1p") {
	// 		assert.Equal(t, 32, len(a.Bytes()), a.String())
	// 		a2, err := sdk.AccAddressFromBech32(a.String())
	// 		assert.NoError(t, err)
	// 		assert.Equal(t, 32, len(a2.Bytes()))
	// 	} else {
	// 		assert.Equal(t, 33, len(a.Bytes()), a.String())
	// 	}
	// 	// a2, err := sdk.AccAddressFromBech32(a.String())
	// 	// assert.Equal(t, 53, len(a2.Bytes()))
	// 	// assert.NoError(t, err, a.String())
	// 	// assert.Equal(t, a.Bytes(), a2.Bytes(), a.String())
	// }
}

func TestValAddressEncodeDecode(t *testing.T) {

	conf := sdk.GetConfig()
	conf.SetBech32PrefixForAccount("side", "side")
	conf.SetBech32PrefixForValidator("sidevaloper", "sidevaloper")
	conf.Seal()

	vals := []string{
		// "sidevaloper1qqwpwrc0qs0pvrc6rvrsxrc2p583vqstpgdqxxsmzgp3y9gfpvqp7srxm9c", // error case
		// "sidevaloper1qqgsc9gfrqfsyrgtp5wpjqgkqct3cqq4rq8pj9cspcgszzqtzv2ssmdxyv7",
		"sidevaloper1qq9qjpswp5t3j9c8pq93xrc7zsysqpq6r5dqvpczpqz3yzqfrqwsc0aalqh",
		"sidevaloper1l90rhzz8tka095tpvhjgzamvsvxyqhu34e9335l9npy2zx83lmcq9y82l7",
		"sidevaloper1qqy3sqqmpv83xqcfry8qvyqazvqp6qqgru23y9q2q52swxggqg8sya58uzu",
		"sidevaloper1qq0pkzghzcvsz8qwzcqq6xs6rv8qwxctzyzqq8shzy9qu8qhrcgsq8gftvt",
		"sidevaloper1qqr3wzgzqvpqxycjzyr3kzcequxskxqxzydsjqgeqygq2xqxqgtpuaaajmf",
		"sidevaloper1yezttrzdh00zmtzzfau9vuy360hxkjnl3gfxw2jzfqz67d8t8myq79pv5g",
	}

	for _, a := range vals {

		addr, err := sdk.ValAddressFromBech32(a)
		assert.NoError(t, err, "invalid address "+a)
		t.Log("addr", addr.String())
		if strings.HasPrefix(a, "sidevaloper1p") {
			assert.Equal(t, 33, len(addr))
		} else if strings.HasPrefix(a, "sidevaloper1q") {
			assert.Equal(t, 33, len(addr))
		} else {
			assert.Equal(t, 32, len(addr))
		}

		text_addr := addr.String()
		assert.EqualValues(t, a, text_addr, "address should equals")

	}

}
