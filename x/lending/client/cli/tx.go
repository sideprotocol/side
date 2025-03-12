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

	"github.com/sideprotocol/side/x/lending/types"
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

	cmd.AddCommand(CmdCreatePool())
	cmd.AddCommand(CmdAddLiquidity())
	cmd.AddCommand(CmdRemoveLiquidity())
	cmd.AddCommand(CmdApply())
	cmd.AddCommand(CmdSubmitLiquidationCet())
	cmd.AddCommand(CmdApprove())
	cmd.AddCommand(CmdCancel())
	cmd.AddCommand(CmdSubmitCancellationSignatures())
	cmd.AddCommand(CmdRedeem())
	cmd.AddCommand(CmdRepay())
	cmd.AddCommand(CmdSubmitRepaymentAdaptorSignatures())
	cmd.AddCommand(CmdSubmitLiquidationCetSignatures())
	cmd.AddCommand(CmdClose())
	cmd.AddCommand(CmdSubmitPrice())

	return cmd
}

func CmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [id] [lending asset]",
		Short: "Create a lending pool with the specified asset",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreatePool(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
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

func CmdAddLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [pool id] [amount]",
		Short: "Add liquidity to the specified lending pool",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgAddLiquidity(
				clientCtx.GetFromAddress().String(),
				args[0],
				amount,
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

func CmdRemoveLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [shares]",
		Short: "Remove liquidity by the specified shares",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			shares, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgRemoveLiquidity(
				clientCtx.GetFromAddress().String(),
				shares,
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

func CmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [btc public key] [secret hash] [maturity time] [final timeout] [pool id] [borrow amount] [agency id]",
		Short: "Apply loan with the related params",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			maturityTime, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return err
			}

			finalTimeout, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			borrowAmount, err := sdk.ParseCoinNormalized(args[5])
			if err != nil {
				return err
			}

			agencyId, err := strconv.ParseUint(args[6], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgApply(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				maturityTime,
				finalTimeout,
				args[4],
				borrowAmount,
				agencyId,
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

func CmdSubmitLiquidationCet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-liquidation-cet [loan id] [event id] [deposit tx] [liquidation cet] [adaptor signature]",
		Short: "Submit liquidation cet",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			eventId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitLiquidationCet(
				clientCtx.GetFromAddress().String(),
				args[0],
				eventId,
				args[2],
				args[3],
				args[4],
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

func CmdApprove() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [deposit tx id] [block hash] [proof]",
		Short: "Approve loan with the deposit tx",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgApprove(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				strings.Split(args[2], listSeparator),
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

func CmdCancel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [loan id] [tx] [signatures]",
		Short: "Cancel the given loan along with the cancellation tx",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancel(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
				strings.Split(args[2], listSeparator),
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

func CmdSubmitCancellationSignatures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-cancellation-signatures [loan id] [DCA signatures]",
		Short: "Submit the DCA signatures for the loan to be cancelled",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signatures := strings.Split(args[1], listSeparator)

			msg := types.NewMsgSubmitCancellationSignatures(
				clientCtx.GetFromAddress().String(),
				args[0],
				signatures,
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

func CmdRedeem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem [loan id] [loan secret]",
		Short: "Redeem the borrowed coin with the loan secret",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRedeem(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
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

func CmdRepay() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay [loan id] [adaptor point]",
		Short: "Repay loan with the adaptor point",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgRepay(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
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

func CmdSubmitRepaymentAdaptorSignatures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-repay-adaptor-signatures [loan id] [DCA adaptor signatures]",
		Short: "Submit the DCA adaptor signatures for loan repayment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signatures := strings.Split(args[1], listSeparator)

			msg := types.NewMsgSubmitRepaymentAdaptorSignatures(
				clientCtx.GetFromAddress().String(),
				args[0],
				signatures,
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

func CmdSubmitLiquidationCetSignatures() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-liquidation-signatures [loan id] [DCA signatures]",
		Short: "Submit the DCA liquidation signatures for the loan to be liquidated",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signatures := strings.Split(args[1], listSeparator)

			msg := types.NewMsgSubmitLiquidationCetSignatures(
				clientCtx.GetFromAddress().String(),
				args[0],
				signatures,
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

func CmdClose() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close [loan id] [repayment tx signature]",
		Short: "Close loan with the repayment tx signature",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClose(
				clientCtx.GetFromAddress().String(),
				args[0],
				args[1],
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

func CmdSubmitPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-price [price]",
		Short: "Submit btc-usd price for testing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitPrice(
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
