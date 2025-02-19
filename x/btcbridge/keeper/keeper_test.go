package keeper_test

// import (
// 	"bytes"
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/suite"
// 	"lukechampine.com/uint128"

// 	"github.com/btcsuite/btcd/btcutil"
// 	"github.com/btcsuite/btcd/btcutil/psbt"
// 	"github.com/btcsuite/btcd/chaincfg/chainhash"
// 	"github.com/btcsuite/btcd/wire"

// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// 	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

// 	simapp "github.com/sideprotocol/side/app"
// 	"github.com/sideprotocol/side/x/btcbridge/types"
// )

// var (
// 	InitCoinAmount = int64(1000000000000)
// )

// type KeeperTestSuite struct {
// 	suite.Suite

// 	ctx sdk.Context
// 	app *simapp.SideApp

// 	btcVault   string
// 	runesVault string
// 	sender     string

// 	btcVaultPkScript   []byte
// 	runesVaultPkScript []byte
// 	senderPkScript     []byte
// }

// // func (suite *KeeperTestSuite) SetupTest() {
// // 	app := simapp.Setup(suite.T())
// // 	ctx := app.BaseApp.NewContext(false)

// // 	suite.ctx = ctx
// // 	suite.app = app

// // 	chainCfg := sdk.GetConfig().GetBtcChainCfg()

// // 	suite.btcVault, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address().Bytes())
// // 	suite.runesVault, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address())
// // 	suite.sender, _ = bech32.Encode(chainCfg.Bech32HRPSegwit, segwit.GenPrivKey().PubKey().Address())

// // 	suite.btcVaultPkScript = types.MustPkScriptFromAddress(suite.btcVault)
// // 	suite.runesVaultPkScript = types.MustPkScriptFromAddress(suite.runesVault)
// // 	suite.senderPkScript = types.MustPkScriptFromAddress(suite.sender)

// // 	suite.setupParams(suite.btcVault, suite.runesVault, suite.sender)
// // 	suite.mintAssets(suite.sender)
// // }

// func TestKeeperSuite(t *testing.T) {
// 	suite.Run(t, new(KeeperTestSuite))
// }

// func (suite *KeeperTestSuite) setupParams(btcVault string, runesVault string, nonBtcRelayer string) {
// 	params := suite.app.BtcBridgeKeeper.GetParams(suite.ctx)

// 	params.TrustedNonBtcRelayers = []string{nonBtcRelayer}
// 	params.Vaults = []*types.Vault{
// 		{
// 			Address:   btcVault,
// 			AssetType: types.AssetType_ASSET_TYPE_BTC,
// 		},
// 		{
// 			Address:   runesVault,
// 			AssetType: types.AssetType_ASSET_TYPE_RUNES,
// 		},
// 	}
// 	params.ProtocolFees.Collector = authtypes.NewModuleAddress(govtypes.ModuleName).String()

// 	suite.app.BtcBridgeKeeper.SetParams(suite.ctx, params)
// }

// func (suite *KeeperTestSuite) mintAssets(addresses ...string) {
// 	for _, addr := range addresses {
// 		coins := sdk.NewCoins(sdk.NewInt64Coin(types.DefaultBtcVoucherDenom, InitCoinAmount))

// 		suite.app.BankKeeper.MintCoins(suite.ctx, types.ModuleName, coins)
// 		suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.ModuleName, sdk.MustAccAddressFromBech32(addr), coins)
// 	}
// }

// func (suite *KeeperTestSuite) setupUTXOs(utxos []*types.UTXO) {
// 	for _, utxo := range utxos {
// 		suite.app.BtcBridgeKeeper.SaveUTXO(suite.ctx, utxo)
// 	}
// }

// func (suite *KeeperTestSuite) TestMintRunes() {
// 	params := suite.app.BtcBridgeKeeper.GetParams(suite.ctx)

// 	runeId := "840000:3"
// 	runeAmount := 500000000
// 	runeOutputIndex := 2

// 	runesScript, err := types.BuildEdictScript(runeId, uint128.From64(uint64(runeAmount)), uint32(runeOutputIndex))
// 	suite.NoError(err)

// 	tx := wire.NewMsgTx(types.TxVersion)
// 	tx.AddTxOut(wire.NewTxOut(0, runesScript))
// 	tx.AddTxOut(wire.NewTxOut(types.RunesOutValue, suite.senderPkScript))
// 	tx.AddTxOut(wire.NewTxOut(types.RunesOutValue, suite.runesVaultPkScript))
// 	tx.AddTxOut(wire.NewTxOut(params.ProtocolFees.DepositFee, suite.btcVaultPkScript))

// 	denom := fmt.Sprintf("%s/%s", types.RunesProtocolName, runeId)

// 	balanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(suite.sender), denom)
// 	suite.True(balanceBefore.Amount.IsZero(), "%s balance before mint should be zero", denom)

// 	recipient, err := suite.app.BtcBridgeKeeper.Mint(suite.ctx, suite.sender, btcutil.NewTx(tx), btcutil.NewTx(tx), 0)
// 	suite.NoError(err)
// 	suite.Equal(suite.sender, recipient.EncodeAddress(), "incorrect recipient")

// 	balanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, sdk.MustAccAddressFromBech32(suite.sender), denom)
// 	suite.Equal(uint64(runeAmount), balanceAfter.Amount.Uint64(), "%s balance after mint should be %d", denom, runeAmount)

// 	utxos := suite.app.BtcBridgeKeeper.GetAllUTXOs(suite.ctx)
// 	suite.Len(utxos, 2, "there should be 1 utxo(s)")

