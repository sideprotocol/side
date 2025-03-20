package keeper

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
)

// handleLiquidationSignatures handles the liquidation signatures
func (k Keeper) handleLiquidationSignatures(ctx sdk.Context, loan *types.Loan, signatures []string) error {
	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
	if len(dlcMeta.LiquidationCet.AgencySignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return types.ErrInvalidSignature
		}
	}

	dlcMeta.LiquidationCet.AgencySignatures = signatures
	k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return nil
}

// handleDefaultLiquidationSignatures handles the default liquidation signatures
func (k Keeper) handleDefaultLiquidationSignatures(ctx sdk.Context, loan *types.Loan, signatures []string) error {
	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
	if len(dlcMeta.DefaultLiquidationCet.AgencySignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.DefaultLiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	agencyPubKey, _ := hex.DecodeString(loan.Agency)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return types.ErrInvalidSignature
		}
	}

	dlcMeta.DefaultLiquidationCet.AgencySignatures = signatures
	k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return nil
}
