package cli

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sideprotocol/side/x/dlc/types"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

const (
	// flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	listSeparator = ","
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "DLC transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSubmitOraclePubKey())
	cmd.AddCommand(CmdSubmitDCMPubKey())
	cmd.AddCommand(CmdSubmitNonce())

	return cmd
}

func CmdSubmitOraclePubKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-oracle-pubkey [pub key] [oracle id] [oracle pub key] [signature]",
		Short: "Submit the oracle public key",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			oracleId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitOraclePubKey(
				clientCtx.GetFromAddress().String(),
				args[0],
				oracleId,
				args[2],
				args[3],
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSubmitDCMPubKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dcm-pubkey [pub key] [dcm id] [dcm pub key] [signature]",
		Short: "Submit the DCM public key",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dcmId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitDCMPubKey(
				clientCtx.GetFromAddress().String(),
				args[0],
				dcmId,
				args[2],
				args[3],
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdSubmitNonce() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-nonce [event type] [nonce] [oracle pub key] [signature]",
		Short: "Submit the nonce along with the signature",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			eventType, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitNonce(
				clientCtx.GetFromAddress().String(),
				types.DlcEventType(eventType),
				args[1],
				args[2],
				args[3],
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
