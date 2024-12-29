package types

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// Branch 1: multisig signature script
func createMultisigScript(pubKeys []string) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddInt64(int64(len(pubKeys))) // Threshold: 2 signatures required
	for _, pubKeyHex := range pubKeys {
		pubKey, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			return nil, err
		}
		builder.AddData(pubKey)
	}
	builder.AddInt64(int64(len(pubKeys))) // Total keys
	builder.AddOp(txscript.OP_CHECKMULTISIG)
	return builder.Script()
}

// Branch 2: Hash Time lock script for DCA
func createHashTimeLockScript(pubkey string, hashlock string, locktime int64) ([]byte, error) {
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
	builder.AddOp(txscript.OP_DUP)                 // Duplicate public key for signature verification
	builder.AddOp(txscript.OP_HASH160)             // Hash public key
	builder.AddData(pubKeyBytes)                   // Replace with actual hash160 of receiver's pubkey
	builder.AddOp(txscript.OP_EQUALVERIFY)         // Verify public key hash builder.AddOp(txscript.OP_CHECKSIG) // Verify signature
	return builder.Script()
}

// Branch 3: PubKey with Time lock script
func createPubKeyTimeLockScript(pubKeyHex string, locktime int64) ([]byte, error) {
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
func createTaprootAddress(branches [][]byte, params *chaincfg.Params) (string, error) {
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
	taprootPubKey := txscript.ComputeTaprootOutputKey(nil, scriptRoot)
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
	thresholdScript, err := createMultisigScript([]string{borrowerPubkey, dcaPubkey})
	if err != nil {
		return "", err
	}
	hashTimeLockScript, err := createHashTimeLockScript(dcaPubkey, loanSecretHash, muturityTime)
	if err != nil {
		return "", err
	}
	pubKeyTimeLockScript, err := createPubKeyTimeLockScript(borrowerPubkey, finalTimeout)
	if err != nil {
		return "", err
	}
	// Combine branches
	branches := [][]byte{thresholdScript, hashTimeLockScript, pubKeyTimeLockScript}
	// Generate Taproot address
	taprootAddress, err := createTaprootAddress(branches, params)
	if err != nil {
		return "", err
	}
	return taprootAddress, nil
}
