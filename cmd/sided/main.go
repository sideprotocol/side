package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/sideprotocol/side/app"
	"github.com/sideprotocol/side/app/params"
	"github.com/sideprotocol/side/cmd/sided/cmd"
)

func main() {
	params.SetAddressPrefixes()

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
