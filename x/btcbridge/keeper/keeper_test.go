package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"lukechampine.com/uint128"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/crypto/keys/segwit"
	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/sideprotocol/side/app"
	"github.com/sideprotocol/side/x/btcbridge/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.App

	btcVault   string
	runesVault string
	sender     string

	btcVaultPkScript   []byte
	runesVaultPkScript []byte
	senderPkScript     []byte
}

func (suite *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(suite.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now().UTC()})

	suite.ctx = ctx
	suite.app = app

	chainCfg := sdk.GetConfig().GetBtcChainCfg()

	suite.btcVault, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address().Bytes())
	suite.runesVault, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address())
	suite.sender, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address())

	suite.btcVaultPkScript = MustPkScriptFromAddress(suite.btcVault, chainCfg)
	suite.runesVaultPkScript = MustPkScriptFromAddress(suite.runesVault, chainCfg)
	suite.senderPkScript = MustPkScriptFromAddress(suite.sender, chainCfg)

	suite.setupParams(suite.btcVault, suite.runesVault)
}

func TestKeeperSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) setupParams(btcVault string, runesVault string) {
	suite.app.BtcBridgeKeeper.SetParams(suite.ctx, types.Params{Vaults: []*types.Vault{
		{
			Address:   btcVault,
			AssetType: types.AssetType_ASSET_TYPE_BTC,
		},
		{
			Address:   runesVault,
			AssetType: types.AssetType_ASSET_TYPE_RUNE,
		},
	}})
}

func (suite *KeeperTestSuite) TestMintRunes() {
	runeId := "840000:3"
	runeAmount := 500000000
	runeOutputIndex := 2

	runesScript, err := types.BuildEdictScript(runeId, uint128.From64(uint64(runeAmount)), uint32(runeOutputIndex))
	suite.NoError(err)

	tx := wire.NewMsgTx(types.TxVersion)
	tx.AddTxOut(wire.NewTxOut(0, runesScript))
	tx.AddTxOut(wire.NewTxOut(types.RunesOutValue, suite.senderPkScript))
	tx.AddTxOut(wire.NewTxOut(types.RunesOutValue, suite.runesVaultPkScript))

	denom := fmt.Sprintf("%s/%s", types.RunesProtocolName, runeId)

	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(suite.sender), denom)
	suite.True(balanceBefore.Amount.IsZero(), "%s balance before mint should be zero", denom)

	recipient, err := suite.app.BtcBridgeKeeper.Mint(suite.ctx, btcutil.NewTx(tx), btcutil.NewTx(tx), 0)
	suite.NoError(err)
	suite.Equal(suite.sender, recipient.EncodeAddress(), "incorrect recipient")

	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(suite.sender), denom)
	suite.Equal(uint64(runeAmount), balanceAfter.Amount.Uint64(), "%s balance after mint should be %d", denom, runeAmount)
}

func (suite *KeeperTestSuite) TestWithdrawRunes() {
	// runeId := "840000:3"
	// runeAmount := 500000000
}

func MustPkScriptFromAddress(addr string, chainCfg *chaincfg.Params) []byte {
	address, err := btcutil.DecodeAddress(addr, chainCfg)
	if err != nil {
		panic(err)
	}

	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		panic(err)
	}

	return pkScript
}
