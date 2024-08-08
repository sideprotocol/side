package keeper

import (
	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetRequestSeqence returns the request sequence
func (k Keeper) GetRequestSeqence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SequenceKey)
	if bz == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bz)
}

// IncrementRequestSequence increments the request sequence and returns the new sequence
func (k Keeper) IncrementRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	seq := k.GetRequestSeqence(ctx) + 1
	store.Set(types.SequenceKey, sdk.Uint64ToBigEndian(seq))
	return seq
}

// NewWithdrawRequest creates a new withdrawal request
func (k Keeper) NewWithdrawRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64) (*types.BitcoinWithdrawRequest, error) {
	sequence := k.IncrementRequestSequence(ctx)

	p := k.GetParams(ctx)
	btcVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_BTC)

	switch types.AssetTypeFromDenom(amount.Denom, p) {
	case types.AssetType_ASSET_TYPE_BTC:
		return k.NewBtcWithdrawRequest(ctx, sender, amount, feeRate, btcVault.Address, sequence)

	case types.AssetType_ASSET_TYPE_RUNE:
		runesVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_RUNE)
		return k.NewRunesWithdrawRequest(ctx, sender, amount, feeRate, runesVault.Address, btcVault.Address, sequence)

	default:
		return nil, types.ErrAssetNotSupported
	}
}

// NewBtcWithdrawRequest creates the request for btc withdrawal
func (k Keeper) NewBtcWithdrawRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string, sequence uint64) (*types.BitcoinWithdrawRequest, error) {
	utxoIterator := k.GetUTXOIteratorByAddr(ctx, vault)

	psbt, selectedUTXOs, _, err := types.BuildPsbt(utxoIterator, sender, amount.Amount.Int64(), feeRate, vault, sequence)
	if err != nil {
		return nil, err
	}

	changeUTXO, err := k.handleBtcTxFee(psbt, vault)
	if err != nil {
		return nil, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	// lock the selected utxos
	_ = k.LockUTXOs(ctx, selectedUTXOs)

	// save the change utxo and mark minted
	k.saveUTXO(ctx, changeUTXO)
	k.addToMintHistory(ctx, psbt.UnsignedTx.TxHash().String())

	withdrawRequest := &types.BitcoinWithdrawRequest{
		Address:  sender,
		Sequence: sequence,
		Txid:     psbt.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.WithdrawStatus_WITHDRAW_STATUS_CREATED,
	}

	k.SetWithdrawRequest(ctx, withdrawRequest)

	return withdrawRequest, nil
}

// NewRunesWithdrawRequest creates the request for runes withdrawal
func (k Keeper) NewRunesWithdrawRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string, btcVault string, sequence uint64) (*types.BitcoinWithdrawRequest, error) {
	var runeId types.RuneId
	runeId.FromDenom(amount.Denom)

	runeAmount := uint128.FromBig(amount.Amount.BigInt())

	runesUTXOs, amountDelta := k.GetTargetRunesUTXOs(ctx, vault, runeId.ToString(), runeAmount)
	if len(runesUTXOs) == 0 {
		return nil, types.ErrInsufficientUTXOs
	}

	paymentUTXOIterator := k.GetUTXOIteratorByAddr(ctx, btcVault)

	psbt, selectedUTXOs, changeUTXO, runesChangeUTXO, err := types.BuildRunesPsbt(runesUTXOs, paymentUTXOIterator, sender, runeId.ToString(), runeAmount, feeRate, amountDelta, vault, btcVault, sequence)
	if err != nil {
		return nil, err
	}

	if err := k.handleRunesTxFee(ctx, psbt, sender); err != nil {
		return nil, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	// lock the involved utxos
	_ = k.LockUTXOs(ctx, runesUTXOs)
	_ = k.LockUTXOs(ctx, selectedUTXOs)

	// save the change utxo and mark minted
	if changeUTXO != nil {
		k.saveUTXO(ctx, changeUTXO)
		k.addToMintHistory(ctx, psbt.UnsignedTx.TxHash().String())
	}

	// save the runes change utxo and mark minted
	if runesChangeUTXO != nil {
		k.saveUTXO(ctx, runesChangeUTXO)
		k.addToMintHistory(ctx, psbt.UnsignedTx.TxHash().String())
	}

	withdrawRequest := &types.BitcoinWithdrawRequest{
		Address:  sender,
		Sequence: sequence,
		Txid:     psbt.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.WithdrawStatus_WITHDRAW_STATUS_CREATED,
	}

	k.SetWithdrawRequest(ctx, withdrawRequest)

	return withdrawRequest, nil
}

// HasWithdrawRequest returns true if the given withdrawal request exists, false otherwise
func (k Keeper) HasWithdrawRequest(ctx sdk.Context, sequence uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcWithdrawRequestKey(sequence))
}

