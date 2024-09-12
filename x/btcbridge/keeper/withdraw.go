package keeper

import (
	"bytes"

	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// GetRequestSequence returns the request sequence
func (k Keeper) GetRequestSequence(ctx sdk.Context) uint64 {
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
	seq := k.GetRequestSequence(ctx) + 1
	store.Set(types.SequenceKey, sdk.Uint64ToBigEndian(seq))
	return seq
}

// NewSigningRequest creates a new withdrawal request
func (k Keeper) NewSigningRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64) (*types.SigningRequest, error) {
	p := k.GetParams(ctx)
	btcVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_BTC)

	switch types.AssetTypeFromDenom(amount.Denom, p) {
	case types.AssetType_ASSET_TYPE_BTC:
		return k.NewBtcSigningRequest(ctx, sender, amount, feeRate, btcVault.Address)

	case types.AssetType_ASSET_TYPE_RUNES:
		runesVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_RUNES)
		return k.NewRunesSigningRequest(ctx, sender, amount, feeRate, runesVault.Address, btcVault.Address)

	default:
		return nil, types.ErrAssetNotSupported
	}
}

// NewBtcSigningRequest creates the signing request for btc withdrawal
func (k Keeper) NewBtcSigningRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string) (*types.SigningRequest, error) {
	utxoIterator := k.GetUTXOIteratorByAddr(ctx, vault)

	psbt, selectedUTXOs, changeUTXO, err := types.BuildPsbt(utxoIterator, sender, amount.Amount.Int64(), feeRate, vault)
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
	if changeUTXO != nil {
		k.saveUTXO(ctx, changeUTXO)
		k.addToMintHistory(ctx, psbt.UnsignedTx.TxHash().String())
	}

	signingRequest := &types.SigningRequest{
		Address:  sender,
		Sequence: k.IncrementRequestSequence(ctx),
		Txid:     psbt.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	k.SetSigningRequest(ctx, signingRequest)

	return signingRequest, nil
}

// NewRunesSigningRequest creates the signing request for runes withdrawal
func (k Keeper) NewRunesSigningRequest(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string, btcVault string) (*types.SigningRequest, error) {
	var runeId types.RuneId
	runeId.FromDenom(amount.Denom)

	runeAmount := uint128.FromBig(amount.Amount.BigInt())

	runesUTXOs, runeBalancesDelta := k.GetTargetRunesUTXOs(ctx, vault, runeId.ToString(), runeAmount)
	if len(runesUTXOs) == 0 {
		return nil, types.ErrInsufficientUTXOs
	}

	paymentUTXOIterator := k.GetUTXOIteratorByAddr(ctx, btcVault)

	psbt, selectedUTXOs, changeUTXO, runesChangeUTXO, err := types.BuildRunesPsbt(runesUTXOs, paymentUTXOIterator, sender, runeId.ToString(), runeAmount, feeRate, runeBalancesDelta, vault, btcVault)
	if err != nil {
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

	signingRequest := &types.SigningRequest{
		Address:  sender,
		Sequence: k.IncrementRequestSequence(ctx),
		Txid:     psbt.UnsignedTx.TxHash().String(),
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	k.SetSigningRequest(ctx, signingRequest)

	return signingRequest, nil
}

// HasSigningRequest returns true if the given signing request exists, false otherwise
func (k Keeper) HasSigningRequest(ctx sdk.Context, sequence uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcSigningRequestKey(sequence))
}

// HasSigningRequestByTxHash returns true if the given withdrawal request exists, false otherwise
func (k Keeper) HasSigningRequestByTxHash(ctx sdk.Context, hash string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.BtcSigningRequestByTxHashKey(hash))
}

// GetSigningRequest returns the withdrawal request by the given sequence
func (k Keeper) GetSigningRequest(ctx sdk.Context, sequence uint64) *types.SigningRequest {
	store := ctx.KVStore(k.storeKey)

	var signingRequest types.SigningRequest
	bz := store.Get(types.BtcSigningRequestKey(sequence))
	k.cdc.MustUnmarshal(bz, &signingRequest)

	return &signingRequest
}

// GetSigningRequestByTxHash returns the withdrawal request by the given hash
func (k Keeper) GetSigningRequestByTxHash(ctx sdk.Context, hash string) *types.SigningRequest {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcSigningRequestByTxHashKey(hash))
	if bz != nil {
		return k.GetSigningRequest(ctx, sdk.BigEndianToUint64(bz))
	}

	return nil
}

// SetSigningRequest sets the withdrawal request
func (k Keeper) SetSigningRequest(ctx sdk.Context, signingRequest *types.SigningRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(signingRequest)

	store.Set(types.BtcSigningRequestKey(signingRequest.Sequence), bz)
	store.Set(types.BtcSigningRequestByTxHashKey(signingRequest.Txid), sdk.Uint64ToBigEndian(signingRequest.Sequence))
}

// IterateSigningRequests iterates through all withdrawal requests
func (k Keeper) IterateSigningRequests(ctx sdk.Context, cb func(signingRequest *types.SigningRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcSigningRequestPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var signingRequest types.SigningRequest
		k.cdc.MustUnmarshal(iterator.Value(), &signingRequest)

		if cb(&signingRequest) {
			break
		}
	}
}

