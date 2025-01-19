package keeper

import (
	"bytes"

	"lukechampine.com/uint128"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

type UTXOViewKeeper interface {
	HasUTXO(ctx sdk.Context, hash string, vout uint64) bool
	IsUTXOLocked(ctx sdk.Context, hash string, vout uint64) bool

	GetUTXO(ctx sdk.Context, hash string, vout uint64) *types.UTXO
	GetAllUTXOs(ctx sdk.Context) []*types.UTXO

	GetUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO
	GetUTXOIteratorByAddr(ctx sdk.Context, addr string) types.UTXOIterator
	GetUnlockedUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO

	GetTargetRunesUTXOs(ctx sdk.Context, addr string, runeId string, targetAmount uint128.Uint128, maxNum int) ([]*types.UTXO, []*types.RuneBalance)

	IterateAllUTXOs(ctx sdk.Context, cb func(utxo *types.UTXO) (stop bool))
	IterateUnlockedUTXOsByAddr(ctx sdk.Context, addr string, cb func(addr string, utxo *types.UTXO) (stop bool))
	IterateUTXOsByTxHash(ctx sdk.Context, hash string, cb func(utxo *types.UTXO) (stop bool))
}

type UTXOKeeper interface {
	UTXOViewKeeper

	LockUTXO(ctx sdk.Context, hash string, vout uint64) error
	LockUTXOs(ctx sdk.Context, utxos []*types.UTXO) error

	UnlockUTXO(ctx sdk.Context, hash string, vout uint64) error
	UnlockUTXOs(ctx sdk.Context, utxos []*types.UTXO) error

	SpendUTXO(ctx sdk.Context, hash string, vout uint64) error
	SpendUTXOs(ctx sdk.Context, utxos []*types.UTXO) error
}

var _ UTXOKeeper = (*BaseUTXOKeeper)(nil)

type BaseUTXOViewKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewBaseUTXOViewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *BaseUTXOViewKeeper {
	return &BaseUTXOViewKeeper{
		cdc,
		storeKey,
	}
}

func (bvk *BaseUTXOViewKeeper) HasUTXO(ctx sdk.Context, hash string, vout uint64) bool {
	store := ctx.KVStore(bvk.storeKey)
	return store.Has(types.BtcUtxoKey(hash, vout))
}

// IsUTXOLocked returns true if the given utxo is locked, false otherwise.
// Note: it returns false if the given utxo does not exist.
func (bvk *BaseUTXOViewKeeper) IsUTXOLocked(ctx sdk.Context, hash string, vout uint64) bool {
	if !bvk.HasUTXO(ctx, hash, vout) {
		return false
	}

	utxo := bvk.GetUTXO(ctx, hash, vout)

	return utxo.IsLocked
}

func (bvk *BaseUTXOViewKeeper) GetUTXO(ctx sdk.Context, hash string, vout uint64) *types.UTXO {
	store := ctx.KVStore(bvk.storeKey)

	var utxo types.UTXO
	bz := store.Get(types.BtcUtxoKey(hash, vout))
	bvk.cdc.MustUnmarshal(bz, &utxo)

	return &utxo
}

func (bvk *BaseUTXOViewKeeper) GetAllUTXOs(ctx sdk.Context) []*types.UTXO {
	utxos := make([]*types.UTXO, 0)

	bvk.IterateAllUTXOs(ctx, func(utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)
		return false
	})

	return utxos
}

func (bvk *BaseUTXOViewKeeper) GetUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO {
	utxos := make([]*types.UTXO, 0)

	bvk.IterateUTXOsByAddr(ctx, addr, func(utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)
		return false
	})

	return utxos
}

func (bvk *BaseUTXOViewKeeper) GetUTXOIteratorByAddr(ctx sdk.Context, addr string) types.UTXOIterator {
	store := ctx.KVStore(bvk.storeKey)

	// get utxo with the minimum amount
	minUTXO := bvk.GetMinimumUTXOInAmount(ctx, addr)

	// iterator in descending order by amount
	iterator := storetypes.KVStoreReversePrefixIterator(store, append(types.BtcOwnerUtxoByAmountKeyPrefix, []byte(addr)...))

	return NewUTXOIterator(ctx, bvk, addr, iterator, minUTXO)
}

func (bvk *BaseUTXOViewKeeper) GetUnlockedUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO {
	utxos := make([]*types.UTXO, 0)

	bvk.IterateUnlockedUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)

		return false
	})

	return utxos
}

