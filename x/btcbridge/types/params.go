package types

import (
	"bytes"
	"encoding/hex"
	"time"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// default reward epoch
	DefaultRewardEpoch = time.Duration(1209600) * time.Second // 14 days

	// default TSS update transition period
	DefaultTssUpdateTransitionPeriod = time.Duration(1209600) * time.Second // 14 days
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		Confirmations:           2,
		MaxAcceptableBlockDepth: 100,
		BtcVoucherDenom:         "sat",
		Vaults: []*Vault{{
			Address:   "",
			PubKey:    "",
			AssetType: AssetType_ASSET_TYPE_BTC,
		}, {
			Address:   "",
			PubKey:    "",
			AssetType: AssetType_ASSET_TYPE_RUNE,
		}},
		ProtocolLimits: &ProtocolLimits{
			BtcMinDeposit:  50000,     // 0.0005 BTC
			BtcMinWithdraw: 30000,     // 0.0003 BTC
			BtcMaxWithdraw: 500000000, // 5 BTC
		},
		ProtocolFees: &ProtocolFees{
			DepositFee:  8000,  // 0.00008 BTC
			WithdrawFee: 12000, // 0.00012 BTC
			Collector:   "",
		},
		NetworkFee:                8000, // 0.00008 BTC
		RewardEpoch:               &DefaultRewardEpoch,
		TssUpdateTransitionPeriod: &DefaultTssUpdateTransitionPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := sdk.ValidateDenom(p.BtcVoucherDenom); err != nil {
		return err
	}

	vaults := make(map[string]bool)

	for _, vault := range p.Vaults {
		if len(vault.Address) != 0 {
			_, err := sdk.AccAddressFromBech32(vault.Address)
			if err != nil {
				return err
			}

			_, ok := vaults[vault.Address]
			if ok {
				return ErrInvalidParams
			}
		}

		if len(vault.PubKey) != 0 {
			pkBytes, err := hex.DecodeString(vault.PubKey)
			if err != nil {
				return err
			}

			_, err = secp256k1.ParsePubKey(pkBytes)
			if err != nil {
				return err
			}
		}

		if vault.AssetType == AssetType_ASSET_TYPE_UNSPECIFIED {
			return ErrInvalidParams
		}
	}

	if p.ProtocolLimits != nil {
		if p.ProtocolLimits.BtcMinWithdraw > p.ProtocolLimits.BtcMaxWithdraw {
			return ErrInvalidParams
		}
	}

	if p.ProtocolFees != nil {
		if len(p.ProtocolFees.Collector) != 0 {
			_, err := sdk.AccAddressFromBech32(p.ProtocolFees.Collector)
			if err != nil {
				return ErrInvalidParams
			}
		}
	}

	return nil
}

// SelectVaultByBitcoinAddress returns the vault if the address is found
// returns the vault if the address is found
func SelectVaultByBitcoinAddress(vaults []*Vault, address string) *Vault {
	for _, v := range vaults {
		if v.Address == address {
			return v
		}
	}
	return nil
}

// SelectVaultByPubKey returns the vault if the public key is found
// returns the vault if the public key is found
func SelectVaultByPubKey(vaults []*Vault, pubKey string) *Vault {
	for _, v := range vaults {
		if v.PubKey == pubKey {
			return v
		}
	}

	return nil
}

// SelectVaultByAssetType returns the vault by the asset type
func SelectVaultByAssetType(vaults []*Vault, assetType AssetType) *Vault {
	for _, v := range vaults {
		if v.AssetType == assetType {
			return v
		}
	}

	return nil
}

// SelectVaultByPkScript returns the vault by the pk script
func SelectVaultByPkScript(vaults []*Vault, pkScript []byte) *Vault {
	chainCfg := sdk.GetConfig().GetBtcChainCfg()

	for _, v := range vaults {
		addr, err := btcutil.DecodeAddress(v.Address, chainCfg)
		if err != nil {
			continue
		}

		addrScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			continue
		}

		if bytes.Equal(addrScript, pkScript) {
			return v
		}
	}

	return nil
}
