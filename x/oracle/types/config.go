package types

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

type OracleConfig struct {
	Enable         bool   `toml:"enable"`
	BitcoinRpc     string `toml:"bitcoin_rpc"`
	BitcoinRpcUser string `toml:"bitcoin_rpc_user"`
	BitcoinRpcPass string `toml:"bitcoin_rpc_password"`
	HTTPPostMode   bool   `toml:"http_post_mode"`
	DisableTLS     bool   `toml:"disable_tls"`
}

func DefaultOracleConfig() OracleConfig {
	return OracleConfig{
		Enable:         false,
		BitcoinRpc:     "192.248.150.102:18332",
		BitcoinRpcUser: "side",
		BitcoinRpcPass: "12345678",
		HTTPPostMode:   true,
		DisableTLS:     true,
	}
}

// ReadWasmConfig reads the wasm specifig configuration
func ReadOracleConfig(opts servertypes.AppOptions) (OracleConfig, error) {
	cfg := DefaultOracleConfig()
	var err error

	// attach contract debugging to global "trace" flag
	if v := opts.Get(flagOracleEnable); v != nil {
		if cfg.Enable, err = cast.ToBoolE(v); err != nil {
			return cfg, err
		}
	}
	if v := opts.Get(flagOracleBitcoinRpc); v != nil {
		if cfg.BitcoinRpc, err = cast.ToStringE(v); err != nil {
			return cfg, err
		}
	}
	if v := opts.Get(flagOracleBitcoinRpcUser); v != nil {
		if cfg.BitcoinRpcUser, err = cast.ToStringE(v); err != nil {
			return cfg, err
		}
	}
	if v := opts.Get(flagOracleBitcoinRpcPass); v != nil {
		if cfg.BitcoinRpcPass, err = cast.ToStringE(v); err != nil {
			return cfg, err
		}
	}
	if v := opts.Get(flagOracleBitcoinRpcPost); v != nil {
		if cfg.HTTPPostMode, err = cast.ToBoolE(v); err != nil {
			return cfg, err
		}
	}
	if v := opts.Get(flagOracleBitcoinRpcSSL); v != nil {
		if cfg.DisableTLS, err = cast.ToBoolE(v); err != nil {
			return cfg, err
		}
	}
	StartProviders = cfg.Enable
	return cfg, validate(&cfg)
}

func validate(conf *OracleConfig) error {
	if len(conf.BitcoinRpc) == 0 {
		return ErrInvalidBitcoinRPC
	}
	// if matched, _ := regexp.MatchString("\\w+:[\\d]{2,5}", conf.BitcoinRpc); matched {
	// 	return ErrInvalidBitcoinRPC
	// }
	if len(conf.BitcoinRpcUser) == 0 {
		return ErrInvalidBitcoinRPCUser
	}
	if len(conf.BitcoinRpcPass) == 0 {
		return ErrInvalidBitcoinRPCPass
	}
	return nil
}
