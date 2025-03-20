package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// SetLoan sets the given loan
func (k Keeper) SetLoan(ctx sdk.Context, loan *types.Loan) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(loan)

	store.Set(types.LoanStoreKey(loan.VaultAddress), bz)
}

// HasLoan returns true if the given loan exists, false otherwise
func (k Keeper) HasLoan(ctx sdk.Context, vault string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.LoanStoreKey(vault))
}

// GetLoan gets the given loan
func (k Keeper) GetLoan(ctx sdk.Context, vault string) *types.Loan {
	store := ctx.KVStore(k.storeKey)

	var loan types.Loan
	bz := store.Get(types.LoanStoreKey(vault))
	k.cdc.MustUnmarshal(bz, &loan)

	return &loan
}

// IterateLoans iterates through all loans
func (k Keeper) IterateLoans(ctx sdk.Context, cb func(loan *types.Loan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanStorePrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var loan types.Loan
		k.cdc.MustUnmarshal(iterator.Value(), &loan)

		if cb(&loan) {
			break
		}
	}
}

// GetLoans gets loans by the given status
func (k Keeper) GetLoans(ctx sdk.Context, status types.LoanStatus) []*types.Loan {
	var loans []*types.Loan

	k.IterateLoans(ctx, func(loan *types.Loan) (stop bool) {
		if loan.Status == status {
			loans = append(loans, loan)
		}

		return false
	})

	return loans
}

// GetAllLoans returns all loans
func (k Keeper) GetAllLoans(ctx sdk.Context) []*types.Loan {
	var loans []*types.Loan

	k.IterateLoans(ctx, func(loan *types.Loan) (stop bool) {
		loans = append(loans, loan)
		return false
	})

	return loans
}

// GetLoansByAddress gets loans by the given address and status
func (k Keeper) GetLoansByAddress(ctx sdk.Context, address string, status types.LoanStatus) []*types.Loan {
	var loans []*types.Loan

	k.IterateLoans(ctx, func(loan *types.Loan) (stop bool) {
		if loan.Borrower == address && (status == types.LoanStatus_Unspecified || loan.Status == status) {
			loans = append(loans, loan)
		}

		return false
	})

	return loans
}

// SetDepositLog sets the given deposit log
func (k Keeper) SetDepositLog(ctx sdk.Context, depositLog *types.DepositLog) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(depositLog)

	store.Set(types.DepositLogKey(depositLog.Txid), bz)
}

// HasDepositLog returns true if the given deposit log exists, false otherwise
func (k Keeper) HasDepositLog(ctx sdk.Context, txid string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.DepositLogKey(txid))
}

// GetDepositLog gets the given deposit log
func (k Keeper) GetDepositLog(ctx sdk.Context, txid string) *types.DepositLog {
	store := ctx.KVStore(k.storeKey)

	var depositLog types.DepositLog
	bz := store.Get(types.DepositLogKey(txid))
	k.cdc.MustUnmarshal(bz, &depositLog)

	return &depositLog
}

// SetRepayment sets the given repayment
func (k Keeper) SetRepayment(ctx sdk.Context, repayment *types.Repayment) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(repayment)

	store.Set(types.RepaymentKey(repayment.LoanId), bz)
}

// HasRepayment returns true if the given repayment exists, false otherwise
func (k Keeper) HasRepayment(ctx sdk.Context, loanId string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.RepaymentKey(loanId))
}

// GetRepayment gets the given repayment
func (k Keeper) GetRepayment(ctx sdk.Context, loanId string) *types.Repayment {
	store := ctx.KVStore(k.storeKey)

	var repayment types.Repayment
	bz := store.Get(types.RepaymentKey(loanId))
	k.cdc.MustUnmarshal(bz, &repayment)

	return &repayment
}

// HasCancellation returns true if there exists cancellation for the given loan, false otherwise
func (k Keeper) HasCancellation(ctx sdk.Context, loanId string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.CancellationKey(loanId))
}

// SetCancellation sets the given cancellation
func (k Keeper) SetCancellation(ctx sdk.Context, cancellation *types.Cancellation) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(cancellation)

	store.Set(types.CancellationKey(cancellation.LoanId), bz)
}

// GetCancellation gets the specified cancellation
func (k Keeper) GetCancellation(ctx sdk.Context, loanId string) *types.Cancellation {
	store := ctx.KVStore(k.storeKey)

	var cancellation types.Cancellation
	bz := store.Get(types.CancellationKey(loanId))
	k.cdc.MustUnmarshal(bz, &cancellation)

	return &cancellation
}
