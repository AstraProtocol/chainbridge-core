package bridge

import (
	"fmt"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/initialize"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/ChainSafe/chainbridge-core/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var retryCmd = &cobra.Command{
	Use:   "retry",
	Short: "Retry a transfer using tx hash",
	Long:  "The retry let account with retrier role to submit a retry transaction",
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
		return RetryCmd(cmd, args, bridge.NewBridgeContract(c, BridgeAddr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateRetryFlags(cmd, args)
		if err != nil {
			return err
		}

		ProcessRetryFlags(cmd, args)
		return nil
	},
}

func BindRetryFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Bridge, "bridge", "", "Bridge contract address")
	cmd.Flags().StringVar(&TxHash, "tx-hash", "", "Deposit transaction hash")
}

func init() {
	BindRetryFlags(retryCmd)
}
func ValidateRetryFlags(cmd *cobra.Command, args []string) error {
	if len(TxHash) != 66 {
		return fmt.Errorf("invalid tx hash %s", TxHash)
	}
	if !common.IsHexAddress(Bridge) {
		return fmt.Errorf("invalid bridge address %s", Bridge)
	}
	return nil
}

func ProcessRetryFlags(cmd *cobra.Command, args []string) {
	DepositTxHash = common.HexToHash(TxHash)
	BridgeAddr = common.HexToAddress(Bridge)
}
func RetryCmd(cmd *cobra.Command, args []string, contract *bridge.BridgeContract) error {
	log.Info().Msgf(
		"Retry deposit with tx hash %s on bridge %s",
		DepositTxHash.String(), BridgeAddr.String(),
	)
	_, err := contract.Retry(
		DepositTxHash, transactor.TransactOptions{GasLimit: gasLimit},
	)
	if err != nil {
		log.Error().Err(err)
		return err
	}
	log.Info().Msg("Retried")
	return nil
}
