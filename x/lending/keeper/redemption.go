package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// HandleRedemptionSignatures handles the redemption signatures
// Assume that signatures have already been verified
func (k Keeper) HandleRedemptionSignatures(ctx sdk.Context, id uint64, signatures []string) error {
	if !k.HasRedemption(ctx, id) {
		return types.ErrRedemptionDoesNotExist
	}

	redemption := k.GetRedemption(ctx, id)
	if len(redemption.DCMSignatures) != 0 {
		return types.ErrDCMSignaturesAlreadyExist
	}

	loan := k.GetLoan(ctx, redemption.LoanId)

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(redemption.Tx)), true)

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, ti := range p.UnsignedTx.TxIn {
		prevTxHash := ti.PreviousOutPoint.Hash.String()

		sigBytes, _ := hex.DecodeString(signatures[i])
		borrowerSig, _ := hex.DecodeString(redemption.Signatures[i])

		leafHash := txscript.NewBaseTapLeaf(p.Inputs[i].TaprootLeafScript[0].Script).TapHash()

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

		// update deposit status
		depositLog := k.GetDepositLog(ctx, prevTxHash)
		depositLog.Status = types.DepositStatus_DEPOSIT_STATUS_REDEEMED
		k.SetDepositLog(ctx, depositLog)
	}

	if err := psbt.MaybeFinalizeAll(p); err != nil {
		return err
	}

	serializedTx, err := p.B64Encode()
	if err != nil {
		return err
	}

	redemption.Tx = serializedTx
	redemption.DCMSignatures = signatures
	k.SetRedemption(ctx, redemption)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGenerateSignedRedemptionTransaction,
			sdk.NewAttribute(types.AttributeKeyId, fmt.Sprintf("%d", id)),
			sdk.NewAttribute(types.AttributeKeyLoanId, redemption.LoanId),
			sdk.NewAttribute(types.AttributeKeyTxHash, redemption.Txid),
		),
	)

	return nil
}

// GetRedemptionScript gets the script along with the corresponding control block for redemption
func (k Keeper) GetRedemptionScript(ctx sdk.Context, loanId string) ([]byte, []byte, error) {
	dlcMeta := k.GetDLCMeta(ctx, loanId)
	if len(dlcMeta.RepaymentCet.Tx) > 0 {
		repaymentCet, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.RepaymentCet.Tx)), true)

		script := repaymentCet.Inputs[0].TaprootLeafScript[0].Script
		controlBlock := repaymentCet.Inputs[0].TaprootLeafScript[0].ControlBlock

		return script, controlBlock, nil
	}

	liquidationScript, _ := hex.DecodeString(dlcMeta.LiquidationScript)
	repaymentScript, _ := hex.DecodeString(dlcMeta.RepaymentScript)
	timeoutRefundScript, _ := hex.DecodeString(dlcMeta.TimeoutRefundScript)

	merkleTree := types.GetTapscriptTree([][]byte{liquidationScript, repaymentScript, timeoutRefundScript})
	repaymentScriptProof := merkleTree.LeafMerkleProofs[1]

	internalKeyBytes, _ := hex.DecodeString(dlcMeta.InternalKey)
	internalKey, _ := schnorr.ParsePubKey(internalKeyBytes)

	repaymentScriptControlBlock, err := types.GetControlBlock(internalKey, repaymentScriptProof)
	if err != nil {
		return nil, nil, err
	}

	return repaymentScript, repaymentScriptControlBlock, nil
}

// GetRedemptionId gets the current redemption id
func (k Keeper) GetRedemptionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.RedemptionIdKey)

	return sdk.BigEndianToUint64(bz)
}

// IncrementRedemptionId increments the redemption id and returns the new id
func (k Keeper) IncrementRedemptionId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetRedemptionId(ctx) + 1
	store.Set(types.RedemptionIdKey, sdk.Uint64ToBigEndian(id))

	return id
}

// HasRedemption returns true if the given redemption exists, false otherwise
func (k Keeper) HasRedemption(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RedemptionKey(id))
}

// SetRedemption sets the given redemption
func (k Keeper) SetRedemption(ctx sdk.Context, redemption *types.Redemption) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(redemption)

	store.Set(types.RedemptionKey(redemption.Id), bz)
}

// GetRedemption gets the specified redemption
func (k Keeper) GetRedemption(ctx sdk.Context, id uint64) *types.Redemption {
	store := ctx.KVStore(k.storeKey)

	var redemption types.Redemption
	bz := store.Get(types.RedemptionKey(id))
	k.cdc.MustUnmarshal(bz, &redemption)

	return &redemption
}
