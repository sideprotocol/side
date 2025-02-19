package network

// import (
// 	"github.com/cosmos/cosmos-sdk/testutil/network"
// )

// type (
// 	Network = network.Network
// 	Config  = network.Config
// )

// // New creates instance with fully configured cosmos network.
// // Accepts optional config, that will be used in place of the DefaultConfig() if provided.
// // func New(t *testing.T, configs ...Config) *Network {
// // 	if len(configs) > 1 {
// // 		panic("at most one config should be provided")
// // 	}
// // 	var cfg network.Config
// // 	if len(configs) == 0 {
// // 		cfg = DefaultConfig()
// // 	} else {
// // 		cfg = configs[0]
// // 	}
// // 	net, err := network.New(t, t.TempDir(), cfg)
// // 	require.NoError(t, err)
// // 	_, err = net.WaitForHeight(1)
// // 	require.NoError(t, err)
// // 	t.Cleanup(net.Cleanup)
// // 	return net
// // }

// // DefaultConfig will initialize config for the network with custom application,
// // genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
// // func DefaultConfig() network.Config {
// // 	var (
// // 		encoding = app.MakeEncodingConfig()
// // 		chainID  = "chain-" + tmrand.NewRand().Str(6)
// // 	)
// // 	return network.Config{
// // 		Codec:             encoding.Codec,
// // 		TxConfig:          encoding.TxConfig,
// // 		LegacyAmino:       encoding.Amino,
// // 		InterfaceRegistry: encoding.InterfaceRegistry,
// // 		AccountRetriever:  authtypes.AccountRetriever{},
// // 		AppConstructor: func(val network.ValidatorI) servertypes.Application {
// // 			return app.NewSideApp(
// // 				val.GetCtx().Logger,
// // 				dbm.NewMemDB(),
// // 				nil,
// // 				true,
// // 				simtestutil.EmptyAppOptions{},
// // 				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
// // 				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
// // 				baseapp.SetChainID(chainID),
// // 			)
// // 		},
// // 		GenesisState:    app.ModuleBasics.DefaultGenesis(encoding.Codec),
// // 		TimeoutCommit:   2 * time.Second,
// // 		ChainID:         chainID,
// // 		NumValidators:   1,
// // 		BondDenom:       sdk.DefaultBondDenom,
// // 		MinGasPrices:    fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
// // 		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
// // 		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
// // 		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
// // 		PruningStrategy: pruningtypes.PruningOptionNothing,
// // 		CleanupDir:      true,
// // 		SigningAlgo:     string(hd.Secp256k1Type),
// // 		KeyringOptions:  []keyring.Option{},
// // 	}
// // }
