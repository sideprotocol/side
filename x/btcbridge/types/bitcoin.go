package types

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// default tx version
	TxVersion = 2

	// default minimum relay fee
	MinRelayFee = 1000

	// default sig hash type
	DefaultSigHashType = txscript.SigHashAll
)

// IsDustOut returns true if the given output is dust, false otherwise
func IsDustOut(out *wire.TxOut) bool {
	return !IsOpReturnOutput(out) && mempool.IsDust(out, MinRelayFee)
}

// CheckOutputAmount checks if the given output amount is dust
func CheckOutputAmount(address string, amount int64) error {
	addr, err := btcutil.DecodeAddress(address, sdk.GetConfig().GetBtcChainCfg())
	if err != nil {
		return err
	}

	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}

	if IsDustOut(&wire.TxOut{Value: amount, PkScript: pkScript}) {
		return ErrDustOutput
	}

	return nil
}

// IsOpReturnOutput returns true if the script of the given out starts with OP_RETURN
func IsOpReturnOutput(out *wire.TxOut) bool {
	return len(out.PkScript) > 0 && out.PkScript[0] == txscript.OP_RETURN
}
