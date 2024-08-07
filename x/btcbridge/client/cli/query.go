package cli

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/sideprotocol/side/x/btcbridge/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	cmd.AddCommand(CmdBestBlock())
	cmd.AddCommand(CmdQueryBlock())
	cmd.AddCommand(CmdQueryWithdrawRequest())
	cmd.AddCommand(CmdQueryUTXOs())
	cmd.AddCommand(CmdQueryDKGRequests())
	cmd.AddCommand(CmdQueryDKGCompletionRequests())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryParams(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdBestBlock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "best-block",
		Short: "shows the best block header of the light client",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryChainTip(cmd.Context(), &types.QueryChainTipRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryBlock returns the command to query the heights of the light client
func CmdQueryBlock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [hash or height]",
		Short: "Query block by hash or height",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				res, err := queryClient.QueryBlockHeaderByHash(cmd.Context(), &types.QueryBlockHeaderByHashRequest{Hash: args[0]})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			res, err := queryClient.QueryBlockHeaderByHeight(cmd.Context(), &types.QueryBlockHeaderByHeightRequest{Height: height})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryWithdrawRequest returns the command to query withdrawal request
func CmdQueryWithdrawRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-request [status | address | tx hash]",
		Short: "Query withdrawal requests by status, address or tx hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			status, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				_, err = sdk.AccAddressFromBech32(args[0])
				if err != nil {
					_, err := chainhash.NewHashFromStr(args[0])
					if err != nil {
						return fmt.Errorf("invalid arg, neither status, address nor tx hash: %s", args[0])
					}

					res, err := queryClient.QueryWithdrawRequestByTxHash(cmd.Context(), &types.QueryWithdrawRequestByTxHashRequest{Txid: args[0]})
					if err != nil {
						return err
					}

					return clientCtx.PrintProto(res)
				}

				res, err := queryClient.QueryWithdrawRequestsByAddress(cmd.Context(), &types.QueryWithdrawRequestsByAddressRequest{Address: args[0]})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			res, err := queryClient.QueryWithdrawRequests(cmd.Context(), &types.QueryWithdrawRequestsRequest{Status: types.WithdrawStatus(status)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryUTXOs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "utxos [address]",
		Short: "query utxos with an optional address",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			if len(args) == 0 {
				res, err := queryClient.QueryUTXOs(cmd.Context(), &types.QueryUTXOsRequest{})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			_, err = sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.QueryUTXOsByAddress(cmd.Context(), &types.QueryUTXOsByAddressRequest{
				Address: args[0],
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

// CmdQueryDKGRequests returns the command to query DKG requests
func CmdQueryDKGRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dkg-requests [status]",
		Short: "Query dkg requests by the optional status",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			if len(args) > 0 {
				status, err := strconv.ParseInt(args[0], 10, 32)
				if err != nil {
					return err
				}

				res, err := queryClient.QueryDKGRequests(cmd.Context(), &types.QueryDKGRequestsRequest{Status: types.DKGRequestStatus(status)})
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			}

			res, err := queryClient.QueryAllDKGRequests(cmd.Context(), &types.QueryAllDKGRequestsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryDKGCompletionRequests returns the command to query DKG completion requests
func CmdQueryDKGCompletionRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dkg-completions [id]",
		Short: "Query dkg completion requests by the given request id",
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

			res, err := queryClient.QueryDKGCompletionRequests(cmd.Context(), &types.QueryDKGCompletionRequestsRequest{Id: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
