package segwit

import (
	"encoding/binary"

	"github.com/cometbft/cometbft/crypto"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

var MagicBytes = []byte("Bitcoin Signed Message:\n")

func VarintBufNum(n uint64) []byte {
	var buf []byte
	if n < 253 {
		buf = make([]byte, 1)
		buf[0] = byte(n)
	} else if n < 0x10000 {
		buf = make([]byte, 1+2)
		buf[0] = 253
		binary.LittleEndian.PutUint16(buf[1:], uint16(n))
	} else if n < 0x100000000 {
		buf = make([]byte, 1+4)
		buf[0] = 254
		binary.LittleEndian.PutUint32(buf[1:], uint32(n))
	} else {
		// This is original code from JS wallet, But it's not clear how to implement n & -1 in Go
		// buf = Buffer.alloc(1 + 8);
		// buf.writeUInt8(255, 0);
		// buf.writeInt32LE(n & -1, 1);
		// buf.writeUInt32LE(Math.floor(n / 0x100000000), 5);
		buf = make([]byte, 1+8)
		buf[0] = 255
		binary.PutVarint(buf[1:], int64(n)&-1) // n & -1, need to check
		binary.LittleEndian.PutUint32(buf[5:], uint32(n/0x100000000))
	}
	return buf
}

func MagicHash(msg []byte) []byte {
	return magicHash(msg)
}
func magicHash(msg []byte) []byte {
	prefix1 := VarintBufNum(uint64(len(MagicBytes)))
	prefix2 := VarintBufNum(uint64(len(msg)))
	buf := append(prefix1, MagicBytes...)
	buf = append(buf, prefix2...)
	buf = append(buf, msg...)

	return crypto.Sha256(crypto.Sha256(buf))
}

// Sign creates an ECDSA signature on curve Secp256k1, using SHA256 on the msg.
func (privKey *PrivKey) Sign(msg []byte) ([]byte, error) {

	derivedKey, _ := btcec.PrivKeyFromBytes(privKey.Key)
	hash := MagicHash(msg)

	return ecdsa.SignCompact(derivedKey, hash, true), nil
}

// VerifySignature validates the signature.
// The msg will be hashed prior to signature verification.
func (pubKey *PubKey) VerifySignature(msg []byte, sigBytes []byte) bool {
	pk, err := btcec.ParsePubKey(pubKey.Key)
	if err != nil {
		return false
	}

	hash := magicHash(msg)
	recoveredPK, compressed, err := ecdsa.RecoverCompact(sigBytes, hash)
	if err != nil {
		return false
	}
	if !compressed {
		return false
	}
	return recoveredPK.IsEqual(pk)

}
