package v5

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// MigrateStore migrates the x/lending module state from the consensus version 4 to
// version 5
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, dlcKeeper types.DLCKeeper, cdc codec.BinaryCodec) error {
	migrateLoans(ctx, storeKey, dlcKeeper, cdc)

	return nil
}

// migrateLoans performs the loans migration
func migrateLoans(ctx sdk.Context, storeKey storetypes.StoreKey, dlcKeeper types.DLCKeeper, cdc codec.BinaryCodec) {
	store := ctx.KVStore(storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var loan types.Loan
		cdc.MustUnmarshal(iterator.Value(), &loan)

		dlcEvent := dlcKeeper.GetEvent(ctx, loan.DlcEventId)

		// set loan by status
		store.Set(types.LoanByStatusKey(loan.Status, loan.VaultAddress), []byte{})

		// set loan by oracle
		store.Set(types.LoanByOracleKey(dlcEvent.Pubkey, loan.VaultAddress), []byte{})
	}
}
