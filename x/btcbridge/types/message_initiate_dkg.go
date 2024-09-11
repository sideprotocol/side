package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const TypeMsgInitiateDKG = "initiate_dkg"

// Route returns the route of MsgInitiateDKG.
func (msg *MsgInitiateDKG) Route() string {
	return RouterKey
}

// Type returns the type of MsgInitiateDKG.
func (msg *MsgInitiateDKG) Type() string {
	return TypeMsgInitiateDKG
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgInitiateDKG) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgInitiateDKG message.
func (m *MsgInitiateDKG) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic MsgInitiateDKG message validation.
func (m *MsgInitiateDKG) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if len(m.Participants) == 0 || m.Threshold == 0 || m.Threshold > uint32(len(m.Participants)) {
		return ErrInvalidDKGParams
	}

	for _, p := range m.Participants {
		if len(p.Moniker) > stakingtypes.MaxMonikerLength {
			return ErrInvalidDKGParams
		}

		if _, err := sdk.ValAddressFromBech32(p.OperatorAddress); err != nil {
			return sdkerrors.Wrap(err, "invalid operator address")
		}

		if _, err := sdk.ConsAddressFromHex(p.ConsensusAddress); err != nil {
			return sdkerrors.Wrap(err, "invalid consensus address")
		}
	}

	if len(m.VaultTypes) == 0 {
		return sdkerrors.Wrap(ErrInvalidDKGParams, "vault types can not be empty")
	}

	vaultTypes := make(map[AssetType]bool)

	for _, t := range m.VaultTypes {
		if t == AssetType_ASSET_TYPE_UNSPECIFIED {
			return sdkerrors.Wrap(ErrInvalidDKGParams, "invalid vault type")
		}

		if vaultTypes[t] {
			return sdkerrors.Wrap(ErrInvalidDKGParams, "duplicate vault type")
		}

		vaultTypes[t] = true
	}

	if m.EnableTransfer {
		if m.TargetUtxoNum == 0 {
			return sdkerrors.Wrap(ErrInvalidDKGParams, "target number of utxos must be greater than 0")
		}

		if feeRate, err := strconv.ParseInt(m.FeeRate, 10, 64); err != nil || feeRate <= 0 {
			return ErrInvalidFeeRate
		}
	}

	return nil
}
