package keeper

import (
	"bytes"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	"github.com/btcsuite/btcd/btcutil/psbt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/crypto/schnorr"
	"github.com/sideprotocol/side/x/auction/types"
)

// HandlePaymentTransactionSignatures handles the payment tx signatures
func (k Keeper) HandlePaymentTransactionSignatures(ctx sdk.Context, sender string, auctionId uint64, signatures []string) error {
	if !k.HasAuction(ctx, auctionId) {
		return types.ErrAuctionDoesNotExist
	}

	auction := k.GetAuction(ctx, auctionId)
	if auction.Status != types.AuctionStatus_AUCTION_STATUS_SETTLED {
		return types.ErrInvalidAuctionStatus
	}

	paymentTxPsbt, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(auction.PaymentTx)), true)
	if err != nil {
		return err
	}

	if len(signatures) != len(paymentTxPsbt.Inputs) {
		return errorsmod.Wrap(types.ErrInvalidSignatures, "mismatched signature number")
	}

	agencyPubKey, _ := hex.DecodeString(auction.Agency)

	for i, input := range paymentTxPsbt.Inputs {
		sigHash, err := types.CalcTaprootSigHash(paymentTxPsbt, i, input.SighashType)
		if err != nil {
			return err
		}

		sigBytes, _ := hex.DecodeString(signatures[i])

		if !schnorr.Verify(sigBytes, sigHash, agencyPubKey) {
			return types.ErrInvalidSignature
		}

		paymentTxPsbt.Inputs[i].TaprootKeySpendSig = sigBytes
	}

	if err := psbt.MaybeFinalizeAll(paymentTxPsbt); err != nil {
		return err
	}

	paymentTxPsbtB64, err := paymentTxPsbt.B64Encode()
	if err != nil {
		return err
	}

	// update auction
	auction.PaymentTx = paymentTxPsbtB64
	k.SetAuction(ctx, auction)

	return nil
}
