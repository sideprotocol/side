package keeper

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/sideprotocol/side/x/dlc/types"
	tsstypes "github.com/sideprotocol/side/x/tss/types"
)

// DKGCompletedHandler is callback handler when the DKG request completed by TSS
func (k Keeper) DKGCompletedHandler(ctx sdk.Context, id uint64, ty string, intent int32, pubKeys []string) error {
	switch ty {
	case types.DKG_TYPE_DCM:
		return k.CreateDCM(ctx, pubKeys[0])

	case types.DKG_TYPE_NONCE:
		// the first pub key is oracle and the remaining are nonces

		if err := k.CreateOracle(ctx, pubKeys[0]); err != nil {
			return err
		}

		return k.HandleNonces(ctx, pubKeys[0], pubKeys[1:], intent)
	}

	return nil
}

// SigningCompletedHandler is callback handler when the signing request completed by TSS
func (k Keeper) SigningCompletedHandler(ctx sdk.Context, sender string, id uint64, scopedId string, ty tsstypes.SigningType, intent int32, pubKey string, signatures []string) error {
	return k.HandleAttestation(ctx, sender, types.FromScopedId(scopedId), signatures[0])
}

// GetOracleParticipants gets oracle participants by voting power
func (k Keeper) GetOracleParticipants(ctx sdk.Context) ([]string, error) {
	consensusPubKeys := []string{}

	var errIterate error

	k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		if len(consensusPubKeys) == int(k.OracleParticipantNum(ctx)) {
			return true
		}

		pk, err := validator.ConsPubKey()
		if err != nil {
			errIterate = err
			return true
		}

		consensusPubKeys = append(consensusPubKeys, base64.StdEncoding.EncodeToString(pk.Bytes()))

		return false
	})

	if errIterate != nil {
		return nil, errIterate
	}

	if len(consensusPubKeys) < int(types.MinOracleParticipantNum) {
		return nil, types.ErrInsufficientOracleParticipants
	}

	return consensusPubKeys, nil
}