// FilterSigningRequestsByStatus filters withdrawal requests by status with pagination
func (k Keeper) FilterSigningRequestsByStatus(ctx sdk.Context, req *types.QuerySigningRequestsRequest) []*types.SigningRequest {
	var signingRequests []*types.SigningRequest

	k.IterateSigningRequests(ctx, func(signingRequest *types.SigningRequest) (stop bool) {
		if signingRequest.Status == req.Status {
			signingRequests = append(signingRequests, signingRequest)
		}

		// pagination TODO: limit the number of withdrawal requests
		if len(signingRequests) >= 100 {
			return true
		}

		return false
	})

	return signingRequests
}

// FilterSigningRequestsByAddr filters signing requests by address with pagination
func (k Keeper) FilterSigningRequestsByAddr(ctx sdk.Context, req *types.QuerySigningRequestsByAddressRequest) []*types.SigningRequest {
	var signingRequests []*types.SigningRequest

	k.IterateSigningRequests(ctx, func(signingRequest *types.SigningRequest) (stop bool) {
		if signingRequest.Address == req.Address {
			signingRequests = append(signingRequests, signingRequest)
		}

		// pagination TODO: limit the number of signing requests
		if len(signingRequests) >= 100 {
			return true
		}

		return false
	})

	return signingRequests
}

// ProcessBitcoinWithdrawTransaction handles the withdrawal transaction
func (k Keeper) ProcessBitcoinWithdrawTransaction(ctx sdk.Context, msg *types.MsgSubmitWithdrawTransaction) (*chainhash.Hash, error) {
	ctx.Logger().Info("accept bitcoin withdraw tx", "blockhash", msg.Blockhash)

	tx, _, err := k.ValidateTransaction(ctx, msg.TxBytes, "", msg.Blockhash, msg.Proof)
	if err != nil {
		return nil, err
	}

	txHash := tx.Hash()

	if !k.HasSigningRequestByTxHash(ctx, txHash.String()) {
		return nil, types.ErrSigningRequestNotExist
	}

	signingRequest := k.GetSigningRequestByTxHash(ctx, txHash.String())
	if signingRequest.Status == types.SigningStatus_SIGNING_STATUS_CONFIRMED {
		return nil, types.ErrSigningRequestConfirmed
	}

	signingRequest.Status = types.SigningStatus_SIGNING_STATUS_CONFIRMED
	k.SetSigningRequest(ctx, signingRequest)

	// spend the locked utxos
	k.spendUTXOs(ctx, tx)

	// burn the locked assets
	if err := k.burnLockedAssets(ctx, txHash.String()); err != nil {
		return nil, err
	}

	return txHash, nil
}

// LockAssets locks the related assets for the given signing request
func (k Keeper) LockAssets(ctx sdk.Context, req *types.SigningRequest, amount sdk.Coin) error {
	btcNetworkFee, err := k.getBtcNetworkFee(ctx, req.Psbt)
	if err != nil {
		return err
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(req.Address), types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	if err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(req.Address), types.ModuleName, sdk.NewCoins(btcNetworkFee)); err != nil {
		return err
	}

	// mark locked assets which will be burned when the withdrawal tx is relayed back
	k.lockAssets(ctx, req.Txid, amount, btcNetworkFee)

	return nil
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

// handleWithdrawProtocolFee performs the protocol fee handling and returns the actual withdrawal amount
func (k Keeper) handleWithdrawProtocolFee(ctx sdk.Context, sender sdk.AccAddress, amount sdk.Coin) (sdk.Coin, error) {
	params := k.GetParams(ctx)

	protocolFee := sdk.NewInt64Coin(params.BtcVoucherDenom, params.ProtocolFees.WithdrawFee)
	protocoFeeCollector := sdk.MustAccAddressFromBech32(params.ProtocolFees.Collector)

	var err error
	withdrawAmount := amount

	if amount.Denom == params.BtcVoucherDenom {
		withdrawAmount, err = amount.SafeSub(protocolFee)
		if err != nil || withdrawAmount.Amount.Int64() < params.ProtocolLimits.BtcMinWithdraw || withdrawAmount.Amount.Int64() > params.ProtocolLimits.BtcMaxWithdraw {
			return sdk.Coin{}, types.ErrInvalidWithdrawAmount
		}
	}

	if err := k.bankKeeper.SendCoins(ctx, sender, protocoFeeCollector, sdk.NewCoins(protocolFee)); err != nil {
		return sdk.Coin{}, err
	}

	return withdrawAmount, nil
}

// getBtcNetworkFee gets the bitcoin network fee for the given withdrawal psbt
func (k Keeper) getBtcNetworkFee(ctx sdk.Context, packet string) (sdk.Coin, error) {
	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(packet)), true)
	if err != nil {
		return sdk.Coin{}, err
	}

	txFee, err := p.GetTxFee()
	if err != nil {
		return sdk.Coin{}, err
	}

	return sdk.NewCoin(k.GetParams(ctx).BtcVoucherDenom, sdk.NewInt(int64(txFee))), nil
}

// lockAssets locks the given assets by the tx hash
func (k Keeper) lockAssets(ctx sdk.Context, txHash string, coins ...sdk.Coin) {
	store := ctx.KVStore(k.storeKey)

	for i, coin := range coins {
		bz := k.cdc.MustMarshal(&coin)
		store.Set(types.BtcLockedAssetKey(txHash, uint8(i)), bz)
	}
}

// burnLockedAssets burns the locked assets
func (k Keeper) burnLockedAssets(ctx sdk.Context, txHash string) error {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.BtcLockedAssetKeyPrefix, []byte(txHash)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var lockedAsset sdk.Coin
		k.cdc.MustUnmarshal(iterator.Value(), &lockedAsset)

		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(lockedAsset)); err != nil {
			return err
		}

		store.Delete(iterator.Key())
	}

	return nil
}
