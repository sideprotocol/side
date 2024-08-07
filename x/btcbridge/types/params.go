package types

import (
	"bytes"
	"encoding/hex"
	"time"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	// default reward epoch
	DefaultRewardEpoch = time.Duration(1209600) * time.Second // 14 days

	// default DKG timeout period
	DefaultDKGTimeoutPeriod = time.Duration(86400) * time.Second // 1 day

	// default TSS participant update transition period
	DefaultTSSParticipantUpdateTransitionPeriod = time.Duration(1209600) * time.Second // 14 days
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		Confirmations:           1,
		MaxAcceptableBlockDepth: 100,
		BtcVoucherDenom:         "sat",
		Vaults: []*Vault{{
			Address:   "",
			PubKey:    "",
			AssetType: AssetType_ASSET_TYPE_BTC,
			Version:   0,
		}, {
			Address:   "",
			PubKey:    "",
			AssetType: AssetType_ASSET_TYPE_RUNES,
			Version:   0,
		}},
		ProtocolLimits: &ProtocolLimits{
			BtcMinDeposit:  50000,     // 0.0005 BTC
			BtcMinWithdraw: 30000,     // 0.0003 BTC
			BtcMaxWithdraw: 500000000, // 5 BTC
		},
		ProtocolFees: &ProtocolFees{
			DepositFee:  8000,  // 0.00008 BTC
			WithdrawFee: 12000, // 0.00012 BTC
			Collector:   authtypes.NewModuleAddress(ModuleName).String(),
		},
		NetworkFee:  8000, // 0.00008 BTC
		RewardEpoch: &DefaultRewardEpoch,
		TssParams: &TSSParams{
			DkgTimeoutPeriod:                  &DefaultDKGTimeoutPeriod,
			ParticipantUpdateTransitionPeriod: &DefaultTSSParticipantUpdateTransitionPeriod,
		},
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

	if err := validateVaults(p.Vaults); err != nil {
		return err
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

// SelectVaultByAddress returns the vault by the address
func SelectVaultByAddress(vaults []*Vault, address string) *Vault {
	for _, v := range vaults {
		if v.Address == address {
			return v
		}
	}
	return nil
}

// SelectVaultByPubKey returns the vault by the public key
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

// validateVaults validates the given vaults
func validateVaults(vaults []*Vault) error {
	vaultMap := make(map[string]bool)

	for _, v := range vaults {
		if len(v.Address) != 0 {
			_, err := sdk.AccAddressFromBech32(v.Address)
			if err != nil {
				return err
			}

			if vaultMap[v.Address] {
				return ErrInvalidParams
			}

			vaultMap[v.Address] = true
		}

		if len(v.PubKey) != 0 {
			pkBytes, err := hex.DecodeString(v.PubKey)
			if err != nil {
				return err
			}

			_, err = secp256k1.ParsePubKey(pkBytes)
			if err != nil {
				return err
			}
		}

		if v.AssetType == AssetType_ASSET_TYPE_UNSPECIFIED {
			return ErrInvalidParams
		}
	}

	return nil
}
