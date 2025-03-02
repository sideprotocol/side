package types

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
)

type OracleConfig struct {
	Enable bool `toml:"enable"`
}

func DefaultOracleConfig() OracleConfig {
	return OracleConfig{
		Enable: false,
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
	return cfg, nil
}
