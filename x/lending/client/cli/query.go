package cli

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

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
	cmd.AddCommand(CmdQueryPool())
	cmd.AddCommand(CmdQueryPools())
	cmd.AddCommand(CmdQueryPoolExchangeRate())
	cmd.AddCommand(CmdQueryCollateralAddress())
	cmd.AddCommand(CmdQueryLiquidationEvent())
	cmd.AddCommand(CmdQueryLoan())
	cmd.AddCommand(CmdQueryLoans())
	cmd.AddCommand(CmdQueryLoansByAddress())
	cmd.AddCommand(CmdQueryLoanCetInfos())
	cmd.AddCommand(CmdQueryDlcMeta())
	cmd.AddCommand(CmdQueryCancellation())
	cmd.AddCommand(CmdQueryRepayment())
	cmd.AddCommand(CmdQueryCurrentInterest())
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

func CmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [id]",
		Short: "Query the given lending pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pool(cmd.Context(), &types.QueryPoolRequest{
				Id: args[0],
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

func CmdQueryPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query all lending pools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pools(cmd.Context(), &types.QueryPoolsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryPoolExchangeRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate [pool id]",
		Short: "Query the current exchange rate of the given lending pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.PoolExchangeRate(cmd.Context(), &types.QueryPoolExchangeRateRequest{
				PoolId: args[0],
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

func CmdQueryCollateralAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateral-address [borrower public key] [dcm public key] [maturity time]",
		Short: "Query the collateral address by the specified loan params",
		Args:  cobra.ExactArgs(3),
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

			res, err := queryClient.CollateralAddress(cmd.Context(), &types.QueryCollateralAddressRequest{
				BorrowerPubkey: args[0],
				DCMPubKey:      args[1],
				MaturityTime:   maturityTime,
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
		Use:   "liquidation-event [pool id] [collateral amount] [borrow amount]",
		Short: "Query the corresponding liquidation event according to the collateral and borrow amounts",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LiquidationEvent(cmd.Context(), &types.QueryLiquidationEventRequest{
				PoolId:           args[0],
				CollateralAmount: args[1],
				BorrowAmount:     args[2],
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

func CmdQueryLoan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loan [loan id]",
		Short: "Query the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Loan(cmd.Context(), &types.QueryLoanRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLoans() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loans [status]",
		Short: "Query loans by the given status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			status, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			res, err := queryClient.Loans(cmd.Context(), &types.QueryLoansRequest{Status: types.LoanStatus(status)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLoansByAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loans-by-address [address] [status]",
		Short: "Query loans by the given address with the optional status",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			status := uint64(0)
			if len(args) == 2 {
				status, err = strconv.ParseUint(args[1], 10, 32)
				if err != nil {
					return err
				}
			}

			res, err := queryClient.LoansByAddress(cmd.Context(), &types.QueryLoansByAddressRequest{
				Address: args[0],
				Status:  types.LoanStatus(status)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLoanCetInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cet-infos [loan id] [collateral amount]",
		Short: "Query the liquidation CET info according to the given loan id or collateral amount",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			collateralAmount := ""
			if len(args) == 2 {
				collateralAmount = args[1]
			}

			res, err := queryClient.LoanCetInfos(cmd.Context(), &types.QueryLoanCetInfosRequest{
				LoanId:           args[0],
				CollateralAmount: collateralAmount,
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

func CmdQueryDlcMeta() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dlc-meta [loan id]",
		Short: "Query the related dlc meta of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LoanDlcMeta(cmd.Context(), &types.QueryLoanDlcMetaRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryCancellation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancellation [loan id]",
		Short: "Query the cancellation of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LoanCancellation(cmd.Context(), &types.QueryLoanCancellationRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryRepayment() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repayment [loan id]",
		Short: "Query repayment of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Repayment(cmd.Context(), &types.QueryRepaymentRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryCurrentInterest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current-interest [loan id]",
		Short: "Query the current interest of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.CurrentInterest(cmd.Context(), &types.QueryCurrentInterestRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
