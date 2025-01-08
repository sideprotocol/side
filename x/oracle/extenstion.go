package oracle

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type OracleVoteExtension struct {
	Height int64
	Prices map[string]math.LegacyDec
}

type VoteExtHandler struct {
	logger          log.Logger
	currentBlock    int64         // current block height
	lastPriceSyncTS time.Time     // last time we synced prices
	providerTimeout time.Duration // timeout for fetching prices from providers
	// providers       map[string]Provider              // mapping of provider name to provider (e.g. Binance -> BinanceProvider)
	// providerPairs   map[string][]keeper.CurrencyPair // mapping of provider name to supported pairs (e.g. Binance -> [ATOM/USD])

	// Keeper keeper.Keeper // keeper of our oracle module
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// here we'd have a helper function that gets all the prices and does a weighted average using the volume of each market
		// prices := h.getAllVolumeWeightedPrices()

		// voteExt := OracleVoteExtension{
		// 	Height: req.Height,
		// 	Prices: prices,
		// }

		bz := []byte{}
		// bz, err := json.Marshal(voteExt)
		// if err != nil {
		// 	return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		// }

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var voteExt OracleVoteExtension
		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if voteExt.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
		}

		// Verify incoming prices from a validator are valid. Note, verification during
		// VerifyVoteExtensionHandler MUST be deterministic. For brevity and demo
		// purposes, we omit implementation.
		// if err := h.verifyOraclePrices(ctx, voteExt.Prices); err != nil {
		// 	return nil, fmt.Errorf("failed to verify oracle prices from validator %X: %w", req.ValidatorAddress, err)
		// }

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

type ProposalHandler struct {
	logger log.Logger
	// keeper   keeper.Keeper // our oracle module keeper
	// valStore baseapp.ValidatorStore // to get the current validators' pubkeys
}
type StakeWeightedPrices struct {
	StakeWeightedPrices map[string]math.LegacyDec
	ExtendedCommitInfo  abci.ExtendedCommitInfo
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		// err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
		// if err != nil {
		//     return nil, err
		// }
		return nil, nil
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		return nil, nil
	}
}

func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return nil, nil
}
