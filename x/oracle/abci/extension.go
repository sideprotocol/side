package abci

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/cosmos/cosmos-sdk/telemetry"

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

		if !h.config.Enable || !voteExtensionEnabled(ctx, req.Height) {
			return &abci.ResponseExtendVote{}, nil
		}
		// here we'd have a helper function that gets all the prices and does a weighted average using the volume of each market

		prices := h.getAllVolumeWeightedPrices()
		h.lastPriceSyncTS = req.Time.UnixMilli()

		headers, err := h.getBitcoinHeaders(ctx, req.Height)
		if err != nil {
			//return nil, fmt.Errorf("failed to fetch bitcoin headers: %w", err)
			h.logger.Error("failed to fetch bitcoin headers", "error", err)
		}
		voteExt := types.OracleVoteExtension{
			Height: req.Height,
			Prices: prices,
			Blocks: headers,
		}

		bz, err := voteExt.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *PriceOracleVoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {

		if !h.config.Enable || !voteExtensionEnabled(ctx, req.Height) {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
		}

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

		for _, blk := range voteExt.Blocks {
			if err = blk.Validate(); err != nil {
				return nil, types.ErrInvalidBlockHeader
			}
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func (h *PriceOracleVoteExtHandler) getBitcoinHeaders(ctx sdk.Context, sideHeight int64) ([]*types.BlockHeader, error) {
	// skip
	if sideHeight%2 == 0 {
		return nil, nil
	}

	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), "fetch blocks")

	bestHeight, err := h.bitcoinClient.GetBlockCount()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch best block header: %w", err)
	}

	confirmation := 6

	bestHeight = bestHeight - int64(confirmation) + 1
	telemetry.SetGauge(float32(bestHeight), types.ModuleName, "bitcoin", "block_height")

	hash, err := h.bitcoinClient.GetBlockHash(bestHeight)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch best block header: %w", err)
	}

	best, err := h.bitcoinClient.GetBlockHeaderVerbose(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch header: %w", err)
	}

	localBest := h.Keeper.GetBestBlockHeader(ctx)

	headers := []*types.BlockHeader{}
	// sync if block header
	if localBest == nil || localBest.Height == 0 || localBest.Hash == best.PreviousHash {

		header := types.BlockHeader{
			Version:           best.Version,
			Hash:              best.Hash,
			Height:            best.Height,
			PreviousBlockHash: best.PreviousHash,
			MerkleRoot:        best.MerkleRoot,
			Nonce:             best.Nonce,
			Bits:              best.Bits,
			Time:              best.Time,
		}
		return append(headers, &header), nil
	} else if localBest.Hash == hash.String() {
		// skip sync if synced to the latest
		return nil, nil
	}

	count := localBest.Height + 1
	for {
		if count > int32(bestHeight) || count > localBest.Height+10 {
			break
		}
		hash, err := h.bitcoinClient.GetBlockHash(int64(count))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch best block header: %w", err)
		}

		bh, err := h.bitcoinClient.GetBlockHeaderVerbose(hash)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch header: %w", err)
		}

		header := types.BlockHeader{
			Version:           bh.Version,
			Hash:              bh.Hash,
			Height:            bh.Height,
			PreviousBlockHash: bh.PreviousHash,
			MerkleRoot:        bh.MerkleRoot,
			Nonce:             bh.Nonce,
			Bits:              bh.Bits,
			Time:              bh.Time,
		}
		headers = append(headers, &header)
		count++
	}

	return headers, nil
}

func (h *PriceOracleVoteExtHandler) getAllVolumeWeightedPrices() map[string]string {

	avgPrices := make(map[string]math.LegacyDec)
	symbolPrices := types.GetPrices(h.lastPriceSyncTS)
	for symbol, prices := range symbolPrices {
		if len(prices) > 0 {
			sum := math.LegacyZeroDec()
			for _, p := range prices {
				sum = sum.Add(p)
			}
			if avg := sum.QuoInt64(int64(len(prices))); avg.GT(math.LegacyZeroDec()) {
				avgPrices[symbol] = avg
			}
		}
	}

	h.logger.Info("AvgPrice", "prices", avgPrices)

	textPrices := make(map[string]string)
	for symbol, price := range avgPrices {
		if price.GT(math.LegacyZeroDec()) {
			textPrices[symbol] = price.String()
		} else {
			h.logger.Error("Invalid Source Price", "symbol", symbol, "price", price)
		}

	}

	return textPrices
}

func (h *PriceOracleVoteExtHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		proposalTxs := req.Txs

		if voteExtensionEnabled(ctx, req.Height) {

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
		if !voteExtensionEnabled(ctx, req.Height) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

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

	if !voteExtensionEnabled(ctx, req.Height) {
		return res, nil
	}

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
		if price.GT(math.LegacyZeroDec()) {
			h.Keeper.SetPrice(ctx, symbol, price.String())
		}
	}

	err = h.Keeper.SetBlockHeaders(ctx, headers)
	if err != nil {
		return nil, err
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

func (h *PriceOracleVoteExtHandler) extractPricesAndBlockHeaders(_ sdk.Context, commit abci.ExtendedCommitInfo) (map[string]math.LegacyDec, []*types.BlockHeader, error) {
	var totalStake int64

	stakeWeightedPrices := make(map[string]math.LegacyDec, len(types.PRICE_CACHE)) // base -> average stake-weighted price
	stakeWeightedVotingPower := make(map[string]int64, len(types.PRICE_CACHE))
	blockHeaders := make(map[string][]*types.BlockHeader)
	headerStakes := make(map[string]int64)

	for _, v := range commit.Votes {
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

		// totalStake += v.Validator.Power

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
			if stakePrice.LTE(math.LegacyZeroDec()) {
				continue
			}
			if _, ok := stakeWeightedPrices[base]; ok {
				stakeWeightedPrices[base] = stakeWeightedPrices[base].Add(stakePrice.MulInt64(v.Validator.Power))
				stakeWeightedVotingPower[base] += v.Validator.Power
			} else {
				stakeWeightedPrices[base] = stakePrice.MulInt64(v.Validator.Power)
				stakeWeightedVotingPower[base] += v.Validator.Power
			}
		}

		sha := sha256.New()
		for _, block := range voteExt.Blocks {
			sha.Write([]byte(block.Hash))
		}
		key := fmt.Sprintf("%x", sha.Sum(nil))

		blockHeaders[key] = voteExt.Blocks
		if power, ok := headerStakes[key]; ok {
			headerStakes[key] = power + v.Validator.Power
		} else {
			headerStakes[key] = v.Validator.Power
		}
	}

	if totalStake == 0 {
		return nil, nil, nil
	}

	// finalize average by dividing by total stake, i.e. total weights
	for base, price := range stakeWeightedPrices {
		if price.GT(math.LegacyZeroDec()) {
			if vp, ok := stakeWeightedVotingPower[base]; ok {
				stakeWeightedPrices[base] = price.QuoInt64(vp)
			}
		} else {
			h.logger.Error("Got invalid price.", "symbal", base, "price", price)
		}
	}

	headers := []*types.BlockHeader{}
	for key, power := range headerStakes {
		if selected, ok := blockHeaders[key]; ok && power*3 > totalStake*2 {
			headers = append(headers, selected...)
			break
		}
	}
	return stakeWeightedPrices, headers, nil
}

func voteExtensionEnabled(ctx sdk.Context, height int64) bool {
	consParams := ctx.ConsensusParams()

	return consParams.Abci != nil && height > consParams.Abci.VoteExtensionsEnableHeight && consParams.Abci.VoteExtensionsEnableHeight != 0
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
