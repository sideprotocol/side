package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BiddedAssetHandler defines the handler to perform bidded asset handling
type BiddedAssetHandler func(ctx sdk.Context, loanId string, moduleAccount string, asset sdk.Coin) error