// GetUnlockedUTXOsByAddrAndThreshold gets the unlocked utxos that satisfy the maximum threshold by the given address and maximum number
// Note: return all satisfying utxos if the maximum number set to 0
func (bvk *BaseUTXOViewKeeper) GetUnlockedUTXOsByAddrAndThreshold(ctx sdk.Context, addr string, threshold int64, maxNum uint32) []*types.UTXO {
	utxos := make([]*types.UTXO, 0)

	bvk.IterateUnlockedUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		if utxo.Amount > uint64(threshold) {
			return false
		}

		utxos = append(utxos, utxo)

		return maxNum != 0 && len(utxos) >= int(maxNum)
	})

	return utxos
}

// GetMinimumUTXOInAmount gets the utxo with the minimum amount
func (bvk *BaseUTXOViewKeeper) GetMinimumUTXOInAmount(ctx sdk.Context, addr string) *types.UTXO {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.BtcOwnerUtxoByAmountKeyPrefix, []byte(addr)...))
	defer iterator.Close()

	if iterator.Valid() {
		key := iterator.Key()
		prefixLen := 1 + len(addr) + 8

		hash := key[prefixLen : prefixLen+64]
		vout := key[prefixLen+64:]

		return bvk.GetUTXO(ctx, string(hash), sdk.BigEndianToUint64(vout))
	}

	return nil
}

// GetTargetRunesUTXOs gets the unlocked runes utxos targeted by the given params
func (bvk *BaseUTXOViewKeeper) GetTargetRunesUTXOs(ctx sdk.Context, addr string, runeId string, targetAmount uint128.Uint128, maxNum int) ([]*types.UTXO, []*types.RuneBalance) {
	utxos := make([]*types.UTXO, 0)

	totalAmount := uint128.Zero
	totalRuneBalances := make(types.RuneBalances, 0)

	bvk.IterateRunesUTXOsReverse(ctx, addr, runeId, func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)

		totalAmount = totalAmount.Add(amount)
		totalRuneBalances = totalRuneBalances.Merge(utxo.Runes)

		return maxNum != 0 && len(utxos) >= maxNum || totalAmount.Cmp(targetAmount) >= 0
	})

	if totalAmount.Cmp(targetAmount) < 0 {
		return nil, nil
	}

	runeBalancesDelta := totalRuneBalances.Update(runeId, totalAmount.Sub(targetAmount))

	return utxos, runeBalancesDelta
}

// GetTargetRunesUTXOsByAddrAndThreshold gets the unlocked runes utxos that satisfy the maximum threshold for the specified rune id by the given address and maximum number
// Note: return all satisfying utxos if the maximum number set to 0
func (bvk *BaseUTXOViewKeeper) GetTargetRunesUTXOsByAddrAndThreshold(ctx sdk.Context, addr string, runeId string, threshold uint128.Uint128, maxNum uint32) ([]*types.UTXO, []*types.RuneBalance) {
	utxos := make([]*types.UTXO, 0)
	runeBalances := make(types.RuneBalances, 0)

	bvk.IterateRunesUTXOs(ctx, addr, runeId, func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool) {
		if amount.Cmp(threshold) > 0 {
			return false
		}

		utxos = append(utxos, utxo)
		runeBalances = runeBalances.Merge(utxo.Runes)

		return maxNum != 0 && len(utxos) >= int(maxNum)
	})

	return utxos, runeBalances
}

func (bvk *BaseUTXOViewKeeper) GetUnlockedUTXOCountAndBalancesByAddr(ctx sdk.Context, addr string) (uint32, int64, []*types.RuneBalance) {
	count := uint32(0)
	value := int64(0)
	runeBalances := make(types.RuneBalances, 0)

	bvk.IterateUnlockedUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		count += 1
		value += int64(utxo.Amount)
		runeBalances = runeBalances.Merge(utxo.Runes)

		return false
	})

	return count, value, runeBalances
}

