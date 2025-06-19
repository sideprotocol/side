package keeper

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
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

// GetCetInfos gets the related cet infos of the given loan
// Assume that the loan exists
func (k Keeper) GetCetInfos(ctx sdk.Context, loanId string, collateralAmount sdk.Coin) ([]*types.CetInfo, error) {
	loan := k.GetLoan(ctx, loanId)
	poolConfig := k.GetPool(ctx, loan.PoolId).Config

	borrowerPubKey, _ := hex.DecodeString(loan.BorrowerPubKey)
	borrowerAuthPubKey, _ := hex.DecodeString(loan.BorrowerAuthPubKey)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	liquidationScript, _ := types.CreateMultisigScript([][]byte{borrowerAuthPubKey, dcmPubKey})
	repaymentScript, _ := types.CreateMultisigScript([][]byte{borrowerPubKey, dcmPubKey})
	timeRefundScript, _ := types.CreatePubKeyTimeLockScript(borrowerPubKey, loan.FinalTimeout)

	merkleTree := types.GetTapscriptTree([][]byte{liquidationScript, repaymentScript, timeRefundScript})

	liquidationScriptProof := merkleTree.LeafMerkleProofs[0]
	repaymentScriptProof := merkleTree.LeafMerkleProofs[1]

	internalKey := types.GetInternalKey(borrowerPubKey, dcmPubKey)

	liquidationScriptControlBlock, err := types.GetControlBlock(internalKey, liquidationScriptProof)
	if err != nil {
		return nil, err
	}

	repaymentScriptControlBlock, err := types.GetControlBlock(internalKey, repaymentScriptProof)
	if err != nil {
		return nil, err
	}

	var liquidationEvent *dlctypes.DLCEvent
	if loan.LiquidationEventId != 0 {
		liquidationEvent = k.dlcKeeper.GetEvent(ctx, loan.LiquidationEventId)
	} else if collateralAmount.Amount.IsPositive() {
		pricePair, found := k.dlcKeeper.PricePair(ctx, types.GetPricePair(poolConfig))
		if !found {
			return nil, errorsmod.Wrap(types.ErrInvalidPricePair, "price pair does not exist in dlc")
		}

		liquidationPrice := types.GetLiquidationPrice(collateralAmount.Amount, int(poolConfig.CollateralAsset.Decimals), loan.BorrowAmount.Amount, int(poolConfig.LendingAsset.Decimals), loan.Maturity, loan.BorrowAPR, k.GetBlocksPerYear(ctx), poolConfig.LiquidationThreshold, int(pricePair.Decimals), pricePair.Interval, poolConfig.CollateralAsset.IsBasePriceAsset)
		liquidationEvent = k.dlcKeeper.GetEventByPrice(ctx, pricePair.Pair, dlctypes.NormalizePrice(liquidationPrice, int(pricePair.Decimals)))
	}

	defaultLiquidationEvent := k.dlcKeeper.GetEvent(ctx, loan.DefaultLiquidationEventId)
	repaymentEvent := k.dlcKeeper.GetEvent(ctx, loan.RepaymentEventId)

	liquidationCetInfo, _ := types.GetCetInfo(liquidationEvent, 0, liquidationScript, liquidationScriptControlBlock)
	defaultLiquidationCetInfo, _ := types.GetCetInfo(defaultLiquidationEvent, 0, liquidationScript, liquidationScriptControlBlock)
	repaymentCetInfo, _ := types.GetCetInfo(repaymentEvent, 0, repaymentScript, repaymentScriptControlBlock)

	return []*types.CetInfo{
		liquidationCetInfo,
		defaultLiquidationCetInfo,
		repaymentCetInfo,
	}, nil
}
