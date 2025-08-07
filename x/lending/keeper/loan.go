package keeper

import (
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/sideprotocol/side/x/lending/types"
)

// SetLoan sets the given loan
func (k Keeper) SetLoan(ctx sdk.Context, loan *types.Loan) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(loan)

	k.SetLoanStatus(ctx, loan.VaultAddress, loan.Status)

	store.Set(types.LoanKey(loan.VaultAddress), bz)
}

// SetLoanByAddress sets the given loan by address
func (k Keeper) SetLoanByAddress(ctx sdk.Context, loan *types.Loan) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.LoanByAddressKey(loan.VaultAddress, loan.Borrower), []byte{})
}

// SetLoanByOracle sets the given loan by oracle
func (k Keeper) SetLoanByOracle(ctx sdk.Context, id string, oraclePubKey string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.LoanByOracleKey(oraclePubKey, id), []byte{})
}

// SetLoanStatus sets the status store of the given loan
func (k Keeper) SetLoanStatus(ctx sdk.Context, id string, status types.LoanStatus) {
	store := ctx.KVStore(k.storeKey)

	if k.HasLoan(ctx, id) {
		k.RemoveLoanStatus(ctx, id)
	}

	store.Set(types.LoanByStatusKey(status, id), []byte{})
}

// RemoveLoanStatus removes the status store of the given loan
func (k Keeper) RemoveLoanStatus(ctx sdk.Context, id string) {
	store := ctx.KVStore(k.storeKey)

	loan := k.GetLoan(ctx, id)

	store.Delete(types.LoanByStatusKey(loan.Status, id))
}

// AddToLiquidationQueue adds the given loan to the liquidation queue
func (k Keeper) AddToLiquidationQueue(ctx sdk.Context, loanId string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.LiquidationQueueKey(loanId), []byte{})
}

// RemoveFromLiquidationQueue removes the given loan from the liquidation queue
func (k Keeper) RemoveFromLiquidationQueue(ctx sdk.Context, loanId string) {
	store := ctx.KVStore(k.storeKey)

	store.Delete(types.LiquidationQueueKey(loanId))
}

// HasLoan returns true if the given loan exists, false otherwise
func (k Keeper) HasLoan(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.LoanKey(id))
}

// GetLoan gets the given loan
func (k Keeper) GetLoan(ctx sdk.Context, id string) *types.Loan {
	store := ctx.KVStore(k.storeKey)

	var loan types.Loan
	bz := store.Get(types.LoanKey(id))
	k.cdc.MustUnmarshal(bz, &loan)

	return &loan
}

// GetLoans gets loans by the given status
func (k Keeper) GetLoans(ctx sdk.Context, status types.LoanStatus) []*types.Loan {
	var loans []*types.Loan

	k.IterateLoansByStatus(ctx, status, func(loan *types.Loan) (stop bool) {
		loans = append(loans, loan)
		return false
	})

	return loans
}

