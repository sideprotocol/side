package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sideprotocol/side/x/btcbridge/types"
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
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSubmitFeeRate())
	cmd.AddCommand(CmdUpdateNonBtcRelayers())
	cmd.AddCommand(CmdUpdateFeeProviders())
	cmd.AddCommand(CmdWithdrawToBitcoin())
	cmd.AddCommand(CmdSubmitSignatures())
	cmd.AddCommand(CmdCompleteDKG())
	cmd.AddCommand(CmdCompleteRefreshing())

	return cmd
}

func CmdSubmitFeeRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-fee-rate [fee rate]",
		Short: "Submit the latest fee rate of the bitcoin network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			feeRate, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitFeeRate(
				clientCtx.GetFromAddress().String(),
				feeRate,
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

func CmdUpdateNonBtcRelayers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-non-btc-relayers [relayers]",
		Short: "Update trusted non-btc asset relayers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateTrustedNonBtcRelayers(
				clientCtx.GetFromAddress().String(),
				strings.Split(args[0], listSeparator),
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

func CmdUpdateFeeProviders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-fee-providers [fee providers]",
		Short: "Update trusted fee providers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateTrustedFeeProviders(
				clientCtx.GetFromAddress().String(),
				strings.Split(args[0], listSeparator),
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

// Withdraw To Bitcoin
func CmdWithdrawToBitcoin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [amount]",
		Short: "Withdraw bitcoin asset to the given sender",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			_, err = sdk.ParseCoinsNormalized(args[0])
			if err != nil {
				return fmt.Errorf("invalid amount")
			}

			msg := types.NewMsgWithdrawToBitcoin(
				clientCtx.GetFromAddress().String(),
				args[0],
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

func CmdSubmitSignatures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-signatures [txid] [signatures]",
		Short: "Submit the signatures of the given signing request",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitSignatures(
				clientCtx.GetFromAddress().String(),
				args[0],
				strings.Split(args[1], listSeparator),
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

// Complete DKG
func CmdCompleteDKG() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete-dkg [id] [vaults] [consensus pub key] [signature]",
		Short: "Complete dkg request with new vaults",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			vaults := strings.Split(args[1], listSeparator)

			msg := types.NewMsgCompleteDKG(
				clientCtx.GetFromAddress().String(),
				id,
				vaults,
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

// Complete refreshing
func CmdCompleteRefreshing() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complete-refreshing [id] [consensus pub key] [signature]",
		Short: "Complete refreshing with the corresponding signature",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCompleteRefreshing(
				clientCtx.GetFromAddress().String(),
				id,
				args[1],
				args[2],
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
