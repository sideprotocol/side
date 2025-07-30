package cli

import (
	"context"
	"encoding/hex"
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
	cmd.AddCommand(CmdQueryPool())
	cmd.AddCommand(CmdQueryPools())
	cmd.AddCommand(CmdQueryPoolExchangeRate())
	cmd.AddCommand(CmdQueryCollateralAddress())
	cmd.AddCommand(CmdQueryLiquidationPrice())
	cmd.AddCommand(CmdQueryDlcEventCount())
	cmd.AddCommand(CmdQueryLoan())
	cmd.AddCommand(CmdQueryLoans())
	cmd.AddCommand(CmdQueryLoanCetInfos())
	cmd.AddCommand(CmdQueryDlcMeta())
	cmd.AddCommand(CmdQueryDeposits())
	cmd.AddCommand(CmdQueryRedemption())
	cmd.AddCommand(CmdQueryRepayment())
	cmd.AddCommand(CmdQueryCurrentInterest())
	cmd.AddCommand(CmdQueryReferrer())
	cmd.AddCommand(CmdQueryReferrers())
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

func CmdQueryLiquidationPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liquidation-price [pool id] [collateral amount] [borrow amount] [maturity]",
		Short: "Query the liquidation price according to the given params",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			maturity, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.LiquidationPrice(cmd.Context(), &types.QueryLiquidationPriceRequest{
				PoolId:           args[0],
				CollateralAmount: args[1],
				BorrowAmount:     args[2],
				Maturity:         maturity,
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

func CmdQueryDlcEventCount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dlc-event-count",
		Short: "Query the available DLC event count",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DlcEventCount(cmd.Context(), &types.QueryDlcEventCountRequest{})
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
		Use:   "loan [id]",
		Short: "Query the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Loan(cmd.Context(), &types.QueryLoanRequest{Id: args[0]})
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
		Use:   "loans [status | address | oracle pub key]",
		Short: "Query loans by the given status, address, or oracle public key",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			if len(args) == 2 {
				_, err := sdk.AccAddressFromBech32(args[0])
				if err != nil {
					return err
				}

				status, err := strconv.ParseUint(args[1], 10, 32)
				if err != nil {
					return err
				}

				return queryLoansByAddress(clientCtx, cmd.Context(), args[0], types.LoanStatus(status))
			}

			status, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				_, err := sdk.AccAddressFromBech32(args[0])
				if err != nil {
					_, err := hex.DecodeString(args[0])
					if err != nil {
						return fmt.Errorf("no status, address, or oracle public key is provided")
					}

					return queryLoansByOracle(clientCtx, cmd.Context(), args[0])
				}

				return queryLoansByAddress(clientCtx, cmd.Context(), args[0], types.LoanStatus_Unspecified)
			}

			return queryLoansByStatus(clientCtx, cmd.Context(), types.LoanStatus(status))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryLoanCetInfos() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cet-infos [loan id]",
		Short: "Query the CET infos according to the given loan id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LoanCetInfos(cmd.Context(), &types.QueryLoanCetInfosRequest{
				LoanId: args[0],
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

func CmdQueryDeposits() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposits [loan id]",
		Short: "Query all deposit txs of the given loan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LoanDeposits(cmd.Context(), &types.QueryLoanDepositsRequest{LoanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryRedemption() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redemption [id]",
		Short: "Query redemption by the given id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Redemption(cmd.Context(), &types.QueryRedemptionRequest{Id: id})
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

func CmdQueryReferrer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "referrer [referral code]",
		Short: "Query the referrer by the given referral code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Referrer(cmd.Context(), &types.QueryReferrerRequest{ReferralCode: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryReferrers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "referrers",
		Short: "Query all registered referrers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Referrers(cmd.Context(), &types.QueryReferrersRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// queryLoansByStatus queries loans by the given status
func queryLoansByStatus(clientCtx client.Context, cmdCtx context.Context, status types.LoanStatus) error {
	queryClient := types.NewQueryClient(clientCtx)

	res, err := queryClient.Loans(cmdCtx, &types.QueryLoansRequest{
		Status: status,
	})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// queryLoansByAddress queries loans by the given address and status
func queryLoansByAddress(clientCtx client.Context, cmdCtx context.Context, address string, status types.LoanStatus) error {
	queryClient := types.NewQueryClient(clientCtx)

	res, err := queryClient.LoansByAddress(cmdCtx, &types.QueryLoansByAddressRequest{
		Address: address,
		Status:  status,
	})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}

// queryLoansByOracle queries loans by the given oracle
func queryLoansByOracle(clientCtx client.Context, cmdCtx context.Context, oraclePubKey string) error {
	queryClient := types.NewQueryClient(clientCtx)

	res, err := queryClient.LoansByOracle(cmdCtx, &types.QueryLoansByOracleRequest{
		OraclePubkey: oraclePubKey,
	})
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}
