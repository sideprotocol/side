package types

import (
	"bytes"
	"strconv"

	"github.com/btcsuite/btcd/btcutil/psbt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgTransferVault = "transfer_vault"

// Route returns the route of MsgTransferVault.
func (msg *MsgTransferVault) Route() string {
	return RouterKey
}

// Type returns the type of MsgTransferVault.
func (msg *MsgTransferVault) Type() string {
	return TypeMsgTransferVault
}

// GetSignBytes implements the LegacyMsg interface.
func (m MsgTransferVault) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgTransferVault message.
func (m *MsgTransferVault) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic performs basic MsgTransferVault message validation.
func (m *MsgTransferVault) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if m.SourceVersion == m.DestVersion {
		return ErrInvalidVaultVersion
	}

	if m.AssetType == AssetType_ASSET_TYPE_UNSPECIFIED {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid asset type")
	}

	for _, p := range m.Psbts {
		packet, err := psbt.NewFromRawBytes(bytes.NewReader([]byte(p)), true)
		if err != nil {
			return err
		}

		if err := CheckTransactionWeight(packet.UnsignedTx, nil); err != nil {
			return err
		}

		for i, ti := range packet.UnsignedTx.TxIn {
			if ti.Sequence != MagicSequence {
				return ErrInvalidPsbt
			}

			if packet.Inputs[i].SighashType != DefaultSigHashType {
				return ErrInvalidPsbt
			}
		}

		for _, out := range packet.UnsignedTx.TxOut {
			if IsDustOut(out) {
				return ErrDustOutput
			}
		}
	}

	if len(m.Psbts) == 0 {
		if m.TargetUtxoNum == 0 {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "target number of utxos must be greater than 0")
		}

		if feeRate, err := strconv.ParseInt(m.FeeRate, 10, 64); err != nil || feeRate <= 0 {
			return ErrInvalidFeeRate
		}
	}

	return nil
}
