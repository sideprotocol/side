package keeper

import (
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// GetAuthorizationId gets the current authorization id for the specified loan
func (k Keeper) GetAuthorizationId(ctx sdk.Context, loanId string) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.AuthorizationIdKey(loanId))

	return sdk.BigEndianToUint64(bz)
}

// IncrementAuthorizationId increments the authorization id for the specified loan and returns the new one
func (k Keeper) IncrementAuthorizationId(ctx sdk.Context, loanId string) uint64 {
	store := ctx.KVStore(k.storeKey)

	id := k.GetAuthorizationId(ctx, loanId) + 1
	store.Set(types.AuthorizationIdKey(loanId), sdk.Uint64ToBigEndian(id))

	return id
}

// HasAuthorization returns true if the given authorization exists, false otherwise
func (k Keeper) HasAuthorization(ctx sdk.Context, loanId string, id uint64) bool {
	if !k.HasLoan(ctx, loanId) {
		return false
	}

	return id != 0 && id <= uint64(len(k.GetLoan(ctx, loanId).Authorizations))
}

// GetAuthorization gets the specified authorization
func (k Keeper) GetAuthorization(ctx sdk.Context, loanId string, id uint64) *types.Authorization {
	loan := k.GetLoan(ctx, loanId)

	if id > uint64(len(loan.Authorizations)) {
		return nil
	}

	return &loan.Authorizations[id-1]
}

// GetDeposits gets deposit details for the given authorization
func (k Keeper) GetDeposits(ctx sdk.Context, authorization *types.Authorization) []*types.DepositLog {
	deposits := []*types.DepositLog{}

	for _, depositTx := range authorization.DepositTxs {
		deposits = append(deposits, k.GetDepositLog(ctx, depositTx))
	}

	return deposits
}

// DepositsVerified returns true if all deposit txs verified for the given authorization, false otherwise
func (k Keeper) DepositsVerified(ctx sdk.Context, authorization *types.Authorization) bool {
	store := ctx.KVStore(k.storeKey)

	for _, depositTx := range authorization.DepositTxs {
		var depositLog types.DepositLog
		bz := store.Get(types.DepositLogKey(depositTx))
		k.cdc.MustUnmarshal(bz, &depositLog)

		if depositLog.Status != types.DepositStatus_DEPOSIT_STATUS_VERIFIED {
			return false
		}
	}

	return true
}

// HasDeposit returns true if the given deposit tx exists in the specified authorization, false otherwise
func (k Keeper) HasDeposit(ctx sdk.Context, depositTxHash string, authorization *types.Authorization) bool {
	return slices.Contains(authorization.DepositTxs, depositTxHash)
}

// CreateAuthorization creates a new authorization from the given loan id and deposit txs
func (k Keeper) CreateAuthorization(ctx sdk.Context, loanId string, depositTxHashes []string) *types.Authorization {
	return &types.Authorization{
		Id:         k.IncrementAuthorizationId(ctx, loanId),
		DepositTxs: depositTxHashes,
		Status:     types.AuthorizationStatus_AUTHORIZATION_STATUS_PENDING,
	}
}

// AddAuthorization adds a new authorization to the given loan
func (k Keeper) AddAuthorization(ctx sdk.Context, loan *types.Loan, depositTxHash string, status types.AuthorizationStatus) {
	loan.Authorizations = append(loan.Authorizations, types.Authorization{
		Id:         uint64(len(loan.Authorizations)) + 1,
		DepositTxs: []string{depositTxHash},
		Status:     status,
	})

	k.SetLoan(ctx, loan)
}

// UpdateAuthorization updates the specified authorization
// Assume that the given authorization exists
func (k Keeper) UpdateAuthorization(ctx sdk.Context, loan *types.Loan, authorizationId uint64, depositTxHash string, status types.AuthorizationStatus) {
	authorization := loan.Authorizations[authorizationId-1]

	if !k.HasDeposit(ctx, depositTxHash, &authorization) {
		loan.Authorizations[authorizationId-1].DepositTxs = append(loan.Authorizations[authorizationId-1].DepositTxs, depositTxHash)
	}

	loan.Authorizations[authorizationId-1].Status = status

	k.SetLoan(ctx, loan)
}
