package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	dlctypes "github.com/sideprotocol/side/x/dlc/types"
	"github.com/sideprotocol/side/x/lending/types"
)

// SetDLCMeta sets the given dlc meta
func (k Keeper) SetDLCMeta(ctx sdk.Context, loanId string, dlcMeta *types.DLCMeta) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(dlcMeta)

	store.Set(types.DLCMetaKey(loanId), bz)
}

// GetDLCMeta gets the specified dlc meta
func (k Keeper) GetDLCMeta(ctx sdk.Context, loanId string) *types.DLCMeta {
	store := ctx.KVStore(k.storeKey)

	var dlcMeta types.DLCMeta
	bz := store.Get(types.DLCMetaKey(loanId))
	k.cdc.MustUnmarshal(bz, &dlcMeta)

	return &dlcMeta
}

// UpdateDLCMeta updates the dlc meta of the given loan with the given params
func (k Keeper) UpdateDLCMeta(ctx sdk.Context, loanId string, depositTxs []*psbt.Packet, liquidationCet string, liquidationAdaptorSignatures []string, defaultLiquidationAdaptorSignatures []string, repaymentCet string, repaymentSignatures []string, feeRate int64) error {
	loan := k.GetLoan(ctx, loanId)
	dlcMeta := k.GetDLCMeta(ctx, loanId)

	vaultPkScript, _ := types.GetPkScriptFromAddress(loanId)

	vaultUtxos, err := types.GetVaultUtxos(depositTxs, vaultPkScript)
	if err != nil {
		return err
	}

	liquidationCetPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(liquidationCet)), true)
	if err != nil {
		return err
	}

	repaymentCetPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(repaymentCet)), true)
	if err != nil {
		return err
	}

	internalKey, _ := hex.DecodeString(dlcMeta.InternalKey)

	liquidationScript, liquidationScriptControlBlock, _ := types.UnwrapLeafScript(dlcMeta.LiquidationScript)
	repaymentScript, repaymentScriptControlBlock, _ := types.UnwrapLeafScript(dlcMeta.RepaymentScript)

	for i := range liquidationCetPsbt.Inputs {
		liquidationCetPsbt.Inputs[i].TaprootInternalKey = internalKey
		liquidationCetPsbt.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: liquidationScriptControlBlock,
				Script:       liquidationScript,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	for i := range repaymentCetPsbt.Inputs {
		repaymentCetPsbt.Inputs[i].TaprootInternalKey = internalKey
		repaymentCetPsbt.Inputs[i].TaprootLeafScript = []*psbt.TaprootTapLeafScript{
			{
				ControlBlock: repaymentScriptControlBlock,
				Script:       repaymentScript,
				LeafVersion:  txscript.BaseLeafVersion,
			},
		}
	}

	liquidationCet, err = liquidationCetPsbt.B64Encode()
	if err != nil {
		return err
	}

	repaymentCet, err = repaymentCetPsbt.B64Encode()
	if err != nil {
		return err
	}

	borrowerPkScript, err := types.GetPkScriptFromPubKey(loan.BorrowerPubKey)
	if err != nil {
		return err
	}

	// timeout refund transaction can be generated offchain as needed
	// err ignored
	timeoutRefundTx, _ := types.CreateTimeoutRefundTransaction(depositTxs, vaultPkScript, borrowerPkScript, internalKey, dlcMeta.TimeoutRefundScript, loan.FinalTimeout, feeRate)

	// update dlc meta
	dlcMeta.LiquidationCet = types.LiquidationCet{
		Tx:                        liquidationCet,
		BorrowerAdaptorSignatures: liquidationAdaptorSignatures,
	}
	dlcMeta.DefaultLiquidationCet = types.LiquidationCet{
		Tx:                        liquidationCet,
		BorrowerAdaptorSignatures: defaultLiquidationAdaptorSignatures,
	}
	dlcMeta.RepaymentCet = types.RepaymentCet{
		Tx:                 repaymentCet,
		BorrowerSignatures: repaymentSignatures,
	}
	dlcMeta.TimeoutRefundTx = timeoutRefundTx
	dlcMeta.VaultUtxos = vaultUtxos

	k.SetDLCMeta(ctx, loanId, dlcMeta)

	return nil
}

// UpdateDLCEvent updates the dlc event of the given loan
func (k Keeper) UpdateDLCEvent(ctx sdk.Context, loan *types.Loan) {
	dlcEvent := k.dlcKeeper.GetEvent(ctx, loan.DlcEventId)

	// update description
	dlcEvent.Description = fmt.Sprintf("DLC event for loan %s", loan.VaultAddress)

	liquidatedOutcome := "" // will be populated once the liquidation price is available
	defaultLiquidatedOutcome := fmt.Sprintf("%d", loan.MaturityTime)
	repaidOutcome := "Repaid"

	// update outcomes
	dlcEvent.Outcomes = []string{
		liquidatedOutcome,
		defaultLiquidatedOutcome,
		repaidOutcome,
	}

	k.dlcKeeper.SetEvent(ctx, dlcEvent)
}

// UpdateDLCEventLiquidatedOutcome populates the 'liquidated' outcome for the dlc event with the specified liquidation price
func (k Keeper) UpdateDLCEventLiquidatedOutcome(ctx sdk.Context, loan *types.Loan, dlcEvent *dlctypes.DLCEvent, liquidationPrice sdkmath.LegacyDec) {
	pool := k.GetPool(ctx, loan.PoolId)

	dlcEvent.Outcomes[types.LiquidatedOutcomeIndex] = types.FormatPriceWithPair(liquidationPrice, types.GetPricePair(pool.Config))
}

// GetCetInfos gets the related cet infos of the given loan
// Assume that the loan exists
func (k Keeper) GetCetInfos(ctx sdk.Context, loanId string, collateralAmount sdk.Coin) ([]*types.CetInfo, error) {
	loan := k.GetLoan(ctx, loanId)
	dlcMeta := k.GetDLCMeta(ctx, loanId)

	liquidationScript, liquidationScriptControlBlock, _ := types.UnwrapLeafScript(dlcMeta.LiquidationScript)
	repaymentScript, repaymentScriptControlBlock, _ := types.UnwrapLeafScript(dlcMeta.RepaymentScript)

	dlcEvent := k.dlcKeeper.GetEvent(ctx, loan.DlcEventId)
	if len(dlcEvent.Outcomes[types.LiquidatedOutcomeIndex]) == 0 {
		// calculate the liquidation price
		liquidationPrice := k.GetLiquidationPrice(ctx, loan, collateralAmount.Amount)

		// update the dlc event outcome
		k.UpdateDLCEventLiquidatedOutcome(ctx, loan, dlcEvent, liquidationPrice)
	}

	liquidationCetInfo, _ := types.GetCetInfo(dlcEvent, types.LiquidatedOutcomeIndex, liquidationScript, liquidationScriptControlBlock, types.DefaultSigHashType)
	defaultLiquidationCetInfo, _ := types.GetCetInfo(dlcEvent, types.DefaultLiquidatedOutcomeIndex, liquidationScript, liquidationScriptControlBlock, types.DefaultSigHashType)
	repaymentCetInfo, _ := types.GetCetInfo(dlcEvent, types.RepaidOutcomeIndex, repaymentScript, repaymentScriptControlBlock, types.DefaultSigHashType)

	return []*types.CetInfo{
		liquidationCetInfo,
		defaultLiquidationCetInfo,
		repaymentCetInfo,
	}, nil
}
