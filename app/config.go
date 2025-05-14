package app

import (
	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/bitcoin/keys"
)

func init() {
	// Set prefixes
	accountPubKeyPrefix := AccountAddressPrefix + "pub"
	validatorAddressPrefix := AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	// config.SetBtcChainCfg(&chaincfg.TestNet3Params)
	config.Seal()

	bech32.BITCOIN_HRP = keys.Network.Bech32HRPSegwit
	bech32.SIDE_HRP = AccountAddressPrefix
}
