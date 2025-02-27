package types

import (
	"bytes"
	"encoding/hex"
	"sort"
	"time"

	secp256k1 "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (

	// default confirmation depth for bitcoin deposit transactions
	DefaultDepositConfirmationDepth = int32(6)

	// default confirmation depth for bitcoin withdrawal transactions
	DefaultWithdrawConfirmationDepth = int32(6)

	// default allowed maximum depth for bitcoin block reorganization
	DefaultMaxReorgDepth = int32(6)

	// default BTC voucher denom
	DefaultBtcVoucherDenom = "sat"

	// default period of validity for the fee rate provided by fee provider
	DefaultFeeRateValidityPeriod = int64(100) // 100 blocks

	// default maximum number of utxos used to build the signing request
	DefaultMaxUtxoNum = uint32(200)

	// default btc batch withdrawal period
	DefaultBtcBatchWithdrawPeriod = int64(10)

	// default maximum number of btc batch withdrawal per batch
	DefaultMaxBtcBatchWithdrawNum = uint32(100)

	// default DKG timeout period
	DefaultDKGTimeoutPeriod = time.Duration(86400) * time.Second // 1 day

	// default TSS participant update transition period; not used for now
	DefaultTSSParticipantUpdateTransitionPeriod = time.Duration(1209600) * time.Second // 14 days
)

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{
		DepositConfirmationDepth:  DefaultDepositConfirmationDepth,
		WithdrawConfirmationDepth: DefaultWithdrawConfirmationDepth,
		MaxReorgDepth:             DefaultMaxReorgDepth,
		MaxAcceptableBlockDepth:   100,
		BtcVoucherDenom:           DefaultBtcVoucherDenom,
		DepositEnabled:            true,
		WithdrawEnabled:           true,
		TrustedBtcRelayers:        []string{},
		TrustedNonBtcRelayers:     []string{},
		TrustedFeeProviders:       []string{},
		FeeRateValidityPeriod:     DefaultFeeRateValidityPeriod,
		Vaults:                    []*Vault{},
		WithdrawParams: WithdrawParams{
			MaxUtxoNum:             DefaultMaxUtxoNum,
			BtcBatchWithdrawPeriod: DefaultBtcBatchWithdrawPeriod,
			MaxBtcBatchWithdrawNum: DefaultMaxBtcBatchWithdrawNum,
		},
		ProtocolLimits: ProtocolLimits{
			BtcMinDeposit:  50000,     // 0.0005 BTC
			BtcMinWithdraw: 30000,     // 0.0003 BTC
			BtcMaxWithdraw: 500000000, // 5 BTC
		},
		ProtocolFees: ProtocolFees{
			DepositFee:  8000,  // 0.00008 BTC
			WithdrawFee: 12000, // 0.00012 BTC
			Collector:   authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		},
		TssParams: TSSParams{
			DkgTimeoutPeriod:                  DefaultDKGTimeoutPeriod,
			ParticipantUpdateTransitionPeriod: DefaultTSSParticipantUpdateTransitionPeriod,
		},
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams()
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateConfirmationAndReorgParams(p.DepositConfirmationDepth, p.WithdrawConfirmationDepth, p.MaxReorgDepth); err != nil {
		return err
	}

	if err := sdk.ValidateDenom(p.BtcVoucherDenom); err != nil {
		return err
	}

	if err := validateBtcRelayers(p.TrustedBtcRelayers); err != nil {
		return err
	}

	if err := validateNonBtcRelayers(p.TrustedNonBtcRelayers); err != nil {
		return err
	}

	if err := validateFeeProviders(p.TrustedFeeProviders); err != nil {
		return err
	}

	if err := validateFeeRateValidityPeriod(p.FeeRateValidityPeriod); err != nil {
		return err
	}

	if err := validateVaults(p.Vaults); err != nil {
		return err
	}

	if err := validateWithdrawParams(&p.WithdrawParams); err != nil {
		return err
	}

	if err := validateProtocolParams(&p.ProtocolLimits, &p.ProtocolFees); err != nil {
		return err
	}

	return validateTSSParams(&p.TssParams)
}

// SelectVaultByAddress returns the vault by the given address
func SelectVaultByAddress(vaults []*Vault, address string) *Vault {
	for _, v := range vaults {
		if v.Address == address {
			return v
		}
	}

	return nil
}

