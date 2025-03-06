package abci

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sideprotocol/side/x/oracle/keeper"
	"github.com/sideprotocol/side/x/oracle/types"
)

type PriceOracleVoteExtHandler struct {
	valStore        baseapp.ValidatorStore // to get the current validators' pubkeys
	logger          log.Logger
	currentBlock    int64 // current block height
	lastPriceSyncTS int64 // last time we synced prices
	bitcoinClient   *rpcclient.Client
	// providerTimeout time.Duration // timeout for fetching prices from providers
	// providers       map[string]Provider              // mapping of provider name to provider (e.g. Binance -> BinanceProvider)
	// providerPairs   map[string][]keeper.CurrencyPair // mapping of provider name to supported pairs (e.g. Binance -> [ATOM/USD])

	Keeper keeper.Keeper // keeper of our oracle module
	config *types.OracleConfig
}

func NewPriceOracleVoteExtHandler(logger log.Logger, valStore baseapp.ValidatorStore, oracleKeeper keeper.Keeper, config *types.OracleConfig) PriceOracleVoteExtHandler {
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         config.BitcoinRpc,
		User:         config.BitcoinRpcUser,
		Pass:         config.BitcoinRpcPass,
		HTTPPostMode: config.HTTPPostMode,
		DisableTLS:   config.DisableTLS,
	}, nil)
	if err != nil {
		panic("unable to create bitcoin rpc")
	}

	return PriceOracleVoteExtHandler{
		logger:        logger,
		currentBlock:  0,
		valStore:      valStore,
		Keeper:        oracleKeeper,
		bitcoinClient: client,
		config:        config,
	}
}

func (h *PriceOracleVoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		// here we'd have a helper function that gets all the prices and does a weighted average using the volume of each market

		types.CleanPrices(h.lastPriceSyncTS)
		prices := h.getAllVolumeWeightedPrices()
		h.lastPriceSyncTS = req.Time.UnixMilli()

		headers, err := h.getBitcoinHeaders()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bitcoin headers: %w", err)
		}
		voteExt := types.OracleVoteExtension{
			Height: req.Height,
			Prices: prices,
			Blocks: headers,
		}

		// bz := []byte{}
		// bz, err := json.Marshal(voteExt)
		bz, err := voteExt.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *PriceOracleVoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {

		h.logger.Info("VerifyVoteExtensionHandler", "height", req.Height, "validator", hex.EncodeToString(req.ValidatorAddress))
		var voteExt types.OracleVoteExtension
		// err := json.Unmarshal(req.VoteExtension, &voteExt)
		err := voteExt.Unmarshal(req.VoteExtension)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if voteExt.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func (h *PriceOracleVoteExtHandler) getBitcoinHeaders() ([]*types.BlockHeader, error) {
	tips, err := h.bitcoinClient.GetChainTips()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch best block header: %w", err)
	}

	hash, err := chainhash.NewHashFromStr(tips[0].Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch best block header: %w", err)
	}
	b, err := h.bitcoinClient.GetBlockHeaderVerbose(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch header: %w", err)
	}

	header := types.BlockHeader{
		Version:           b.Version,
		Hash:              b.Hash,
		Height:            b.Height,
		PreviousBlockHash: b.PreviousHash,
		MerkleRoot:        b.MerkleRoot,
		Nonce:             b.Nonce,
		Bits:              b.Bits,
		Time:              b.Time,
	}
	return []*types.BlockHeader{&header}, nil
}

func (h *PriceOracleVoteExtHandler) getAllVolumeWeightedPrices() map[string]string {

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
		for ex, price_queue := range pairs {
			if len(price_queue) > 0 {
				p, err := math.LegacyNewDecFromStr(price_queue[0].Price)
				if err == nil {
					symbolPrices[symbol] = append(symbolPrices[symbol], p)
					providers = append(providers, ex)
					prices = append(prices, price_queue[0].Price)
				}
			}

		}
		h.logger.Info("fetch price", "symbol", symbol, "providers", providers, "price", prices)
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

	textPrices := make(map[string]string)
	for symbol, price := range avgPrices {
		textPrices[symbol] = price.String()
	}

	return textPrices
}

