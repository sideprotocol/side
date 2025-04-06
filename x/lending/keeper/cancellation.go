package keeper

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
)

// HandleCancellationSignatures handles the cancellation signatures
func (k Keeper) HandleCancellationSignatures(ctx sdk.Context, loanId string, signatures []string) error {
	if !k.HasLoan(ctx, loanId) {
		return types.ErrLoanDoesNotExist
	}

	if !k.HasCancellation(ctx, loanId) {
		return types.ErrCancellationDoesNotExist
	}

	cancellation := k.GetCancellation(ctx, loanId)
	if len(cancellation.DCMSignatures) != 0 {
		return types.ErrDCMSignaturesAlreadyExist
	}

	loan := k.GetLoan(ctx, loanId)

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(cancellation.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	script, _ := hex.DecodeString(k.GetDLCMeta(ctx, loanId).MultisigScript)
	leafHash := txscript.NewBaseTapLeaf(script).TapHash()

	for i := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, types.DefaultSigHashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return types.ErrInvalidSignature
		}

		borrowerSig, _ := hex.DecodeString(cancellation.Signatures[i])

		p.Inputs[i].TaprootScriptSpendSig = []*psbt.TaprootScriptSpendSig{
			{
				XOnlyPubKey: dcmPubKey,
				LeafHash:    leafHash[:],
				Signature:   sigBytes,
				SigHash:     txscript.SigHashDefault,
			},
			{
				XOnlyPubKey: borrowerPubKey,
				LeafHash:    leafHash[:],
				Signature:   borrowerSig,
				SigHash:     txscript.SigHashDefault,
			},
		}
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return err
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return err
	}

	cancellation.Tx = serializedTx
	cancellation.DCMSignatures = signatures
	k.SetCancellation(ctx, cancellation)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateSignedCancellationTransaction,
			sdk.NewAttribute(types.AttributeKeyLoanId, loanId),
			sdk.NewAttribute(types.AttributeKeyTxHash, cancellation.Txid),
		),
	)

	return nil
}
