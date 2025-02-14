package cli

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sideprotocol/side/x/lending/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(_ string) *cobra.Command {
	// Group yield queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryCollateralAddress())
	cmd.AddCommand(CmdQueryLiquidationEvent())
	cmd.AddCommand(CmdQueryCETs())
	cmd.AddCommand(CmdQueryRepaymentTx())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryCollateralAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateral-address [btc public key] [hash of loan secret] [maturity time] [final timeout]",
		Short: "Query the collateral address by the specified loan params",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			maturityTime, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			finalTimeout, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.CollateralAddress(cmd.Context(), &types.QueryCollateralAddressRequest{
				BorrowerPubkey:   args[0],
				HashOfLoanSecret: args[1],
				MaturityTime:     maturityTime,
				FinalTimeout:     finalTimeout,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLiquidationEvent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-event [collateral amount] [borrowed amount]",
		Short: "Query the corresponding liquidation event according to the collateral amount and borrowed amount",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			collateralAmount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			borrowedAmount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			res, err := queryClient.LiquidationEvent(cmd.Context(), &types.QueryLiquidationEventRequest{
				BorrowAmount:      &borrowedAmount,
				CollateralAcmount: &collateralAmount,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryCETs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cets [loan id]",
		Short: "Query CETs of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LoanCETs(cmd.Context(), &types.QueryLoanCetsRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryRepaymentTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repayment-tx [loan id]",
		Short: "Query the unsigned btc repayment tx of the repaid loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.UnsignedPaymentTx(cmd.Context(), &types.QueryRepaymentTxRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
