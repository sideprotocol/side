package cli

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/sideprotocol/side/x/farming/types"
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
		Short:                      "Farming transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdStake())
	cmd.AddCommand(CmdUnstake())
	cmd.AddCommand(CmdClaim())

	return cmd
}

func CmdStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stake [amount] [lock duration in days]",
		Short: "Stake the given amount for the specified lock duration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			lockDurationInDays, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgStake(
				clientCtx.GetFromAddress().String(),
				amount,
				time.Duration(lockDurationInDays)*24*time.Hour,
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

func CmdUnstake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unstake [staking id]",
		Short: "Unstake the given staking",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnstake(
				clientCtx.GetFromAddress().String(),
				id,
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

func CmdClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [staking id]",
		Short: "Claim the pending rewards by the given staking or all stakings",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				msg := types.NewMsgClaimAll(
					clientCtx.GetFromAddress().String(),
				)

				if err := msg.ValidateBasic(); err != nil {
					return err
				}

				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaim(
				clientCtx.GetFromAddress().String(),
				id,
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