// GetLoansByStatusWithPagination gets loans by the given status with pagination
func (k Keeper) GetLoansByStatusWithPagination(ctx sdk.Context, status types.LoanStatus, pagination *query.PageRequest) ([]*types.Loan, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	loanStatusStore := prefix.NewStore(store, append(types.LoanByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...))

	var loans []*types.Loan

	pageRes, err := query.Paginate(loanStatusStore, pagination, func(key []byte, value []byte) error {
		id := string(key)
		loan := k.GetLoan(ctx, id)

		loans = append(loans, loan)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return loans, pageRes, nil
}

// GetLoansWithPagination gets loans with pagination
func (k Keeper) GetLoansWithPagination(ctx sdk.Context, pagination *query.PageRequest) ([]*types.Loan, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	loanStore := prefix.NewStore(store, types.LoanKeyPrefix)

	var loans []*types.Loan

	pageRes, err := query.Paginate(loanStore, pagination, func(key []byte, value []byte) error {
		var loan types.Loan
		k.cdc.MustUnmarshal(value, &loan)

		loans = append(loans, &loan)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return loans, pageRes, nil
}

// GetLoansByAddress gets loans by the given address and status with pagination
func (k Keeper) GetLoansByAddress(ctx sdk.Context, address string, status types.LoanStatus, pagination *query.PageRequest) ([]*types.Loan, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	loanByAddressStore := prefix.NewStore(store, append(types.LoanByAddressKeyPrefix, []byte(address)...))

	var loans []*types.Loan

	pageRes, err := query.Paginate(loanByAddressStore, pagination, func(key []byte, value []byte) error {
		id := string(key)
		loan := k.GetLoan(ctx, id)

		if status == types.LoanStatus_Unspecified || loan.Status == status {
			loans = append(loans, loan)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return loans, pageRes, nil
}

// GetLoansByOracle gets loans by the given oracle with pagination
func (k Keeper) GetLoansByOracle(ctx sdk.Context, oraclePubKey []byte, pagination *query.PageRequest) ([]*types.Loan, *query.PageResponse, error) {
	store := ctx.KVStore(k.storeKey)
	loanByOracleStore := prefix.NewStore(store, append(types.LoanByOracleKeyPrefix, oraclePubKey...))

	var loans []*types.Loan

	pageRes, err := query.Paginate(loanByOracleStore, pagination, func(key []byte, value []byte) error {
		id := string(key)
		loan := k.GetLoan(ctx, id)

		loans = append(loans, loan)

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return loans, pageRes, nil
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

// IterateLoans iterates through all loans
func (k Keeper) IterateLoans(ctx sdk.Context, cb func(loan *types.Loan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LoanKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var loan types.Loan
		k.cdc.MustUnmarshal(iterator.Value(), &loan)

		if cb(&loan) {
			break
		}
	}
}

// IterateLoansByStatus iterates through loans by the given status
func (k Keeper) IterateLoansByStatus(ctx sdk.Context, status types.LoanStatus, cb func(loan *types.Loan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	keyPrefix := append(types.LoanByStatusKeyPrefix, sdk.Uint64ToBigEndian(uint64(status))...)

	iterator := storetypes.KVStorePrefixIterator(store, keyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		id := string(key[len(keyPrefix):])
		loan := k.GetLoan(ctx, id)

		if cb(loan) {
			break
		}
	}
}

// IterateLiquidationQueue iterates through the liquidation queue
func (k Keeper) IterateLiquidationQueue(ctx sdk.Context, cb func(loan *types.Loan) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.LiquidationQueueKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		id := string(iterator.Key()[1:])
		loan := k.GetLoan(ctx, id)

		if cb(loan) {
			break
		}
	}
}

// SetDepositLog sets the given deposit log
func (k Keeper) SetDepositLog(ctx sdk.Context, depositLog *types.DepositLog) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(depositLog)

	store.Set(types.DepositLogKey(depositLog.Txid), bz)
}

// NewDepositLog creates a new deposit log according to the given params
func (k Keeper) NewDepositLog(ctx sdk.Context, txid string, vault string, authorizationId uint64, tx string) {
	depositLog := &types.DepositLog{
		Txid:            txid,
		VaultAddress:    vault,
		AuthorizationId: authorizationId,
		DepositTx:       tx,
		Status:          types.DepositStatus_DEPOSIT_STATUS_PENDING,
	}

	k.SetDepositLog(ctx, depositLog)
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

// GetDepositLogs gets deposit logs by the given loan
func (k Keeper) GetDepositLogs(ctx sdk.Context, loanId string) []*types.DepositLog {
	var depositLogs []*types.DepositLog

	k.IterateDepositLogs(ctx, func(depositLog *types.DepositLog) (stop bool) {
		if depositLog.VaultAddress == loanId {
			depositLogs = append(depositLogs, depositLog)
		}

		return false
	})

	return depositLogs
}

// IterateDepositLogs iterates through all deposit logs
func (k Keeper) IterateDepositLogs(ctx sdk.Context, cb func(depositLog *types.DepositLog) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.DepositLogKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var depositLog types.DepositLog
		k.cdc.MustUnmarshal(iterator.Value(), &depositLog)

		if cb(&depositLog) {
			break
		}
	}
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

// GetLiquidationPrice gets the liquidation price of the given loan according to the specified collateral amount
func (k Keeper) GetLiquidationPrice(ctx sdk.Context, loan *types.Loan, collateralAmount sdkmath.Int) sdkmath.LegacyDec {
	pool := k.GetPool(ctx, loan.PoolId)

	collateralDecimals := int(pool.Config.CollateralAsset.Decimals)
	borrowDecimals := int(pool.Config.LendingAsset.Decimals)
	collateralIsBaseAsset := pool.Config.CollateralAsset.IsBasePriceAsset

	return types.GetLiquidationPrice(collateralAmount, collateralDecimals, loan.BorrowAmount.Amount, borrowDecimals, loan.Maturity, loan.BorrowAPR, k.GetBlocksPerYear(ctx), pool.Config.LiquidationThreshold, collateralIsBaseAsset)
}

// GetCurrentBorrowIndex gets the current borrow index of the given loan
// Assume that the loan maturity exists in the pool tranches
func (k Keeper) GetCurrentBorrowIndex(ctx sdk.Context, loan *types.Loan) sdkmath.LegacyDec {
	tranche, _ := types.GetTranche(k.GetPool(ctx, loan.PoolId).Tranches, loan.Maturity)

	return tranche.BorrowIndex
}

// GetCurrentInterest gets the current interest of the given loan
func (k Keeper) GetCurrentInterest(ctx sdk.Context, loan *types.Loan) sdk.Coin {
	var interest sdkmath.Int

	switch loan.Status {
	case types.LoanStatus_Open:
		interest = types.GetInterest(loan.BorrowAmount.Amount, loan.StartBorrowIndex, k.GetCurrentBorrowIndex(ctx, loan))

	case types.LoanStatus_Repaid, types.LoanStatus_Closed:
		repayment := k.GetRepayment(ctx, loan.VaultAddress)
		interest = repayment.Amount.Sub(loan.BorrowAmount).Amount

	case types.LoanStatus_Defaulted:
		interest = loan.Interest

	case types.LoanStatus_Liquidated:
		liquidation := k.liquidationKeeper.GetLiquidation(ctx, loan.LiquidationId)
		interest = liquidation.DebtAmount.Amount.Sub(loan.BorrowAmount.Amount)

	default:
		interest = sdkmath.ZeroInt()
	}

	return sdk.NewCoin(loan.BorrowAmount.Denom, interest)
}
