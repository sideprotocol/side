package keeper

import (
	"bytes"

	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// IncreaseWithdrawRequestSequence increases the withdrawal request sequence by 1
func (k Keeper) IncreaseWithdrawRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	sequence := k.GetWithdrawRequestSequence(ctx)
	store.Set(types.BtcWithdrawRequestSequenceKey, sdk.Uint64ToBigEndian(sequence+1))

	return sequence + 1
}

// GetWithdrawRequestSequence gets the withdrawal request sequence
func (k Keeper) GetWithdrawRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcWithdrawRequestSequenceKey)
	if bz != nil {
		return sdk.BigEndianToUint64(bz)
	}

	return 0
}

// GetSigningRequestSequence returns the signing request sequence
func (k Keeper) GetSigningRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcSigningRequestSequenceKey)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// IncrementSigningRequestSequence increments the signing request sequence and returns the new sequence
func (k Keeper) IncrementSigningRequestSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	seq := k.GetSigningRequestSequence(ctx) + 1
	store.Set(types.BtcSigningRequestSequenceKey, sdk.Uint64ToBigEndian(seq))

	return seq
}

// HandleWithdrawal handles the given withdrawal request
func (k Keeper) HandleWithdrawal(ctx sdk.Context, sender string, amount sdk.Coin) (*types.WithdrawRequest, error) {
	switch types.AssetTypeFromDenom(amount.Denom, k.GetParams(ctx)) {
	case types.AssetType_ASSET_TYPE_BTC:
		return k.HandleBtcWithdrawal(ctx, sender, amount)

	case types.AssetType_ASSET_TYPE_RUNES:
		return k.HandleRunesWithdrawal(ctx, sender, amount)

	default:
		return nil, types.ErrAssetNotSupported
	}
}

// HandleBtcWithdrawal handles the given btc withdrawal request
// Btc withdrawal request will be dispatched to the batch withdrawal queue which is handled periodically
func (k Keeper) HandleBtcWithdrawal(ctx sdk.Context, sender string, amount sdk.Coin) (*types.WithdrawRequest, error) {
	// set the withdrawal request
	withdrawRequest := k.NewWithdrawRequest(ctx, sender, amount.String())
	k.SetWithdrawRequest(ctx, withdrawRequest)

	// add to the pending queue
	k.AddToBtcWithdrawRequestQueue(ctx, withdrawRequest)

	feeRate := k.GetFeeRate(ctx)
	if feeRate == 0 {
		return nil, types.ErrInvalidFeeRate
	}

	// estimate the btc network fee
	networkFee, err := k.EstimateWithdrawalNetworkFee(ctx, sender, amount, feeRate)
	if err != nil {
		return nil, err
	}

	// burn asset
	if err := k.BurnAsset(ctx, sender, amount); err != nil {
		return nil, err
	}

	// burn btc network fee
	if err := k.BurnAsset(ctx, sender, networkFee); err != nil {
		return nil, err
	}

	return withdrawRequest, nil
}

// HandleRunesWithdrawal handles the given runes withdrawal request
// Runes withdrawal will generate a signing request immediately
func (k Keeper) HandleRunesWithdrawal(ctx sdk.Context, sender string, amount sdk.Coin) (*types.WithdrawRequest, error) {
	// build the withdrawal request
	withdrawRequest := k.NewWithdrawRequest(ctx, sender, amount.String())

	// build the signing request

	vaults := k.GetParams(ctx).Vaults
	btcVault := types.SelectVaultByAssetType(vaults, types.AssetType_ASSET_TYPE_BTC)
	runesVault := types.SelectVaultByAssetType(vaults, types.AssetType_ASSET_TYPE_RUNES)

	feeRate := k.GetFeeRate(ctx)
	if feeRate == 0 {
		return nil, types.ErrInvalidFeeRate
	}

	signingRequest, err := k.NewRunesSigningRequest(ctx, sender, amount, feeRate, runesVault.Address, btcVault.Address)
	if err != nil {
		return nil, err
	}

	// set the withdrawal request
	withdrawRequest.Txid = signingRequest.Txid
	k.SetWithdrawRequest(ctx, withdrawRequest)

	// burn asset
	if err := k.BurnAsset(ctx, sender, amount); err != nil {
		return nil, err
	}

	// burn btc network fee
	if err := k.BurnBtcNetworkFee(ctx, sender, signingRequest.Psbt); err != nil {
		return nil, err
	}

	return withdrawRequest, nil
}

