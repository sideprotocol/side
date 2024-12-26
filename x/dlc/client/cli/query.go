package cli

import (
	"fmt"
	"strconv"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/sideprotocol/side/x/dlc/types"
)

var (
	FlagTriggered = "triggered"
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
	cmd.AddCommand(CmdQueryOracles())
	cmd.AddCommand(CmdQueryAgencies())
	cmd.AddCommand(CmdQueryNonce())
	cmd.AddCommand(CmdQueryNonces())
	cmd.AddCommand(CmdQueryEvent())
	cmd.AddCommand(CmdQueryEvents())
	cmd.AddCommand(CmdQueryAttestation())
	cmd.AddCommand(CmdQueryAttestations())
	cmd.AddCommand(CmdQueryPrice())
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

func CmdQueryOracles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracles [status]",
		Short: "Query oracles by the given status",
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

			res, err := queryClient.Oracles(cmd.Context(), &types.QueryOraclesRequest{Status: types.DLCOracleStatus(status)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryAgencies() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agencies [status]",
		Short: "Query agencies by the given status",
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

			res, err := queryClient.Agencies(cmd.Context(), &types.QueryAgenciesRequest{Status: types.AgencyStatus(status)})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryNonce() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nonce [oracle id] [index]",
		Short: "Query the nonce by the oracle id and index",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			oracleId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			index, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Nonce(cmd.Context(), &types.QueryNonceRequest{OracleId: oracleId, Index: index})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryNonces() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nonces [oracle id]",
		Short: "Query all nonces of the given oracle",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			oracleId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Nonces(cmd.Context(), &types.QueryNoncesRequest{OracleId: oracleId})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryEvent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event [id]",
		Short: "Query the event by the given id",
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

			res, err := queryClient.Event(cmd.Context(), &types.QueryEventRequest{Id: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryEvents() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "events [flag]",
		Short:   "Query events by the given status",
		Args:    cobra.NoArgs,
		Example: "events --triggered",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			triggered, err := cmd.Flags().GetBool(FlagTriggered)
			if err != nil {
				return err
			}

			res, err := queryClient.Events(cmd.Context(), &types.QueryEventsRequest{Triggered: triggered})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(FlagTriggered, false, "Indicates if the events have been triggered")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryAttestation() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attestation [id]",
		Short: "Query the attestation by the given id",
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

			res, err := queryClient.Attestation(cmd.Context(), &types.QueryAttestationRequest{Id: id})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryAttestations() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attestations",
		Short: "Query all attestations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Attestations(cmd.Context(), &types.QueryAttestationsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdQueryPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [symbol]",
		Short: "Query the current price of the given symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Price(cmd.Context(), &types.QueryPriceRequest{Symbol: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
