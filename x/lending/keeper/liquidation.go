package keeper

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/lending/types"
)

// handleLiquidationSignatures handles the liquidation signatures
func (k Keeper) handleLiquidationSignatures(ctx sdk.Context, loan *types.Loan, signatures []string) error {
	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
	if len(dlcMeta.LiquidationCet.DCMSignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.LiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return types.ErrInvalidSignature
		}
	}

	dlcMeta.LiquidationCet.DCMSignatures = signatures
	k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return nil
}

// handleDefaultLiquidationSignatures handles the default liquidation signatures
func (k Keeper) handleDefaultLiquidationSignatures(ctx sdk.Context, loan *types.Loan, signatures []string) error {
	dlcMeta := k.GetDLCMeta(ctx, loan.VaultAddress)
	if len(dlcMeta.DefaultLiquidationCet.DCMSignatures) > 0 {
		return types.ErrLiquidationSignaturesAlreadyExist
	}

	p, _ := psbt.NewFromRawBytes(bytes.NewReader([]byte(dlcMeta.DefaultLiquidationCet.Tx)), true)
	if len(signatures) != len(p.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	script, _ := hex.DecodeString(dlcMeta.MultisigScript)
	dcmPubKey, _ := hex.DecodeString(loan.DCM)

	for i, input := range p.Inputs {
		sigHash, err := types.CalcTapscriptSigHash(p, i, input.SighashType, script)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, dcmPubKey) {
			return types.ErrInvalidSignature
		}
	}

	dlcMeta.DefaultLiquidationCet.DCMSignatures = signatures
	k.SetDLCMeta(ctx, loan.VaultAddress, dlcMeta)

	return nil
}

// HandleLiquidatedDebt handles the liquidated debt for the liquidated loan
func (k Keeper) HandleLiquidatedDebt(ctx sdk.Context, liquidationId uint64, loanId string, moduleAccount string, debtAmount sdk.Coin) error {
	loan := k.GetLoan(ctx, loanId)
	pool := k.GetPool(ctx, loan.PoolId)

	interest := k.GetCurrentInterest(ctx, loan).Amount
	protocolFee := interest.Mul(sdkmath.NewInt(int64(pool.Config.ReserveFactor))).Quo(sdkmath.NewInt(1000))

	referralFee := sdkmath.ZeroInt()
	actualProtocolFee := protocolFee
	if protocolFee.IsPositive() && types.HasReferralFee(loan, pool) {
		referralFee = protocolFee.Mul(sdkmath.NewInt(int64(pool.Config.ReferralFeeFactor))).Quo(types.Permille)
		actualProtocolFee = protocolFee.Sub(referralFee)
	}

	if debtAmount.Amount.LTE(interest) {
		// TODO
		return nil
	}

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, moduleAccount, types.ModuleName, sdk.NewCoins(debtAmount.SubAmount(protocolFee))); err != nil {
		return err
	}

	if actualProtocolFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(k.ProtocolFeeCollector(ctx)), sdk.NewCoins(sdk.NewCoin(debtAmount.Denom, actualProtocolFee))); err != nil {
			return err
		}
	}

	if referralFee.IsPositive() {
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, moduleAccount, sdk.MustAccAddressFromBech32(loan.Referrer), sdk.NewCoins(sdk.NewCoin(debtAmount.Denom, referralFee))); err != nil {
			return err
		}
	}

	k.AfterPoolRepaid(ctx, loan.PoolId, loan.Maturity, debtAmount.SubAmount(interest), interest, protocolFee, actualProtocolFee, false)

	return nil
}