// SelectVaultByPubKey returns the vault by the given public key
func SelectVaultByPubKey(vaults []*Vault, pubKey string) *Vault {
	for _, v := range vaults {
		if v.PubKey == pubKey {
			return v
		}
	}

	return nil
}

// SelectVaultByAssetType returns the vault by the asset type of the highest version
func SelectVaultByAssetType(vaults []*Vault, assetType AssetType) *Vault {
	sort.SliceStable(vaults, func(i int, j int) bool {
		return vaults[i].Version > vaults[j].Version
	})

	for _, v := range vaults {
		if v.AssetType == assetType {
			return v
		}
	}

	return nil
}

// SelectVaultByPkScript returns the vault by the given pk script for convenience
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

// validateConfirmationAndReorgParams validates the given confirmation and reorg params
func validateConfirmationAndReorgParams(depositConfirmationDepth int32, withdrawConfirmationDepth int32, maxReorgDepth int32) error {
	if depositConfirmationDepth <= 0 || withdrawConfirmationDepth <= 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "confirmation depth must be greater than 0")
	}

	if maxReorgDepth <= 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "max reorg depth must be greater than 0")
	}

	return nil
}

// validateBtcRelayers validates the given btc relayers
func validateBtcRelayers(relayers []string) error {
	for _, relayer := range relayers {
		_, err := sdk.AccAddressFromBech32(relayer)
		if err != nil {
			return ErrInvalidRelayers
		}
	}

	return nil
}

// validateNonBtcRelayers validates the given non btc relayers
func validateNonBtcRelayers(relayers []string) error {
	for _, relayer := range relayers {
		_, err := sdk.AccAddressFromBech32(relayer)
		if err != nil {
			return ErrInvalidRelayers
		}
	}

	return nil
}

// validateFeeProviders validates the given fee providers
func validateFeeProviders(providers []string) error {
	for _, provider := range providers {
		_, err := sdk.AccAddressFromBech32(provider)
		if err != nil {
			return ErrInvalidFeeProviders
		}
	}

	return nil
}

// validateFeeRateValidityPeriod validates the given fee rate validity period
func validateFeeRateValidityPeriod(feeRateValidityPeriod int64) error {
	if feeRateValidityPeriod <= 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "fee rate validity period must be greater than 0")
	}

	return nil
}

// validateVaults validates the given vaults
func validateVaults(vaults []*Vault) error {
	vaultMap := make(map[string]bool)

	for _, v := range vaults {
		_, err := sdk.AccAddressFromBech32(v.Address)
		if err != nil {
			return err
		}

		if vaultMap[v.Address] {
			return errorsmod.Wrapf(ErrInvalidParams, "duplicate vault")
		}

		vaultMap[v.Address] = true

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
			return errorsmod.Wrapf(ErrInvalidParams, "invalid asset type")
		}
	}

	return nil
}

// validateWithdrawParams validates the given withdrawal params
func validateWithdrawParams(withdrawParams *WithdrawParams) error {
	if withdrawParams.MaxUtxoNum == 0 || withdrawParams.BtcBatchWithdrawPeriod == 0 || withdrawParams.MaxBtcBatchWithdrawNum == 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid withdrawal params")
	}

	return nil
}

// validateProtocolParams validates the given protocol limits and fees
func validateProtocolParams(protocolLimits *ProtocolLimits, protocolFees *ProtocolFees) error {
	if protocolLimits.BtcMinWithdraw > protocolLimits.BtcMaxWithdraw {
		return errorsmod.Wrapf(ErrInvalidParams, "minimum btc withdrawal amount must not be greater than maximum withdrawal amount")
	}

	if (protocolFees.DepositFee != 0 || protocolFees.WithdrawFee != 0) && len(protocolFees.Collector) == 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol fee params")
	}

	if len(protocolFees.Collector) != 0 {
		_, err := sdk.AccAddressFromBech32(protocolFees.Collector)
		if err != nil {
			return errorsmod.Wrapf(ErrInvalidParams, "invalid protocol fee collector")
		}
	}

	return nil
}

// validateTSSParams validates the given TSS params
func validateTSSParams(params *TSSParams) error {
	if params.DkgTimeoutPeriod == 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid dkg timeout period")
	}

	if params.ParticipantUpdateTransitionPeriod == 0 {
		return errorsmod.Wrapf(ErrInvalidParams, "invalid participant update transition period")
	}

	return nil
}
