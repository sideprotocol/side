package types

import (
	"encoding/hex"
	"fmt"

	h2c "github.com/bytemare/hash2curve/secp256k1"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/adaptor"
	"github.com/sideprotocol/side/crypto/hash"
	"github.com/sideprotocol/side/x/dlc/types"
)

// HashLoanSecret hashes the given secret
// Assume that the secret is a valid hex string
func HashLoanSecret(secret string) string {
	secretBytes, _ := hex.DecodeString(secret)

	return hex.EncodeToString(hash.Sha256(secretBytes))
}

// AdaptorPoint gets the corresponding adaptor point from the given secret
func AdaptorPoint(secret []byte) string {
	return hex.EncodeToString(adaptor.SecretToPubKey(secret))
}

func GetTaprootAddress(script []byte) (*btcutil.AddressTaproot, error) {
	conf := sdk.GetConfig().GetBtcChainCfg()
	return btcutil.NewAddressTaproot(script, conf)
}

// Branch 1: multisig signature script
func CreateMultisigScript(pubKeys []string) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	for i, pubKeyHex := range pubKeys {
		pubKey, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			return nil, err
		}

		builder.AddData(pubKey)

		if i == 0 {
			builder.AddOp(txscript.OP_CHECKSIG)
		} else {
			builder.AddOp(txscript.OP_CHECKSIGADD)
		}
	}

	builder.AddInt64(int64(len(pubKeys)))
	builder.AddOp(txscript.OP_NUMEQUAL)

	return builder.Script()
}

// Branch 2: Hash Time lock script for DCA
func CreateHashTimeLockScript(pubkey string, hashlock string, locktime int64) ([]byte, error) {
	pubKeyBytes, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, err
	}
	// locktime := int64(500000) // Example block height
	hashBytes, err := hex.DecodeString(hashlock)
	if err != nil {
		return nil, err
	}
	builder := txscript.NewScriptBuilder()
	builder.AddInt64(locktime)                     // Add locktime
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY) // Enforce time lock
	builder.AddOp(txscript.OP_DROP)                // Drop locktime from the stack
	builder.AddOp(txscript.OP_SHA256)              // Add hash lock
	builder.AddData(hashBytes)                     // Push hash
	builder.AddOp(txscript.OP_EQUALVERIFY)         // Verify hash preimage
	builder.AddData(pubKeyBytes)                   // Push pubkey
	builder.AddOp(txscript.OP_CHECKSIG)            // Verify signature
	return builder.Script()
}

// Branch 3: PubKey with Time lock script
func CreatePubKeyTimeLockScript(pubKeyHex string, locktime int64) ([]byte, error) {
	// pubKeyHex := "03abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	// locktime := int64(600000) // Example block height
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, err
	}
	builder := txscript.NewScriptBuilder()
	builder.AddInt64(locktime)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddData(pubKey)
	builder.AddOp(txscript.OP_CHECKSIG)
	return builder.Script()
}

// Create Taproot address with complex script branches
func CreateTaprootAddress(branches [][]byte, params *chaincfg.Params) (string, error) {
	// Create Taproot script tree
	leaves := []txscript.TapLeaf{}
	for _, b := range branches {
		leaves = append(leaves, txscript.NewBaseTapLeaf(b))
	}
	tree := txscript.AssembleTaprootScriptTree(leaves...)

	scriptRoot, err := hex.DecodeString(tree.RootNode.TapHash().String())
	if err != nil {
		return "", err
	}

	// Derive Taproot output key
	taprootPubKey := txscript.ComputeTaprootOutputKey(GetInternalKey(), scriptRoot)
	// Generate Taproot address
	address, err := btcutil.NewAddressTaproot(taprootPubKey.SerializeCompressed(), params)
	if err != nil {
		return "", err
	}
	return address.EncodeAddress(), nil
}
func CreateVaultAddress(borrowerPubkey string, dcaPubkey string, loanSecretHash string, muturityTime int64, finalTimeout int64) (string, error) {
	// Define network parameters (e.g., MainNet, TestNet)
	// params := &chaincfg.MainNetParams
	params := sdk.GetConfig().GetBtcChainCfg()
	// Create script branches
	thresholdScript, err := CreateMultisigScript([]string{borrowerPubkey, dcaPubkey})
	if err != nil {
		return "", err
	}
	hashTimeLockScript, err := CreateHashTimeLockScript(dcaPubkey, loanSecretHash, muturityTime)
	if err != nil {
		return "", err
	}
	pubKeyTimeLockScript, err := CreatePubKeyTimeLockScript(borrowerPubkey, finalTimeout)
	if err != nil {
		return "", err
	}
	// Combine branches
	branches := [][]byte{thresholdScript, hashTimeLockScript, pubKeyTimeLockScript}
	// Generate Taproot address
	taprootAddress, err := CreateTaprootAddress(branches, params)
	if err != nil {
		return "", err
	}
	return taprootAddress, nil
}

// GetInternalKey gets the pub key used for taproot address generation
// Generated by hashToCurve("lending") for now
func GetInternalKey() *btcec.PublicKey {
	input := types.ModuleName
	domain := "side.lending.vault"

	p := h2c.HashToCurve([]byte(input), []byte(domain))

	var X, Y btcec.FieldVal
	X.SetByteSlice(p.X.Bytes())
	Y.SetByteSlice(p.Y.Bytes())

	return btcec.NewPublicKey(&X, &Y)
}

// GetTapscriptsMerkleRoot gets the merkle root of the given tapscripts
func GetTapscriptsMerkleRoot(scripts [][]byte) string {
	tapLeaves := []txscript.TapLeaf{}

	for _, s := range scripts {
		tapLeaves = append(tapLeaves, txscript.NewBaseTapLeaf(s))
	}

	return txscript.AssembleTaprootScriptTree(tapLeaves...).RootNode.TapHash().String()
}

// GetVaultPkScript gets the pk script of the given vault
// Assume that the given vault is valid
func GetVaultPkScript(vault string) []byte {
	vaultAddr, err := btcutil.DecodeAddress(vault, sdk.GetConfig().GetBtcChainCfg())
	if err != nil {
		panic(err)
	}

	return vaultAddr.ScriptAddress()
}

// GetAgencyPkScript gets the pk script from the given agency pubkey
// Assume that the given pubkey is valid
func GetAgencyPkScript(agencyPubKey string) []byte {
	pubKey, err := hex.DecodeString(fmt.Sprintf("02%s", agencyPubKey))
	if err != nil {
		panic(err)
	}

	parsedPubKey, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		panic(err)
	}

	taprootOutKey := txscript.ComputeTaprootKeyNoScript(parsedPubKey)

	address, err := btcutil.NewAddressTaproot(taprootOutKey.SerializeCompressed(), sdk.GetConfig().GetBtcChainCfg())
	if err != nil {
		panic(err)
	}

	return address.ScriptAddress()
}
