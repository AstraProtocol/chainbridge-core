package account

import (
	"fmt"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var getKMSSignerAddressCmd = &cobra.Command{
	Use:   "kms-address",
	Short: "Retrieve the KMSSigner address (if have)",
	Long:  "The address subcommand is used to retrieve the address of the KMSSigner if the KMSSigner is used. If no KMSSigner is used, it will return an empty string.",
	RunE:  kmsSignerAddress,
	PreRun: func(cmd *cobra.Command, args []string) {
		logger.LoggerMetadata(cmd.Name(), cmd.Flags())
	},
	PersistentPreRunE: initGlobalFlags,
}

func kmsSignerAddress(_ *cobra.Command, _ []string) error {
	if kmsSigner == nil {
		return fmt.Errorf("no KMSSigner provided")
	}

	log.Info().Msgf("Address: %x", kmsSigner.GetAddress())
	return nil
}
