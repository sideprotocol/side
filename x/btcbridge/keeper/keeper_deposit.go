package keeper

import (
	"github.com/btcsuite/btcd/btcutil"
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

	recipient, err := k.Mint(ctx, tx, prevTx, k.GetBlockHeader(ctx, msg.Blockhash).Height)
	if err != nil {
		return nil, nil, err
	}

	return tx.Hash(), recipient, nil
}

// Mint performs the minting operation of the voucher token
func (k Keeper) Mint(ctx sdk.Context, tx *btcutil.Tx, prevTx *btcutil.Tx, height uint64) (btcutil.Address, error) {
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

	// extract the recipient for minting voucher token
	recipient, err := types.ExtractRecipientAddr(tx.MsgTx(), prevTx.MsgTx(), params.Vaults, isRunes, chainCfg)
	if err != nil {
		return nil, err
	}

	// mint voucher token if the receiver is a vault address
	for i, out := range tx.MsgTx().TxOut {
		if types.IsOpReturnOutput(out) {
			continue
		}

		// check if the output is a valid address
		pks, err := txscript.ParsePkScript(out.PkScript)
		if err != nil {
			return nil, err
		}
		addr, err := pks.Address(chainCfg)
		if err != nil {
			return nil, err
		}

		// check if the receiver is one of the vault addresses
		vault := types.SelectVaultByAddress(params.Vaults, addr.EncodeAddress())
		if vault == nil {
			continue
		}

		// mint the voucher token by asset type
		// skip if the asset type of the sender address is unspecified
		switch vault.AssetType {
		case types.AssetType_ASSET_TYPE_BTC:
			err := k.mintBTC(ctx, tx, height, recipient.EncodeAddress(), vault, out, i, params.BtcVoucherDenom)
			if err != nil {
				return nil, err
			}

		case types.AssetType_ASSET_TYPE_RUNES:
			if isRunes && edict.Output == uint32(i) {
				if err := k.mintRunes(ctx, tx, height, recipient.EncodeAddress(), vault, out, i, edict.Id, edict.Amount); err != nil {
					return nil, err
				}
			}
		}
	}

	return recipient, nil
}

func (k Keeper) mintBTC(ctx sdk.Context, tx *btcutil.Tx, height uint64, sender string, vault *types.Vault, out *wire.TxOut, vout int, denom string) error {
	// save the hash of the transaction to prevent double minting
	hash := tx.Hash().String()
	if k.existsInHistory(ctx, hash) {
		return types.ErrTransactionAlreadyMinted
	}
	k.addToMintHistory(ctx, hash)

	params := k.GetParams(ctx)

	protocolFee := sdk.NewInt64Coin(denom, params.ProtocolFees.DepositFee)
	protocolFeeCollector := sdk.MustAccAddressFromBech32(params.ProtocolFees.Collector)

	amount := sdk.NewInt64Coin(denom, out.Value)

	depositAmount := amount.Sub(protocolFee)
	if depositAmount.Amount.Int64() < params.ProtocolLimits.BtcMinDeposit {
		return types.ErrInvalidDepositAmount
	}

	receipient, err := sdk.AccAddressFromBech32(sender)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receipient, sdk.NewCoins(depositAmount)); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, protocolFeeCollector, sdk.NewCoins(protocolFee)); err != nil {
		return err
	}

	utxo := types.UTXO{
		Txid:         hash,
		Vout:         uint64(vout),
		Amount:       uint64(out.Value),
		PubKeyScript: out.PkScript,
		Height:       height,
		Address:      vault.Address,
		IsLocked:     false,
	}

	k.saveUTXO(ctx, &utxo)

	return nil
}

func (k Keeper) mintRunes(ctx sdk.Context, tx *btcutil.Tx, height uint64, recipient string, vault *types.Vault, out *wire.TxOut, vout int, id *types.RuneId, amount string) error {
	// save the hash of the transaction to prevent double minting
	hash := tx.Hash().String()
	if k.existsInHistory(ctx, hash) {
		return types.ErrTransactionAlreadyMinted
	}
	k.addToMintHistory(ctx, hash)

	coins := sdk.NewCoins(sdk.NewCoin(id.Denom(), sdk.NewIntFromBigInt(types.RuneAmountFromString(amount).Big())))

	receipientAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receipientAddr, coins); err != nil {
		return err
	}

	utxo := types.UTXO{
		Txid:         hash,
		Vout:         uint64(vout),
		Amount:       uint64(out.Value),
		PubKeyScript: out.PkScript,
		Height:       height,
		Address:      vault.Address,
		IsLocked:     false,
		Runes: []*types.RuneBalance{{
			Id:     id.ToString(),
			Amount: amount,
		}},
	}

	k.saveUTXO(ctx, &utxo)

	return nil
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
