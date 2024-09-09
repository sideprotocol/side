package keeper

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/btcbridge/types"
)

// ProcessBitcoinDepositTransaction handles the deposit transaction
func (k Keeper) ProcessBitcoinDepositTransaction(ctx sdk.Context, msg *types.MsgSubmitDepositTransaction) (*chainhash.Hash, btcutil.Address, error) {
	ctx.Logger().Info("accept bitcoin deposit tx", "blockhash", msg.Blockhash)

	tx, prevTx, err := k.ValidateTransaction(ctx, msg.TxBytes, msg.PrevTxBytes, msg.Blockhash, msg.Proof)
	if err != nil {
		return nil, nil, err
	}

	recipient, err := k.Mint(ctx, msg.Sender, tx, prevTx, k.GetBlockHeader(ctx, msg.Blockhash).Height)
	if err != nil {
		return nil, nil, err
	}

	return tx.Hash(), recipient, nil
}

// Mint performs the minting operation of the voucher token
func (k Keeper) Mint(ctx sdk.Context, sender string, tx *btcutil.Tx, prevTx *btcutil.Tx, height uint64) (btcutil.Address, error) {
	hash := tx.Hash().String()
	if k.existsInHistory(ctx, hash) {
		return nil, types.ErrTransactionAlreadyMinted
	}

	k.addToMintHistory(ctx, hash)

	params := k.GetParams(ctx)
	chainCfg := sdk.GetConfig().GetBtcChainCfg()

	// check if this is a valid runes deposit tx
	// if any error encountered, this tx is illegal runes deposit
	// if the edict is not nil, it indicates that this is a legal runes deposit tx
	edict, err := types.CheckRunesDepositTransaction(tx.MsgTx(), params.Vaults)
	if err != nil {
		return nil, err
	}

	isRunes := edict != nil

	// check if the sender is trusted to relay runes deposit
	if isRunes && !k.IsTrustedNonBtcRelayer(ctx, sender) {
		return nil, types.ErrUntrustedNonBtcRelayer
	}

	// extract the recipient for minting voucher token
	recipient, err := types.ExtractRecipientAddr(tx.MsgTx(), prevTx.MsgTx(), params.Vaults, isRunes, chainCfg)
	if err != nil {
		return nil, err
	}

	if !isRunes {
		out, vout, vault, err := k.getOutputForMintBTC(ctx, tx.MsgTx(), chainCfg)
		if err != nil {
			return nil, err
		}

		if err := k.mintBTC(ctx, tx, height, recipient.EncodeAddress(), vault, out, vout, params.BtcVoucherDenom); err != nil {
			return nil, err
		}
	} else {
		outs, vouts, vaults, err := k.getOutputsForMintRunes(ctx, tx.MsgTx(), edict, chainCfg)
		if err != nil {
			return nil, err
		}

		if err := k.mintRunes(ctx, tx, height, recipient.EncodeAddress(), vaults, outs, vouts, edict.Id, edict.Amount); err != nil {
			return nil, err
		}
	}

	return recipient, nil
}

func (k Keeper) mintBTC(ctx sdk.Context, tx *btcutil.Tx, height uint64, recipient string, vault string, out *wire.TxOut, vout int, denom string) error {
	amount := sdk.NewInt64Coin(denom, out.Value)

	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	if err := k.mintBTCWithProtocolFee(ctx, recipientAddr, amount); err != nil {
		return err
	}

	utxo := types.UTXO{
		Txid:         tx.Hash().String(),
		Vout:         uint64(vout),
		Amount:       uint64(out.Value),
		PubKeyScript: out.PkScript,
		Height:       height,
		Address:      vault,
		IsLocked:     false,
	}

	k.saveUTXO(ctx, &utxo)

	return nil
}

func (k Keeper) mintRunes(ctx sdk.Context, tx *btcutil.Tx, height uint64, recipient string, vaults []string, outs []*wire.TxOut, vouts []int, id *types.RuneId, amount string) error {
	coins := sdk.NewCoins(sdk.NewCoin(id.Denom(), sdk.NewIntFromBigInt(types.RuneAmountFromString(amount).Big())))

	recipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipientAddr, coins); err != nil {
		return err
	}

	if k.ProtocolDepositFeeEnabled(ctx) {
		if err := k.handleRunesProtocolFee(ctx, tx.Hash().String(), height, outs[1], vouts[1], vaults[1]); err != nil {
			return err
		}
	}

	utxo := types.UTXO{
		Txid:         tx.Hash().String(),
		Vout:         uint64(vouts[0]),
		Amount:       uint64(outs[0].Value),
		PubKeyScript: outs[0].PkScript,
		Height:       height,
		Address:      vaults[0],
		IsLocked:     false,
		Runes: []*types.RuneBalance{{
			Id:     id.ToString(),
			Amount: amount,
		}},
	}

	k.saveUTXO(ctx, &utxo)

	return nil
}

