package keeper

import (
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

// CreateAuthorization creates a new authorization from the given loan id and deposit txs
func (k Keeper) CreateAuthorization(ctx sdk.Context, loanId string, depositTxHashes []string) *types.Authorization {
	return &types.Authorization{
		Id:         k.IncrementAuthorizationId(ctx, loanId),
		DepositTxs: depositTxHashes,
		Status:     types.AuthorizationStatus_AUTHORIZATION_STATUS_PENDING,
	}
}
