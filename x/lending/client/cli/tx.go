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
	cmd.AddCommand(CmdApprove())
	cmd.AddCommand(CmdRedeem())
	cmd.AddCommand(CmdRepay())
	cmd.AddCommand(CmdSubmitRepaymentAdaptorSignature())
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
		Use:   "apply [btc public key] [secret hash] [maturity time] [final timeout] [deposit tx] [pool id] [borrow amount] [event id] [agency id] [liquidation cet] [adaptor signature]",
		Short: "Apply loan with the related params",
		Args:  cobra.ExactArgs(11),
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

			borrowAmount, err := sdk.ParseCoinNormalized(args[6])
			if err != nil {
				return err
			}

			eventId, err := strconv.ParseUint(args[7], 10, 64)
			if err != nil {
				return err
			}

			agencyId, err := strconv.ParseUint(args[8], 10, 64)
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
				args[5],
				borrowAmount,
				eventId,
				agencyId,
				args[9],
				args[10],
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

func CmdSubmitRepaymentAdaptorSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-repay-adaptor-signature [loan id] [DCA adaptor signature]",
		Short: "Submit the DCA adaptor signature for loan repayment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitRepaymentAdaptorSignature(
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