// NewWithdrawRequest builds a new withdrawal request
func (k Keeper) NewWithdrawRequest(ctx sdk.Context, sender string, amount string) *types.WithdrawRequest {
	return &types.WithdrawRequest{
		Address:  sender,
		Amount:   amount,
		Sequence: k.IncreaseWithdrawRequestSequence(ctx),
	}
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

	if err := k.checkUtxos(ctx, selectedUTXOs); err != nil {
		return nil, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	txHash := psbt.UnsignedTx.TxHash().String()

	// lock the selected utxos
	_ = k.LockUTXOs(ctx, selectedUTXOs)

	// save the change utxo
	k.saveChangeUTXOs(ctx, txHash, changeUTXO)

	signingRequest := &types.SigningRequest{
		Address:  sender,
		Sequence: k.IncrementSigningRequestSequence(ctx),
		Txid:     txHash,
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

	if err := k.checkUtxos(ctx, runesUTXOs, selectedUTXOs); err != nil {
		return nil, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	txHash := psbt.UnsignedTx.TxHash().String()

	// lock the involved utxos
	_ = k.LockUTXOs(ctx, runesUTXOs)
	_ = k.LockUTXOs(ctx, selectedUTXOs)

	// save the change utxos
	k.saveChangeUTXOs(ctx, txHash, changeUTXO, runesChangeUTXO)

	signingRequest := &types.SigningRequest{
		Address:  sender,
		Sequence: k.IncrementSigningRequestSequence(ctx),
		Txid:     txHash,
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	k.SetSigningRequest(ctx, signingRequest)

	return signingRequest, nil
}

// BuildBtcBatchWithdrawSigningRequest builds the signing request for btc batch withdrawal
func (k Keeper) BuildBtcBatchWithdrawSigningRequest(ctx sdk.Context, withdrawRequests []*types.WithdrawRequest, feeRate int64, vault string) (*types.SigningRequest, error) {
	utxoIterator := k.GetUTXOIteratorByAddr(ctx, vault)

	psbt, selectedUTXOs, changeUTXO, err := types.BuildBtcBatchWithdrawPsbt(utxoIterator, withdrawRequests, feeRate, vault)
	if err != nil {
		return nil, err
	}

	if err := k.checkUtxos(ctx, selectedUTXOs); err != nil {
		return nil, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return nil, types.ErrFailToSerializePsbt
	}

	txHash := psbt.UnsignedTx.TxHash().String()

	// lock the selected utxos
	_ = k.LockUTXOs(ctx, selectedUTXOs)

	// save the change utxo
	k.saveChangeUTXOs(ctx, txHash, changeUTXO)

	signingRequest := &types.SigningRequest{
		Address:  authtypes.NewModuleAddress(types.ModuleName).String(),
		Sequence: k.IncrementSigningRequestSequence(ctx),
		Txid:     txHash,
		Psbt:     psbtB64,
		Status:   types.SigningStatus_SIGNING_STATUS_PENDING,
	}

	k.SetSigningRequest(ctx, signingRequest)

	return signingRequest, nil
}

// BuildWithdrawTx builds the bitcoin tx for the given withdrawal
func (k Keeper) BuildWithdrawTx(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64) (*psbt.Packet, error) {
	p := k.GetParams(ctx)
	btcVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_BTC)

	switch types.AssetTypeFromDenom(amount.Denom, p) {
	case types.AssetType_ASSET_TYPE_BTC:
		return k.BuildWithdrawBtcTx(ctx, sender, amount, feeRate, btcVault.Address)

	case types.AssetType_ASSET_TYPE_RUNES:
		runesVault := types.SelectVaultByAssetType(p.Vaults, types.AssetType_ASSET_TYPE_RUNES)
		return k.BuildWithdrawRunesTx(ctx, sender, amount, feeRate, runesVault.Address, btcVault.Address)

	default:
		return nil, types.ErrAssetNotSupported
	}
}

// BuildWithdrawBtcTx builds the bitcoin tx for the btc withdrawal
func (k Keeper) BuildWithdrawBtcTx(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string) (*psbt.Packet, error) {
	utxoIterator := k.GetUTXOIteratorByAddr(ctx, vault)

	psbt, selectedUTXOs, _, err := types.BuildPsbt(utxoIterator, sender, amount.Amount.Int64(), feeRate, vault)
	if err != nil {
		return nil, err
	}

	if err := k.checkUtxos(ctx, selectedUTXOs); err != nil {
		return nil, err
	}

	return psbt, nil
}

// BuildWithdrawRunesTx builds the bitcoin tx for the runes withdrawal
func (k Keeper) BuildWithdrawRunesTx(ctx sdk.Context, sender string, amount sdk.Coin, feeRate int64, vault string, btcVault string) (*psbt.Packet, error) {
	var runeId types.RuneId
	runeId.FromDenom(amount.Denom)

	runeAmount := uint128.FromBig(amount.Amount.BigInt())

	runesUTXOs, runeBalancesDelta := k.GetTargetRunesUTXOs(ctx, vault, runeId.ToString(), runeAmount)
	if len(runesUTXOs) == 0 {
		return nil, types.ErrInsufficientUTXOs
	}

	paymentUTXOIterator := k.GetUTXOIteratorByAddr(ctx, btcVault)

	psbt, selectedUTXOs, _, _, err := types.BuildRunesPsbt(runesUTXOs, paymentUTXOIterator, sender, runeId.ToString(), runeAmount, feeRate, runeBalancesDelta, vault, btcVault)
	if err != nil {
		return nil, err
	}

	if err := k.checkUtxos(ctx, runesUTXOs, selectedUTXOs); err != nil {
		return nil, err
	}

	return psbt, nil
}

// EstimateBtcNetworkFee estimates the btc network fee for the given withdrawal
func (k Keeper) EstimateWithdrawalNetworkFee(ctx sdk.Context, address string, amount sdk.Coin, feeRate int64) (sdk.Coin, error) {
	psbt, err := k.BuildWithdrawTx(ctx, address, amount, feeRate)
	if err != nil {
		return sdk.Coin{}, err
	}

	psbtB64, err := psbt.B64Encode()
	if err != nil {
		return sdk.Coin{}, types.ErrFailToSerializePsbt
	}

	networkFee, err := k.getBtcNetworkFee(ctx, psbtB64)
	if err != nil {
		return sdk.Coin{}, err
	}

	return networkFee, nil
}

// GetWithdrawRequest gets the withdrawal request by the given sequence
func (k Keeper) GetWithdrawRequest(ctx sdk.Context, sequence uint64) *types.WithdrawRequest {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcWithdrawRequestKey(sequence))
	var req types.WithdrawRequest
	k.cdc.MustUnmarshal(bz, &req)

	return &req
}

// GetWithdrawRequestsByAddress gets the withdrawal requests by the given address
func (k Keeper) GetWithdrawRequestsByAddress(ctx sdk.Context, address string) []*types.WithdrawRequest {
	requests := make([]*types.WithdrawRequest, 0)

	k.IterateWithdrawRequests(ctx, func(req *types.WithdrawRequest) (stop bool) {
		if req.Address != address {
			return false
		}

		requests = append(requests, req)

		// TODO: pagination
		return len(requests) >= 100
	})

	return requests
}

// GetPendingBtcWithdrawRequests gets the pending btc withdrawal requests up to the given maximum number
func (k Keeper) GetPendingBtcWithdrawRequests(ctx sdk.Context, maxNum uint32) []*types.WithdrawRequest {
	requests := make([]*types.WithdrawRequest, 0)

	k.IterateBtcWithdrawRequestQueue(ctx, func(req *types.WithdrawRequest) (stop bool) {
		requests = append(requests, req)

		return maxNum != 0 && len(requests) >= int(maxNum)
	})

	return requests
}

// GetWithdrawRequestsByTxHash gets the withdrawal requests by the given tx hash
func (k Keeper) GetWithdrawRequestsByTxHash(ctx sdk.Context, txHash string) []*types.WithdrawRequest {
	requests := make([]*types.WithdrawRequest, 0)

	k.IterateWithdrawRequestsByTxHash(ctx, txHash, func(req *types.WithdrawRequest) (stop bool) {
		requests = append(requests, req)

		return false
	})

	return requests
}

// SetWithdrawRequest sets the given withdrawal request
func (k Keeper) SetWithdrawRequest(ctx sdk.Context, req *types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(req)
	store.Set(types.BtcWithdrawRequestKey(req.Sequence), bz)

	if len(req.Txid) > 0 {
		k.SetWithdrawRequestByTxHash(ctx, req)
	}
}

// SetWithdrawRequestByTxHash sets the given withdrawal request by tx hash
func (k Keeper) SetWithdrawRequestByTxHash(ctx sdk.Context, req *types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.BtcWithdrawRequestByTxHashKey(req.Txid, req.Sequence), []byte{})
}

// AddToBtcWithdrawRequestQueue adds the given btc withdrawal request to the pending queue
func (k Keeper) AddToBtcWithdrawRequestQueue(ctx sdk.Context, req *types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.BtcWithdrawRequestQueueKey(req.Sequence), []byte{})
}