// HasWithdrawRequestByTxHash returns true if the given withdrawal request exists, false otherwise
func (k Keeper) HasWithdrawRequestByTxHash(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcWithdrawRequestByTxHashKey(hash))
}

// GetWithdrawRequest returns the withdrawal request by the given sequence
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, sequence uint64) *types.BitcoinWithdrawRequest {
	store := ctx.KVStore(k.storeKey)

	var withdrawRequest types.BitcoinWithdrawRequest
	bz := store.Get(types.BtcWithdrawRequestKey(sequence))
	k.cdc.MustUnmarshal(bz, &withdrawRequest)

	return &withdrawRequest
}

// GetWithdrawRequestByTxHash returns the withdrawal request by the given hash
func (k Keeper) GetWithdrawRequestByTxHash(ctx sdk.Context, hash string) *types.BitcoinWithdrawRequest {
	store := ctx.KVStore(k.storeKey)

	var withdrawRequest types.BitcoinWithdrawRequest
	bz := store.Get(types.BtcWithdrawRequestByTxHashKey(hash))
	k.cdc.MustUnmarshal(bz, &withdrawRequest)

	return &withdrawRequest
}

// SetWithdrawRequest sets the withdrawal request
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, withdrawRequest *types.BitcoinWithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(withdrawRequest)

	store.Set(types.BtcWithdrawRequestKey(withdrawRequest.Sequence), bz)
	store.Set(types.BtcWithdrawRequestByTxHashKey(withdrawRequest.Txid), types.Int64ToBytes(withdrawRequest.Sequence))
}

// IterateWithdrawRequests iterates through all withdrawal requests
func (k Keeper) IterateWithdrawRequests(ctx sdk.Context, cb func(withdrawRequest *types.BitcoinWithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcWithdrawRequestPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var withdrawRequest types.BitcoinWithdrawRequest
		k.cdc.MustUnmarshal(iterator.Value(), &withdrawRequest)

		if cb(&withdrawRequest) {
			break
		}
	}
}

// filter WithdrawRequest by status with pagination
func (k Keeper) FilterWithdrawRequestsByStatus(ctx sdk.Context, req *types.QueryWithdrawRequestsRequest) []*types.BitcoinWithdrawRequest {
	var withdrawRequests []*types.BitcoinWithdrawRequest

	k.IterateWithdrawRequests(ctx, func(withdrawRequest *types.BitcoinWithdrawRequest) (stop bool) {
		if withdrawRequest.Status == req.Status {
			withdrawRequests = append(withdrawRequests, withdrawRequest)
		}

		// pagination TODO: limit the number of withdrawal requests
		if len(withdrawRequests) >= 100 {
			return true
		}

		return false
	})

	return withdrawRequests
}

// filter WithdrawRequest by address with pagination
func (k Keeper) FilterWithdrawRequestsByAddr(ctx sdk.Context, req *types.QueryWithdrawRequestsByAddressRequest) []*types.BitcoinWithdrawRequest {
	var withdrawRequests []*types.BitcoinWithdrawRequest

	k.IterateWithdrawRequests(ctx, func(withdrawRequest *types.BitcoinWithdrawRequest) (stop bool) {
		if withdrawRequest.Address == req.Address {
			withdrawRequests = append(withdrawRequests, withdrawRequest)
		}

		// pagination TODO: limit the number of withdrawal requests
		if len(withdrawRequests) >= 100 {
			return true
		}

		return false
	})

	return withdrawRequests
}