func (h *PriceOracleVoteExtHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		proposalTxs := req.Txs

		if req.Height >= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight && ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight != 0 {

			err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			extInfo := req.LocalLastCommit
			bz, err := extInfo.Marshal()

			// NOTE: We use stdlib JSON encoding, but an application may choose to use
			// a performant mechanism. This is for demo purposes only.
			// bz, err := json.Marshal(injectedVoteExtTx)
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

func (h *PriceOracleVoteExtHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		var injectedVoteExtTx abci.ExtendedCommitInfo
		if err := injectedVoteExtTx.Unmarshal(req.Txs[0]); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx)
		if err != nil {
			return nil, err
		}

		// Verify the proposer's stake-weighted oracle prices by computing the same
		// calculation and comparing the results. We omit verification for brevity
		// and demo purposes.
		// stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, injectedVoteExtTx)
		// if err != nil {
		// 	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		// }
		// if err := compareOraclePrices(injectedVoteExtTx.StakeWeightedPrices, stakeWeightedPrices); err != nil {
		// 	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		// }

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *PriceOracleVoteExtHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {

	res := &sdk.ResponsePreBlock{}
	if len(req.Txs) == 0 {
		return res, nil
	}

	var injectedVoteExtTx abci.ExtendedCommitInfo
	if err := injectedVoteExtTx.Unmarshal(req.Txs[0]); err != nil {
		h.logger.Error("failed to decode injected vote extension tx", "err", err)
		return nil, err
	}

	prices, headers, err := h.extractPricesAndBlockHeaders(ctx, injectedVoteExtTx)
	if err != nil {
		return nil, err
	}

	for symbol, price := range prices {
		h.Keeper.SetPrice(ctx, symbol, price.String())
	}

	for _, head := range headers {
		h.Keeper.SetBlockHeader(ctx, head)
	}

	h.logger.Warn("Oracle Final States", "price", prices)

	return res, nil
}

// func (h *PriceOracleVoteExtHandler) selectBlockHeaders(extInfo *abci.ExtendedCommitInfo) error {
// 	for i, v := range extInfo.Votes {
// 		if v.VoteExtension
// 	}
// 	return nil
// }

func (h *PriceOracleVoteExtHandler) extractPricesAndBlockHeaders(ctx sdk.Context, commit abci.ExtendedCommitInfo) (map[string]math.LegacyDec, []*types.BlockHeader, error) {
	// requiredPairs := h.keeper.GetSupportedPairs(ctx)
	// requiredPairs := []string{"BTCUSD"}
	stakeWeightedPrices := make(map[string]math.LegacyDec, len(types.PRICE_CACHE)) // base -> average stake-weighted price
	// for _, pair := range requiredPairs {
	// 	stakeWeightedPrices[pair] = math.LegacyZeroDec()
	// }
	length := len(commit.Votes)
	selectedIndex := -1

done:
	for i, v1 := range commit.Votes {
		count := 0
		if v1.VoteExtension != nil {
			for _, v2 := range commit.Votes {
				if bytes.Equal(v1.VoteExtension, v2.VoteExtension) {
					count++
					if count*3 > length*2 {
						selectedIndex = i
						break done
					}
				}
			}
		}
	}

	var totalStake int64
	var blockHeaders []*types.BlockHeader
	for i, v := range commit.Votes {
		if v.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		var voteExt types.OracleVoteExtension
		// if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
		if err := voteExt.Unmarshal(v.VoteExtension); err != nil {
			h.logger.Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
			return nil, nil, err
		}

		h.logger.Warn("extension", "validator", hex.EncodeToString(v.Validator.Address), "extension", voteExt)

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
			stakePrice, err := math.LegacyNewDecFromStr(price)
			if err != nil {
				continue
			}
			if _, ok := stakeWeightedPrices[base]; ok {
				stakeWeightedPrices[base] = stakeWeightedPrices[base].Add(stakePrice.MulInt64(v.Validator.Power))
			} else {
				stakeWeightedPrices[base] = stakePrice.MulInt64(v.Validator.Power)
			}
		}

		if i == selectedIndex {
			blockHeaders = voteExt.Blocks
		}
	}

	if totalStake == 0 {
		return nil, nil, nil
	}

	// finalize average by dividing by total stake, i.e. total weights
	for base, price := range stakeWeightedPrices {
		stakeWeightedPrices[base] = price.QuoInt64(totalStake)
	}

	return stakeWeightedPrices, blockHeaders, nil
}

// func compareOraclePrices(p1, p2 map[string]math.LegacyDec) error {
// 	if len(p1) != len(p2) {
// 		return fmt.Errorf("price maps are different, length %s != length %s", slices.Collect(maps.Keys(p1)), slices.Collect(maps.Keys(p2)))
// 	}
// 	for k, v := range p1 {
// 		if v2, ok := p2[k]; !ok || !v.Equal(v2) {
// 			return fmt.Errorf("[%s] prices are different, %s!=%s", k, v, v2)
// 		}
// 	}
// 	return nil
// }