// RemoveFromBtcWithdrawRequestQueue removes the given btc withdrawal request from the pending queue
func (k Keeper) RemoveFromBtcWithdrawRequestQueue(ctx sdk.Context, req *types.WithdrawRequest) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.BtcWithdrawRequestQueueKey(req.Sequence))
}

// IterateWithdrawRequests iterates through all withdrawal requests
func (k Keeper) IterateWithdrawRequests(ctx sdk.Context, cb func(req *types.WithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcWithdrawRequestKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var request types.WithdrawRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)

		if cb(&request) {
			break
		}
	}
}

// IterateWithdrawRequestsByTxHash iterates through all withdrawal requests by the given tx hash
func (k Keeper) IterateWithdrawRequestsByTxHash(ctx sdk.Context, txHash string, cb func(req *types.WithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.BtcWithdrawRequestByTxHashKeyPrefix, []byte(txHash)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		sequence := sdk.BigEndianToUint64(iterator.Key()[1+64:])
		request := k.GetWithdrawRequest(ctx, sequence)

		if cb(request) {
			break
		}
	}
}

// IterateBtcWithdrawRequestQueue iterates through the btc withdrawal request queue
func (k Keeper) IterateBtcWithdrawRequestQueue(ctx sdk.Context, cb func(req *types.WithdrawRequest) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcWithdrawRequestQueueKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		sequence := sdk.BigEndianToUint64(iterator.Key()[1:])
		request := k.GetWithdrawRequest(ctx, sequence)

		if cb(request) {
			break
		}
	}
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