// Process Bitcoin Withdraw Transaction
func (k Keeper) ProcessBitcoinWithdrawTransaction(ctx sdk.Context, msg *types.MsgSubmitWithdrawTransaction) (*chainhash.Hash, error) {
	ctx.Logger().Info("accept bitcoin withdraw tx", "blockhash", msg.Blockhash)

	tx, _, err := k.ValidateTransaction(ctx, msg.TxBytes, msg.PrevTxBytes, msg.Blockhash, msg.Proof)
	if err != nil {
		return nil, err
	}

	txHash := tx.Hash()

	if !k.HasWithdrawRequestByTxHash(ctx, txHash.String()) {
		return nil, types.ErrWithdrawRequestNotExist
	}

	withdrawRequest := k.GetWithdrawRequestByTxHash(ctx, txHash.String())
	if withdrawRequest.Status == types.WithdrawStatus_WITHDRAW_STATUS_CONFIRMED {
		return nil, types.ErrWithdrawRequestConfirmed
	}

	withdrawRequest.Status = types.WithdrawStatus_WITHDRAW_STATUS_CONFIRMED
	k.SetWithdrawRequest(ctx, withdrawRequest)

	// spend the locked utxos
	k.spendUTXOs(ctx, tx)

	// burn the locked assets
	if err := k.burnLockedAssets(ctx, txHash.String()); err != nil {
		return nil, err
	}

	return txHash, nil
}

// spendUTXOs spends locked utxos
func (k Keeper) spendUTXOs(ctx sdk.Context, uTx *btcutil.Tx) {
	for _, in := range uTx.MsgTx().TxIn {
		hash := in.PreviousOutPoint.Hash.String()
		vout := in.PreviousOutPoint.Index

		if k.IsUTXOLocked(ctx, hash, uint64(vout)) {
			_ = k.SpendUTXO(ctx, hash, uint64(vout))
		}
	}
}

// handleTxFee performs the fee handling for the btc withdrawal tx
// Make sure that the given psbt is valid
// There are at most two outputs and the change output is the last one if any
func (k Keeper) handleBtcTxFee(p *psbt.Packet, changeAddr string) (*types.UTXO, error) {
	recipientOut := p.UnsignedTx.TxOut[0]

	changeOut := new(wire.TxOut)
	if len(p.UnsignedTx.TxOut) > 1 {
		changeOut = p.UnsignedTx.TxOut[1]
	} else {
		changeOut = wire.NewTxOut(0, types.MustPkScriptFromAddress(changeAddr))
		p.UnsignedTx.TxOut = append(p.UnsignedTx.TxOut, changeOut)
	}

	txFee, err := p.GetTxFee()
	if err != nil {
		return nil, err
	}

	recipientOut.Value -= int64(txFee)
	changeOut.Value += int64(txFee)

	if types.IsDustOut(recipientOut) || types.IsDustOut(changeOut) {
		return nil, types.ErrDustOutput
	}

	return &types.UTXO{
		Txid:         p.UnsignedTx.TxHash().String(),
		Vout:         1,
		Address:      changeAddr,
		Amount:       uint64(changeOut.Value),
		PubKeyScript: changeOut.PkScript,
	}, nil
}

// handleRunesTxFee performs the fee handling for the runes withdrawal tx
func (k Keeper) handleRunesTxFee(ctx sdk.Context, p *psbt.Packet, recipient string) error {
	txFee, err := p.GetTxFee()
	if err != nil {
		return err
	}

	feeCoin := sdk.NewCoin(k.GetParams(ctx).BtcVoucherDenom, sdk.NewInt(int64(txFee)))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(recipient), types.ModuleName, sdk.NewCoins(feeCoin)); err != nil {
		return err
	}

	k.lockAsset(ctx, p.UnsignedTx.TxHash().String(), feeCoin)

	return nil
}

// lockAsset locks the given asset by the tx hash
// we can guarantee that the keys do not overlap
func (k Keeper) lockAsset(ctx sdk.Context, txHash string, coin sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(&coin)
	store.Set(types.BtcLockedAssetKey(txHash, bz), []byte{})
}

// burnLockedAssets burns the locked assets
func (k Keeper) burnLockedAssets(ctx sdk.Context, txHash string) error {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.BtcLockedAssetKeyPrefix, []byte(txHash)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		var lockedAsset sdk.Coin
		k.cdc.MustUnmarshal(key[1+len(txHash):], &lockedAsset)

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(lockedAsset)); err != nil {
			return err
		}

		store.Delete(key)
	}

	return nil
}