// 	expectedRunesUTXO := &types.UTXO{
// 		Txid:         tx.TxHash().String(),
// 		Vout:         uint64(runeOutputIndex),
// 		Address:      suite.runesVault,
// 		Amount:       types.RunesOutValue,
// 		PubKeyScript: suite.runesVaultPkScript,
// 		IsLocked:     false,
// 		Runes: []*types.RuneBalance{
// 			{
// 				Id:     runeId,
// 				Amount: fmt.Sprintf("%d", runeAmount),
// 			},
// 		},
// 	}

// 	expectedBtcUTXO := &types.UTXO{
// 		Txid:         tx.TxHash().String(),
// 		Vout:         3,
// 		Address:      suite.btcVault,
// 		Amount:       uint64(params.ProtocolFees.DepositFee),
// 		PubKeyScript: suite.btcVaultPkScript,
// 		IsLocked:     false,
// 	}

// 	suite.Equal(expectedRunesUTXO, utxos[0], "runes utxo does not match")
// 	suite.Equal(expectedBtcUTXO, utxos[1], "btc utxo does not match")
// }

// func (suite *KeeperTestSuite) TestWithdrawRunes() {
// 	runeId := "840000:3"
// 	runeAmount := 500000000

// 	runesUTXOs := []*types.UTXO{
// 		{
// 			Txid:         chainhash.HashH([]byte("runes")).String(),
// 			Vout:         1,
// 			Address:      suite.runesVault,
// 			Amount:       types.RunesOutValue,
// 			PubKeyScript: suite.runesVaultPkScript,
// 			IsLocked:     false,
// 			Runes: []*types.RuneBalance{
// 				{
// 					Id:     runeId,
// 					Amount: fmt.Sprintf("%d", runeAmount),
// 				},
// 			},
// 		},
// 	}
// 	suite.setupUTXOs(runesUTXOs)

// 	feeRate := 100
// 	amount := runeAmount + 1

// 	denom := fmt.Sprintf("%s/%s", types.RunesProtocolName, runeId)
// 	coin := sdk.NewInt64Coin(denom, int64(amount))

// 	_, err := suite.app.BtcBridgeKeeper.NewRunesSigningRequest(suite.ctx, suite.sender, coin, int64(feeRate), suite.runesVault, suite.btcVault)
// 	suite.ErrorIs(err, types.ErrInsufficientUTXOs, "should fail due to insufficient runes utxos")

// 	amount = 100000000
// 	coin = sdk.NewInt64Coin(denom, int64(amount))

// 	_, err = suite.app.BtcBridgeKeeper.NewRunesSigningRequest(suite.ctx, suite.sender, coin, int64(feeRate), suite.runesVault, suite.btcVault)
// 	suite.ErrorIs(err, types.ErrInsufficientUTXOs, "should fail due to insufficient payment utxos")

// 	paymentUTXOs := []*types.UTXO{
// 		{
// 			Txid:         chainhash.HashH([]byte("payment")).String(),
// 			Vout:         1,
// 			Address:      suite.btcVault,
// 			Amount:       100000,
// 			PubKeyScript: suite.btcVaultPkScript,
// 			IsLocked:     false,
// 		},
// 	}
// 	suite.setupUTXOs(paymentUTXOs)

// 	req, err := suite.app.BtcBridgeKeeper.NewRunesSigningRequest(suite.ctx, suite.sender, coin, int64(feeRate), suite.runesVault, suite.btcVault)
// 	suite.NoError(err)

// 	suite.False(suite.app.BtcBridgeKeeper.HasUTXO(suite.ctx, runesUTXOs[0].Txid, runesUTXOs[0].Vout), "runes utxo should be spent")
// 	suite.False(suite.app.BtcBridgeKeeper.HasUTXO(suite.ctx, paymentUTXOs[0].Txid, paymentUTXOs[0].Vout), "payment utxo should be spent")

// 	runesUTXOs = suite.app.BtcBridgeKeeper.GetUTXOsByAddr(suite.ctx, suite.runesVault)
// 	suite.Len(runesUTXOs, 1, "there should be 1 runes utxo(s)")

// 	suite.True(runesUTXOs[0].IsLocked, "the rune utxo should be locked")
// 	suite.Len(runesUTXOs[0].Runes, 1, "there should be 1 rune in the runes utxo")
// 	suite.Equal(runeId, runesUTXOs[0].Runes[0].Id, "incorrect rune id")
// 	suite.Equal(uint64(runeAmount-amount), types.RuneAmountFromString(runesUTXOs[0].Runes[0].Amount).Big().Uint64(), "incorrect rune amount")

// 	p, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(req.Psbt)), true)
// 	suite.NoError(err)

// 	suite.Len(p.Inputs, 2, "there should be 2 inputs")
// 	suite.Equal(suite.runesVaultPkScript, p.Inputs[0].WitnessUtxo.PkScript, "the first input should be runes vault")
// 	suite.Equal(suite.btcVaultPkScript, p.Inputs[1].WitnessUtxo.PkScript, "the second input should be btc vault")

// 	expectedRunesScript, err := types.BuildEdictScript(runeId, uint128.From64(uint64(amount)), 2)
// 	suite.NoError(err)

// 	suite.Len(p.UnsignedTx.TxOut, 4, "there should be 4 outputs")
// 	suite.Equal(expectedRunesScript, p.UnsignedTx.TxOut[0].PkScript, "incorrect runes script")
// 	suite.Equal(suite.runesVaultPkScript, p.UnsignedTx.TxOut[1].PkScript, "the second output should be runes change output")
// 	suite.Equal(suite.senderPkScript, p.UnsignedTx.TxOut[2].PkScript, "the third output should be sender output")
// 	suite.Equal(suite.btcVaultPkScript, p.UnsignedTx.TxOut[3].PkScript, "the fouth output should be btc change output")
// }
