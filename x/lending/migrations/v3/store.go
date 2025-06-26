package v3

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 2 to
// version 3
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	migrateLoans(ctx, storeKey, cdc)

	return nil
}

// migrateLoans performs the loan migration
func migrateLoans(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	loanIterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer loanIterator.Close()

	for ; loanIterator.Valid(); loanIterator.Next() {
		// delete loan
		store.Delete(loanIterator.Key())
	}

	loanByAddriterator := storetypes.KVStorePrefixIterator(store, types.LoanByAddressKeyPrefix)
	defer loanByAddriterator.Close()

	for ; loanByAddriterator.Valid(); loanByAddriterator.Next() {
		// delete loan by address
		store.Delete(loanByAddriterator.Key())
	}
}
