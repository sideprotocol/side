package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"

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
	cmd.AddCommand(CmdQueryPool())
	cmd.AddCommand(CmdQueryPools())
	cmd.AddCommand(CmdQueryCollateralAddress())
	cmd.AddCommand(CmdQueryLiquidationEvent())
	cmd.AddCommand(CmdQueryLiquidationCet())
	cmd.AddCommand(CmdQueryLoan())
	cmd.AddCommand(CmdQueryLoans())
	cmd.AddCommand(CmdQueryDlcMeta())
	cmd.AddCommand(CmdQueryRepayment())
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

func CmdQueryLiquidationCet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-cet [loan id] [borrower public key] [agency public key]",
		Short: "Query the liquidation CET info according to the given loan id or public keys",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			if len(args) == 1 {
				res, err := queryClient.LiquidationCet(cmd.Context(), &types.QueryLiquidationCetRequest{
					LoanId: args[0],
				})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			borrowerPubKey, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			_, err = schnorr.ParsePubKey(borrowerPubKey)
			if err != nil {
				return err
			}

			agencyPubKey, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			_, err = schnorr.ParsePubKey(agencyPubKey)
			if err != nil {
				return err
			}

			res, err := queryClient.LiquidationCet(cmd.Context(), &types.QueryLiquidationCetRequest{
				BorrowerPubkey: args[0],
				AgencyPubkey:   args[1],
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