// mintBTCWithProtocolFee performs btc minting along with the protocol fee handling
func (k Keeper) mintBTCWithProtocolFee(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coin) error {
	params := k.GetParams(ctx)

	var err error
	depositAmount := amount

	if k.ProtocolDepositFeeEnabled(ctx) {
		protocolFee := sdk.NewInt64Coin(params.BtcVoucherDenom, params.ProtocolFees.DepositFee)
		protocolFeeCollector := sdk.MustAccAddressFromBech32(params.ProtocolFees.Collector)

		depositAmount, err = depositAmount.SafeSub(protocolFee)
		if err != nil {
			return err
		}

		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, protocolFeeCollector, sdk.NewCoins(protocolFee)); err != nil {
			return err
		}
	}

	if depositAmount.Amount.Int64() < params.ProtocolLimits.BtcMinDeposit {
		return types.ErrInvalidDepositAmount
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(depositAmount))
}

// handleRunesProtocolFee performs the protocol fee handling for runes deposit
// Assume that the protocol deposit fee is enabled
func (k Keeper) handleRunesProtocolFee(ctx sdk.Context, txHash string, height uint64, btcOut *wire.TxOut, btcVout int, btcVault string) error {
	params := k.GetParams(ctx)

	btcAmount := sdk.NewInt64Coin(params.BtcVoucherDenom, btcOut.Value)

	protocolFee := sdk.NewInt64Coin(params.BtcVoucherDenom, params.ProtocolFees.DepositFee)
	protocolFeeCollector := sdk.MustAccAddressFromBech32(params.ProtocolFees.Collector)

	if btcAmount.IsLT(protocolFee) {
		return types.ErrInvalidDepositAmount
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(btcAmount)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, protocolFeeCollector, sdk.NewCoins(protocolFee)); err != nil {
		return err
	}

	utxo := types.UTXO{
		Txid:         txHash,
		Vout:         uint64(btcVout),
		Amount:       uint64(btcOut.Value),
		PubKeyScript: btcOut.PkScript,
		Height:       height,
		Address:      btcVault,
		IsLocked:     false,
	}

	k.saveUTXO(ctx, &utxo)

	return nil
}

func (k Keeper) getOutputForMintBTC(ctx sdk.Context, tx *wire.MsgTx, chainCfg *chaincfg.Params) (*wire.TxOut, int, string, error) {
	params := k.GetParams(ctx)

	for i, out := range tx.TxOut {
		if types.IsOpReturnOutput(out) {
			continue
		}

		// check if the output is a valid address
		pks, err := txscript.ParsePkScript(out.PkScript)
		if err != nil {
			return nil, 0, "", err
		}

		addr, err := pks.Address(chainCfg)
		if err != nil {
			return nil, 0, "", err
		}

		// check if the address is one of the vault addresses
		vault := types.SelectVaultByAddress(params.Vaults, addr.EncodeAddress())
		if vault == nil {
			continue
		}

		if vault.AssetType == types.AssetType_ASSET_TYPE_BTC {
			return out, i, vault.Address, nil
		}
	}

	return nil, 0, "", types.ErrInvalidDepositTransaction
}

func (k Keeper) getOutputsForMintRunes(ctx sdk.Context, tx *wire.MsgTx, edict *types.Edict, chainCfg *chaincfg.Params) ([]*wire.TxOut, []int, []string, error) {
	params := k.GetParams(ctx)

	outs := make([]*wire.TxOut, 2)
	vouts := make([]int, 2)
	vaults := make([]string, 2)

	for i, out := range tx.TxOut {
		if types.IsOpReturnOutput(out) {
			continue
		}

		// check if the output is a valid address
		pks, err := txscript.ParsePkScript(out.PkScript)
		if err != nil {
			return nil, nil, nil, err
		}

		addr, err := pks.Address(chainCfg)
		if err != nil {
			return nil, nil, nil, err
		}

		// check if the address is one of the vault addresses
		vault := types.SelectVaultByAddress(params.Vaults, addr.EncodeAddress())
		if vault == nil {
			continue
		}

		switch vault.AssetType {
		case types.AssetType_ASSET_TYPE_RUNES:
			outs[0] = out
			vouts[0] = i
			vaults[0] = vault.Address

		case types.AssetType_ASSET_TYPE_BTC:
			outs[1] = out
			vouts[1] = i
			vaults[1] = vault.Address
		}
	}

	if outs[0] == nil || vouts[0] != int(edict.Output) {
		return nil, nil, nil, types.ErrInvalidDepositTransaction
	}

	if k.ProtocolDepositFeeEnabled(ctx) && outs[1] == nil {
		return nil, nil, nil, types.ErrInvalidDepositTransaction
	}

	return outs, vouts, vaults, nil
}

func (k Keeper) existsInHistory(ctx sdk.Context, txHash string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.BtcMintedTxHashKey(txHash))
}

func (k Keeper) addToMintHistory(ctx sdk.Context, txHash string) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.BtcMintedTxHashKey(txHash), []byte{1})
}

// need a query all history for exporting