// GetSigningRequestByTxHash returns the signing request by the given hash
func (k Keeper) GetSigningRequestByTxHash(ctx sdk.Context, hash string) *types.SigningRequest {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.BtcSigningRequestByTxHashKey(hash))
	if bz != nil {
		return k.GetSigningRequest(ctx, sdk.BigEndianToUint64(bz))
	}

	return nil
}

// SetSigningRequest sets the signing request
func (k Keeper) SetSigningRequest(ctx sdk.Context, signingRequest *types.SigningRequest) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(signingRequest)

	store.Set(types.BtcSigningRequestKey(signingRequest.Sequence), bz)
	store.Set(types.BtcSigningRequestByTxHashKey(signingRequest.Txid), sdk.Uint64ToBigEndian(signingRequest.Sequence))
}

// IterateSigningRequests iterates through all signing requests
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

// FilterSigningRequestsByStatus filters signing requests by status with pagination
func (k Keeper) FilterSigningRequestsByStatus(ctx sdk.Context, req *types.QuerySigningRequestsRequest) ([]*types.SigningRequest, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	signingRequestStore := prefix.NewStore(store, types.BtcSigningRequestPrefix)

	var signingRequests []*types.SigningRequest

	pageRes, err := query.Paginate(signingRequestStore, req.Pagination, func(key []byte, value []byte) error {
		var signingRequest types.SigningRequest
		k.cdc.MustUnmarshal(value, &signingRequest)

		if signingRequest.Status == req.Status {
			signingRequests = append(signingRequests, &signingRequest)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return signingRequests, pageRes, nil
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

	// unlock the change utxos
	k.unlockChangeUTXOs(ctx, txHash.String())

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

// saveChangeUTXOs saves the change utxos of the given tx and marks minted
func (k Keeper) saveChangeUTXOs(ctx sdk.Context, txHash string, utxos ...*types.UTXO) {
	for _, utxo := range utxos {
		if utxo == nil {
			continue
		}

		utxo.IsLocked = true
		k.saveUTXO(ctx, utxo)

		k.addToMintHistory(ctx, txHash)
	}
}

// unlockChangeUTXOs unlocks the change utxos of the given tx
func (k Keeper) unlockChangeUTXOs(ctx sdk.Context, txHash string) {
	k.IterateUTXOsByTxHash(ctx, txHash, func(utxo *types.UTXO) (stop bool) {
		utxo.IsLocked = false
		k.SetUTXO(ctx, utxo)

		return false
	})
}

// BurnAsset burns the asset related to the withdrawal
func (k Keeper) BurnAsset(ctx sdk.Context, address string, amount sdk.Coin) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32(address), types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	return nil
}

// BurnBtcNetworkFee burns the bitcoin network fee of the withdrawal psbt
func (k Keeper) BurnBtcNetworkFee(ctx sdk.Context, sender string, packet string) error {
	networkFee, err := k.getBtcNetworkFee(ctx, packet)
	if err != nil {
		return err
	}

	return k.BurnAsset(ctx, sender, networkFee)
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

// checkUtxos checks if the total count of the given utxos exceeds the allowed maximum number
func (k Keeper) checkUtxos(ctx sdk.Context, utxoSets ...[]*types.UTXO) error {
	count := 0

	for _, utxoSet := range utxoSets {
		count += len(utxoSet)
	}

	return k.checkUtxoCount(ctx, count)
}

// checkUtxoCount checks if the given utxo count exceeds the allowed maximum number
func (k Keeper) checkUtxoCount(ctx sdk.Context, utxoCount int) error {
	maxUtxoNum := k.GetParams(ctx).WithdrawParams.MaxUtxoNum
	if maxUtxoNum != 0 && utxoCount > int(maxUtxoNum) {
		return types.ErrMaxUTXONumExceeded
	}

	return nil
}
