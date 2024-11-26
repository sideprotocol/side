package types

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/psbt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgTransferVault{}

// ValidateBasic performs basic MsgTransferVault message validation.
func (m *MsgTransferVault) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if m.SourceVersion == m.DestVersion {
		return ErrInvalidVaultVersion
	}

	if m.AssetType == AssetType_ASSET_TYPE_UNSPECIFIED {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid asset type")
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
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "target number of utxos must be greater than 0")
		}
	}

	return nil
}
