package admin

import (
	"fmt"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/util"

	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/flags"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/initialize"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var addRetrierCmd = &cobra.Command{
	Use:   "add-retrier",
	Short: "Add a new retrier",
	Long:  "The add-retrier subcommand sets an address as a bridge retrier",
	PreRun: func(cmd *cobra.Command, args []string) {
		logger.LoggerMetadata(cmd.Name(), cmd.Flags())
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return util.CallPersistentPreRun(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := initialize.InitializeClient(url, senderKeyPair, kmsSigner)
		if err != nil {
			return err
		}
		t, err := initialize.InitializeTransactor(gasPrice, evmtransaction.NewTransaction, c, prepare)
		if err != nil {
			return err
		}
		return AddRetrierEVMCMD(cmd, args, bridge.NewBridgeContract(c, BridgeAddr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateAddRetrierFlags(cmd, args)
		if err != nil {
			return err
		}

		ProcessAddRetrierFlags(cmd, args)
		return nil
	},
}

func BindAddRetrierFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Retrier, "retrier", "", "Address to add")
	cmd.Flags().StringVar(&Bridge, "bridge", "", "Bridge contract address")
	flags.MarkFlagsAsRequired(cmd, "retrier", "bridge")
}

func init() {
	BindAddRetrierFlags(addRetrierCmd)
}

func ValidateAddRetrierFlags(cmd *cobra.Command, args []string) error {
	if !common.IsHexAddress(Retrier) {
		return fmt.Errorf("invalid retrier address %s", Retrier)
	}
	if !common.IsHexAddress(Bridge) {
		return fmt.Errorf("invalid bridge address %s", Bridge)
	}
	return nil
}

func ProcessAddRetrierFlags(cmd *cobra.Command, args []string) {
	RetrierAddr = common.HexToAddress(Retrier)
	BridgeAddr = common.HexToAddress(Bridge)
}

func AddRetrierEVMCMD(cmd *cobra.Command, args []string, contract *bridge.BridgeContract) error {
	log.Debug().Msgf(`
Adding retrier
Retrier address: %s
Bridge address: %s`, RetrierAddr, Bridge)
	_, err := contract.AddRetrier(RetrierAddr, transactor.TransactOptions{GasLimit: gasLimit})
	return err
}
