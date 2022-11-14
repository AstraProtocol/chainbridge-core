package utils

import (
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmgaspricer"
	"os"

	"github.com/spf13/cobra"
)

var UtilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Set of utility commands",
	Long:  "Set of utility commands",
}

func init() {
	UtilsCmd.AddCommand(simulateCmd)
	UtilsCmd.AddCommand(hashListCmd)
}

type GasPricerWithPostConfig interface {
	calls.GasPricer
	SetClient(client evmgaspricer.LondonGasClient)
	SetOpts(opts *evmgaspricer.GasPricerOpts)
}

func PrintSubCommandHelp(cmd *cobra.Command, args []string) {
	shouldPrintHelp := true
	if len(args) != 0 {
		for _, subCommand := range cmd.Commands() {
			if subCommand.Name() == args[0] {
				shouldPrintHelp = false
				break
			}
		}
	}

	if shouldPrintHelp {
		cmd.Help()
		os.Exit(0)
	}
}
