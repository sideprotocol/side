package keeper

import (
	"math/big"
	"sort"

	"lukechampine.com/uint128"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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
	GetOrderedUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO

	GetTargetRunesUTXOs(ctx sdk.Context, addr string, runeId string, targetAmount uint128.Uint128) ([]*types.UTXO, []*types.RuneBalance)

	IterateAllUTXOs(ctx sdk.Context, cb func(utxo *types.UTXO) (stop bool))
	IterateUTXOsByAddr(ctx sdk.Context, addr string, cb func(addr string, utxo *types.UTXO) (stop bool))
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

	bvk.IterateUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		utxos = append(utxos, utxo)
		return false
	})

	return utxos
}

func (bvk *BaseUTXOViewKeeper) GetUTXOIteratorByAddr(ctx sdk.Context, addr string) types.UTXOIterator {
	store := ctx.KVStore(bvk.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, append(types.BtcOwnerUtxoKeyPrefix, []byte(addr)...))

	return &UTXOIterator{
		ctx:      ctx,
		keeper:   bvk,
		iterator: iterator,
		address:  addr,
	}
}

func (bvk *BaseUTXOViewKeeper) GetUnlockedUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO {
	utxos := make([]*types.UTXO, 0)

	bvk.IterateUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		if !utxo.IsLocked {
			utxos = append(utxos, utxo)
		}

		return false
	})

	return utxos
}

// GetOrderedUTXOsByAddr gets all unlocked utxos of the given address in the descending order by amount
// Note: high gas is required due to sorting
func (bvk *BaseUTXOViewKeeper) GetOrderedUTXOsByAddr(ctx sdk.Context, addr string) []*types.UTXO {
	// get unlocked utxos
	utxos := bvk.GetUnlockedUTXOsByAddr(ctx, addr)

	if len(utxos) > 1 {
		// sort utxos in the descending order
		sort.SliceStable(utxos, func(i int, j int) bool {
			return utxos[i].Amount > utxos[j].Amount
		})
	}

	return utxos
}

// GetTargetRunesUTXOs gets the unlocked runes utxos targeted by the given params
func (bvk *BaseUTXOViewKeeper) GetTargetRunesUTXOs(ctx sdk.Context, addr string, runeId string, targetAmount uint128.Uint128) ([]*types.UTXO, []*types.RuneBalance) {
	utxos := make([]*types.UTXO, 0)

	totalAmount := uint128.Zero
	totalRuneBalances := make(types.RuneBalances, 0)

	bvk.IterateRunesUTXOs(ctx, addr, runeId, func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool) {
		if utxo.IsLocked {
			return false
		}

		utxos = append(utxos, utxo)

		totalAmount = totalAmount.Add(amount)
		totalRuneBalances = append(totalRuneBalances, utxo.Runes...)

		return totalAmount.Cmp(targetAmount) >= 0
	})

	if totalAmount.Cmp(targetAmount) < 0 {
		return nil, nil
	}

	runeBalancesDelta := totalRuneBalances.Compact().Update(runeId, totalAmount.Sub(targetAmount))

	return utxos, runeBalancesDelta
}

func (bvk *BaseUTXOViewKeeper) GetUnlockedUTXOCountAndBalancesByAddr(ctx sdk.Context, addr string) (uint32, int64, []*types.RuneBalance) {
	count := uint32(0)
	value := int64(0)
	runeBalances := make(types.RuneBalances, 0)

	bvk.IterateUTXOsByAddr(ctx, addr, func(addr string, utxo *types.UTXO) (stop bool) {
		if !utxo.IsLocked {
			count += 1
			value += int64(utxo.Amount)
			runeBalances = append(runeBalances, utxo.Runes...)
		}

		return false
	})

	return count, value, runeBalances.Compact()
}

func (bvk *BaseUTXOViewKeeper) IterateAllUTXOs(ctx sdk.Context, cb func(utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.BtcUtxoKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var utxo types.UTXO
		bvk.cdc.MustUnmarshal(iterator.Value(), &utxo)

		if cb(&utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateUTXOsByAddr(ctx sdk.Context, addr string, cb func(addr string, utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(types.BtcOwnerUtxoKeyPrefix, []byte(addr)...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		hash := key[1+len(addr) : 1+len(addr)+64]
		vout := key[1+len(addr)+64:]

		utxo := bvk.GetUTXO(ctx, string(hash), new(big.Int).SetBytes(vout).Uint64())
		if cb(addr, utxo) {
			break
		}
	}
}

func (bvk *BaseUTXOViewKeeper) IterateRunesUTXOs(ctx sdk.Context, addr string, id string, cb func(addr string, id string, amount uint128.Uint128, utxo *types.UTXO) (stop bool)) {
	store := ctx.KVStore(bvk.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, append(append(types.BtcOwnerRunesUtxoKeyPrefix, []byte(addr)...), []byte(id)...))
	defer iterator.Close()

	prefixLen := 1 + len(addr) + len(id)

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		value := iterator.Value()

		hash := key[prefixLen : prefixLen+64]
		vout := key[prefixLen+64:]

		amount := types.RuneAmountFromString(string(value))

		utxo := bvk.GetUTXO(ctx, string(hash), new(big.Int).SetBytes(vout).Uint64())
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

func (bk *BaseUTXOKeeper) SetOwnerRunesUTXO(ctx sdk.Context, utxo *types.UTXO, id string, amount string) {
	store := ctx.KVStore(bk.storeKey)

	store.Set(types.BtcOwnerRunesUtxoKey(utxo.Address, id, utxo.Txid, utxo.Vout), []byte(amount))
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

	for _, r := range utxo.Runes {
		store.Delete(types.BtcOwnerRunesUtxoKey(utxo.Address, r.Id, hash, vout))
	}
}

// UTXOIterator defines the iterator over utxos by address
type UTXOIterator struct {
	ctx    sdk.Context
	keeper UTXOViewKeeper

	iterator sdk.Iterator
	address  string
}

func (i *UTXOIterator) Valid() bool {
	return i.iterator.Valid()
}

func (i *UTXOIterator) Next() {
	i.iterator.Next()
}

func (i *UTXOIterator) Close() error {
	return i.iterator.Close()
}

func (i *UTXOIterator) GetUTXO() *types.UTXO {
	key := i.iterator.Key()

	hash := key[1+len(i.address) : 1+len(i.address)+64]
	vout := key[1+len(i.address)+64:]

	return i.keeper.GetUTXO(i.ctx, string(hash), new(big.Int).SetBytes(vout).Uint64())
}