func (bvk *BaseUTXOViewKeeper) IterateAllUTXOs(ctx sdk.Context, cb func(utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BtcUtxoKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var utxo types.UTXO
		bvk.cdc.MustUnmarshal(iterator.Value(), &utxo)

		if cb(&utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateUTXOsByAddr(ctx sdk.Context, addr string, cb func(utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.BtcUtxoKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var utxo types.UTXO
		bvk.cdc.MustUnmarshal(iterator.Value(), &utxo)

		if utxo.Address != addr {
			continue
		}

		if cb(&utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateUnlockedUTXOsByAddr(ctx sdk.Context, addr string, cb func(addr string, utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.BtcOwnerUtxoKeyPrefix, []byte(addr)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		hash := string(key[1+len(addr) : 1+len(addr)+64])
		vout := sdk.BigEndianToUint64(key[1+len(addr)+64:])

		utxo := bvk.GetUTXO(ctx, hash, vout)
		if cb(addr, utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateUTXOsByTxHash(ctx sdk.Context, hash string, cb func(utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(types.BtcUtxoKeyPrefix, []byte(hash)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var utxo types.UTXO
		bvk.cdc.MustUnmarshal(iterator.Value(), &utxo)

		if cb(&utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateRunesUTXOs(ctx sdk.Context, addr string, id string, cb func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, append(append(types.BtcOwnerRunesUtxoKeyPrefix, []byte(addr)...), types.MarshalRuneIdFromString(id)...))
	defer iterator.Close()

	prefixLen := 1 + len(addr) + 12

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		amount := types.UnmarshalRuneAmount(key[prefixLen : prefixLen+16])
		hash := string(key[prefixLen+16 : prefixLen+16+64])
		vout := sdk.BigEndianToUint64(key[prefixLen+16+64:])

		utxo := bvk.GetUTXO(ctx, hash, vout)
		if cb(addr, id, amount, utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateRunesUTXOsReverse(ctx sdk.Context, addr string, id string, cb func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := storetypes.KVStoreReversePrefixIterator(store, append(append(types.BtcOwnerRunesUtxoKeyPrefix, []byte(addr)...), types.MarshalRuneIdFromString(id)...))
	defer iterator.Close()

	prefixLen := 1 + len(addr) + 12

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		amount := types.UnmarshalRuneAmount(key[prefixLen : prefixLen+16])
		hash := string(key[prefixLen+16 : prefixLen+16+64])
		vout := sdk.BigEndianToUint64(key[prefixLen+16+64:])

		utxo := bvk.GetUTXO(ctx, hash, vout)
		if cb(addr, id, amount, utxo) {
			break
		}
	}
}

type BaseUTXOKeeper struct {
	BaseUTXOViewKeeper

	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

func NewBaseUTXOKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey) *BaseUTXOKeeper {
	return &BaseUTXOKeeper{
		BaseUTXOViewKeeper: *NewBaseUTXOViewKeeper(cdc, storeKey),
		cdc:                cdc,
		storeKey:           storeKey,
	}
}

func (bk *BaseUTXOKeeper) SetUTXO(ctx sdk.Context, utxo *types.UTXO) {
	store := ctx.KVStore(bk.storeKey)

	bz := bk.cdc.MustMarshal(utxo)
	store.Set(types.BtcUtxoKey(utxo.Txid, utxo.Vout), bz)
}

func (bk *BaseUTXOKeeper) SetOwnerUTXO(ctx sdk.Context, utxo *types.UTXO) {
	store := ctx.KVStore(bk.storeKey)

	store.Set(types.BtcOwnerUtxoKey(utxo.Address, utxo.Txid, utxo.Vout), []byte{})
}

func (bk *BaseUTXOKeeper) SetOwnerUTXOByAmount(ctx sdk.Context, utxo *types.UTXO) {
	store := ctx.KVStore(bk.storeKey)

	store.Set(types.BtcOwnerUtxoByAmountKey(utxo.Address, utxo.Amount, utxo.Txid, utxo.Vout), []byte{})
}

func (bk *BaseUTXOKeeper) SetOwnerRunesUTXO(ctx sdk.Context, utxo *types.UTXO, id string, amount string) {
	store := ctx.KVStore(bk.storeKey)

	store.Set(types.BtcOwnerRunesUtxoKey(utxo.Address, id, amount, utxo.Txid, utxo.Vout), []byte{})
}

func (bk *BaseUTXOKeeper) LockUTXO(ctx sdk.Context, hash string, vout uint64) error {
	if !bk.HasUTXO(ctx, hash, vout) {
		return types.ErrUTXODoesNotExist
	}

	utxo := bk.GetUTXO(ctx, hash, vout)
	if utxo.IsLocked {
		return types.ErrUTXOLocked
	}

	utxo.IsLocked = true
	bk.SetUTXO(ctx, utxo)

	return nil
}

func (bk *BaseUTXOKeeper) LockUTXOs(ctx sdk.Context, utxos []*types.UTXO) error {
	for _, utxo := range utxos {
		if err := bk.LockUTXO(ctx, utxo.Txid, utxo.Vout); err != nil {
			return err
		}
	}

	return nil
}

func (bk *BaseUTXOKeeper) UnlockUTXO(ctx sdk.Context, hash string, vout uint64) error {
	if !bk.HasUTXO(ctx, hash, vout) {
		return types.ErrUTXODoesNotExist
	}

	utxo := bk.GetUTXO(ctx, hash, vout)
	if !utxo.IsLocked {
		return types.ErrUTXOUnlocked
	}

	utxo.IsLocked = false
	bk.SetUTXO(ctx, utxo)

	return nil
}

func (bk *BaseUTXOKeeper) UnlockUTXOs(ctx sdk.Context, utxos []*types.UTXO) error {
	for _, utxo := range utxos {
		if err := bk.UnlockUTXO(ctx, utxo.Txid, utxo.Vout); err != nil {
			return err
		}
	}

	return nil
}

func (bk *BaseUTXOKeeper) SpendUTXO(ctx sdk.Context, hash string, vout uint64) error {
	if !bk.HasUTXO(ctx, hash, vout) {
		return types.ErrUTXODoesNotExist
	}

	bk.removeUTXO(ctx, hash, vout)

	return nil
}

func (bk *BaseUTXOKeeper) SpendUTXOs(ctx sdk.Context, utxos []*types.UTXO) error {
	for _, utxo := range utxos {
		if err := bk.SpendUTXO(ctx, utxo.Txid, utxo.Vout); err != nil {
			return err
		}
	}

	return nil
}

// SaveUTXO saves the given utxo
// Intended to be used out of the module, such as genesis import
func (bk *BaseUTXOKeeper) SaveUTXO(ctx sdk.Context, utxo *types.UTXO) {
	bk.saveUTXO(ctx, utxo)
}

// saveUTXO saves the given utxo
func (bk *BaseUTXOKeeper) saveUTXO(ctx sdk.Context, utxo *types.UTXO) {
	bk.SetUTXO(ctx, utxo)
	bk.SetOwnerUTXO(ctx, utxo)
	bk.SetOwnerUTXOByAmount(ctx, utxo)

	for _, r := range utxo.Runes {
		bk.SetOwnerRunesUTXO(ctx, utxo, r.Id, r.Amount)
	}
}

// removeUTXO deletes the given utxo which is assumed to exist.
func (bk *BaseUTXOKeeper) removeUTXO(ctx sdk.Context, hash string, vout uint64) {
	store := ctx.KVStore(bk.storeKey)
	utxo := bk.GetUTXO(ctx, hash, vout)

	store.Delete(types.BtcUtxoKey(hash, vout))
	store.Delete(types.BtcOwnerUtxoKey(utxo.Address, hash, vout))
	store.Delete(types.BtcOwnerUtxoByAmountKey(utxo.Address, utxo.Amount, hash, vout))

	for _, r := range utxo.Runes {
		store.Delete(types.BtcOwnerRunesUtxoKey(utxo.Address, r.Id, r.Amount, hash, vout))
	}
}

// UTXOIterator implements types.UTXOIterator
// The iterator iterates over utxos by address and amount
type UTXOIterator struct {
	ctx    sdk.Context
	keeper UTXOViewKeeper

	address  string
	iterator storetypes.Iterator

	minUTXO    *types.UTXO
	minUTXOKey []byte

	currentKey []byte
}

func NewUTXOIterator(ctx sdk.Context, keeper UTXOViewKeeper, addr string, iterator storetypes.Iterator, minUTXO *types.UTXO) *UTXOIterator {
	utxoIterator := &UTXOIterator{
		ctx:      ctx,
		keeper:   keeper,
		address:  addr,
		iterator: iterator,
		minUTXO:  minUTXO,
	}

	if minUTXO != nil {
		utxoIterator.minUTXOKey = types.BtcOwnerUtxoByAmountKey(addr, minUTXO.Amount, minUTXO.Txid, minUTXO.Vout)
	}

	return utxoIterator
}

func (i *UTXOIterator) Valid() bool {
	if !i.iterator.Valid() {
		return false
	}

	i.currentKey = i.iterator.Key()

	return !bytes.Equal(i.currentKey, i.minUTXOKey)
}

func (i *UTXOIterator) Next() {
	i.iterator.Next()
}

func (i *UTXOIterator) Close() error {
	return i.iterator.Close()
}

func (i *UTXOIterator) GetUTXO() *types.UTXO {
	key := i.currentKey
	prefixLen := 1 + len(i.address) + 8

	hash := key[prefixLen : prefixLen+64]
	vout := key[prefixLen+64:]

	return i.keeper.GetUTXO(i.ctx, string(hash), sdk.BigEndianToUint64(vout))
}

func (i *UTXOIterator) GetMinimumUTXO() *types.UTXO {
	return i.minUTXO
}
