package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/crypto/schnorr"
	"github.com/sideprotocol/side/x/liquidation/types"
)

// HandleSettlementSignatures handles the settlement tx signatures
func (k Keeper) HandleSettlementSignatures(ctx sdk.Context, sender string, liquidationId uint64, signatures []string) error {
	if !k.HasLiquidation(ctx, liquidationId) {
		return types.ErrLiquidationDoesNotExist
	}

	liquidation := k.GetLiquidation(ctx, liquidationId)
	if liquidation.Status != types.LiquidationStatus_LIQUIDATION_STATUS_SETTLING {
		return errorsmod.Wrap(types.ErrInvalidLiquidationStatus, "non settling status")
	}

	settlementTxPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidation.SettlementTx)), true)
	if err != nil {
		return err
	}

	if len(signatures) != len(settlementTxPsbt.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	dcmPubKey, _ := hex.DecodeString(liquidation.DCM)
	verificationKey := types.GetTaprootOutKey(dcmPubKey)

	for i, input := range settlementTxPsbt.Inputs {
		sigHash, err := types.CalcTaprootSigHash(settlementTxPsbt, i, input.SighashType)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, verificationKey) {
			return types.ErrInvalidSignature
		}

		settlementTxPsbt.Inputs[i].TaprootKeySpendSig = sigBytes
	}

	if err := psbt.MaybeFinalizeAll(settlementTxPsbt); err != nil {
		return err
	}

	settlementTxPsbtB64, err := settlementTxPsbt.B64Encode()
	if err != nil {
		return err
	}

	// update liquidation
	liquidation.SettlementTx = settlementTxPsbtB64
	liquidation.Status = types.LiquidationStatus_LIQUIDATION_STATUS_SETTLED
	k.SetLiquidation(ctx, liquidation)

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateSignedSettlementTransaction,
			sdk.NewAttribute(types.AttributeKeyLiquidationId, fmt.Sprintf("%d", liquidationId)),
			sdk.NewAttribute(types.AttributeKeyTxHash, k.GetLiquidation(ctx, liquidationId).SettlementTxId),
		),
	)

	return nil
}
