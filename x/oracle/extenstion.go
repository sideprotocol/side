package oracle

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/oracle/types"
)

type OracleVoteExtension struct {
	Height int64
	Prices map[string]math.LegacyDec
}

type VoteExtHandler struct {
	logger          log.Logger
	currentBlock    int64 // current block height
	lastPriceSyncTS int64 // last time we synced prices
	// providerTimeout time.Duration // timeout for fetching prices from providers
	// providers       map[string]Provider              // mapping of provider name to provider (e.g. Binance -> BinanceProvider)
	// providerPairs   map[string][]keeper.CurrencyPair // mapping of provider name to supported pairs (e.g. Binance -> [ATOM/USD])

	// Keeper keeper.Keeper // keeper of our oracle module
}

func NewVoteExtHandler(logger log.Logger) VoteExtHandler {
	return VoteExtHandler{
		logger:       logger,
		currentBlock: 0,
	}
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// here we'd have a helper function that gets all the prices and does a weighted average using the volume of each market

		types.CleanPrices(h.lastPriceSyncTS)
		prices := h.getAllVolumeWeightedPrices()
		h.lastPriceSyncTS = req.Time.UnixMilli()

		voteExt := OracleVoteExtension{
			Height: req.Height,
			Prices: prices,
		}

		// bz := []byte{}
		bz, err := json.Marshal(voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {

		h.logger.Info("VerifyVoteExtensionHandler", "height", req.Height, "validator", hex.EncodeToString(req.ValidatorAddress))
		var voteExt OracleVoteExtension
		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if voteExt.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func (h *VoteExtHandler) getAllVolumeWeightedPrices() map[string]math.LegacyDec {

	for _, v := range types.PRICE_CACHE {

		output := make(map[string]int)
		for k1, v1 := range v {
			output[k1] = len(v1)
		}
		h.logger.Info("Current Price Cache", "cache", output)
	}

	// calculate the weighted average
	symbolPrices := make(map[string][]math.LegacyDec)
	for symbol, pairs := range types.PRICE_CACHE {
		providers := []string{}
		prices := []string{}
		for ex, price := range pairs {
			if len(price) > 0 {
				p, err := math.LegacyNewDecFromStr(price[0].Price)
				if err == nil { // TODO fitler price by time
					symbolPrices[symbol] = append(symbolPrices[symbol], p)
					providers = append(providers, ex)
					prices = append(prices, price[0].Price)
				}
			}

		}
		h.logger.Info("fetch price", "providers", providers, "price", prices)
	}

	avgPrices := make(map[string]math.LegacyDec)
	for symbol, prices := range symbolPrices {
		if len(prices) > 0 {
			sum := math.LegacyNewDec(0)
			for _, p := range prices {
				sum = sum.Add(p)
			}
			if avg := sum.QuoInt64(int64(len(prices))); avg.GT(math.LegacyNewDec(0)) {
				avgPrices[symbol] = avg
			}
		}
	}

	h.logger.Info("AvgPrice", "prices", avgPrices)

	return avgPrices
}

type ProposalHandler struct {
	logger log.Logger
	// keeper   keeper.Keeper // our oracle module keeper
	valStore baseapp.ValidatorStore // to get the current validators' pubkeys
}

func NewProposalHandler(logger log.Logger, valStore baseapp.ValidatorStore) ProposalHandler {
	return ProposalHandler{
		logger:   logger,
		valStore: valStore,
	}
}

type StakeWeightedPrices struct {
	StakeWeightedPrices map[string]math.LegacyDec
	ExtendedCommitInfo  abci.ExtendedCommitInfo
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		proposalTxs := req.Txs

		if req.Height >= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight && ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight != 0 {
			// if req.Height >= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, req.LocalLastCommit)
			if err != nil {
				return nil, errors.New("failed to compute stake-weighted oracle prices")
			}

			injectedVoteExtTx := StakeWeightedPrices{
				StakeWeightedPrices: stakeWeightedPrices,
				ExtendedCommitInfo:  req.LocalLastCommit,
			}

			// NOTE: We use stdlib JSON encoding, but an application may choose to use
			// a performant mechanism. This is for demo purposes only.
			bz, err := json.Marshal(injectedVoteExtTx)
			if err != nil {
				h.logger.Error("failed to encode injected vote extension tx", "err", err)
				return nil, errors.New("failed to encode injected vote extension tx")
			}

			// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
			// and store the canonical stake-weighted average prices.
			proposalTxs = append([][]byte{bz}, proposalTxs...)
		}

		// proceed with normal block proposal construction, e.g. POB, normal txs, etc...

		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) computeStakeWeightedOraclePrices(ctx sdk.Context, commit abci.ExtendedCommitInfo) (map[string]math.LegacyDec, error) {
	// requiredPairs := h.keeper.GetSupportedPairs(ctx)
	// requiredPairs := []string{"BTCUSD"}
	stakeWeightedPrices := make(map[string]math.LegacyDec, len(types.PRICE_CACHE)) // base -> average stake-weighted price
	// for _, pair := range requiredPairs {
	// 	stakeWeightedPrices[pair] = math.LegacyZeroDec()
	// }

	var totalStake int64
	for _, v := range commit.Votes {
		if v.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		var voteExt OracleVoteExtension
		if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
			h.logger.Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
			return nil, err
		}

		totalStake += v.Validator.Power

		// Compute stake-weighted average of prices for each supported pair, i.e.
		// (P1)(W1) + (P2)(W2) + ... + (Pn)(Wn) / (W1 + W2 + ... + Wn)
		//
		// NOTE: These are the prices computed at the PREVIOUS height, i.e. H-1
		for base, price := range voteExt.Prices {
			// Only compute stake-weighted average for supported pairs.
			//
			// NOTE: VerifyVoteExtension should be sufficient to ensure that only
			// supported pairs are supplied, but we add this here for demo purposes.
			if _, ok := stakeWeightedPrices[base]; ok {
				stakeWeightedPrices[base] = stakeWeightedPrices[base].Add(price.MulInt64(v.Validator.Power))
			} else {
				stakeWeightedPrices[base] = price.MulInt64(v.Validator.Power)
			}
		}
	}

	if totalStake == 0 {
		return nil, nil
	}

	// finalize average by dividing by total stake, i.e. total weights
	for base, price := range stakeWeightedPrices {
		stakeWeightedPrices[base] = price.QuoInt64(totalStake)
	}

	return stakeWeightedPrices, nil
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		var injectedVoteExtTx StakeWeightedPrices
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			return nil, err
		}

		// Verify the proposer's stake-weighted oracle prices by computing the same
		// calculation and comparing the results. We omit verification for brevity
		// and demo purposes.
		stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		if err := compareOraclePrices(injectedVoteExtTx.StakeWeightedPrices, stakeWeightedPrices); err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {

	res := &sdk.ResponsePreBlock{}
	if len(req.Txs) == 0 {
		return res, nil
	}

	var injectedVoteExtTx StakeWeightedPrices
	if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
		h.logger.Error("failed to decode injected vote extension tx", "err", err)
		return nil, err
	}

	// set oracle prices using the passed in context, which will make these prices available in the current block
	// if err := h.keeper.SetOraclePrices(ctx, injectedVoteExtTx.StakeWeightedPrices); err != nil {
	// 	return nil, err
	// }

	h.logger.Warn("Oracle Final States", "price", injectedVoteExtTx.StakeWeightedPrices)

	return res, nil
}

func compareOraclePrices(p1, p2 map[string]math.LegacyDec) error {
	if len(p1) != len(p2) {
		return fmt.Errorf("price maps are different, length %s != length %s", slices.Collect(maps.Keys(p1)), slices.Collect(maps.Keys(p2)))
	}
	for k, v := range p1 {
		if v2, ok := p2[k]; !ok || !v.Equal(v2) {
			return fmt.Errorf("[%s] prices are different, %s!=%s", k, v, v2)
		}
	}
	return nil
}
