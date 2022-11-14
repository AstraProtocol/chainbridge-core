package flags

import (
	"context"
	"encoding/hex"
	"fmt"
	kms "github.com/LampardNguyen234/evm-kms"
	"github.com/LampardNguyen234/evm-kms/gcpkms"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
	"math/big"
	"strings"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"

	"github.com/ChainSafe/chainbridge-core/keystore"
	"github.com/ChainSafe/chainbridge-core/types"

	"github.com/ChainSafe/chainbridge-core/crypto/secp256k1"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const DefaultGasLimit = 2000000

var (
	urlFlag           = "url"
	gasLimitFlag      = "gas-limit"
	gasPriceFlag      = "gas-price"
	prepareFlag       = "prepare"
	privateKeyFlag    = "private-key"
	kmsConfigFileFlag = "kms-config-file"
)

func BindGlobalFlags(rootCMD *cobra.Command) {
	rootCMD.PersistentFlags().String(urlFlag, "", "The RPC endpoint")
	_ = viper.BindPFlag(urlFlag, rootCMD.PersistentFlags().Lookup(urlFlag))

	rootCMD.PersistentFlags().Uint64(gasLimitFlag, DefaultGasLimit, "The gas limit")
	_ = viper.BindPFlag(gasLimitFlag, rootCMD.PersistentFlags().Lookup(gasLimitFlag))

	rootCMD.PersistentFlags().Uint64(gasPriceFlag, 0, "The gas price")
	_ = viper.BindPFlag(gasPriceFlag, rootCMD.PersistentFlags().Lookup(gasPriceFlag))

	rootCMD.PersistentFlags().Bool(prepareFlag, false, "The prepare flag (true for generating payload only)")
	_ = viper.BindPFlag(prepareFlag, rootCMD.PersistentFlags().Lookup(prepareFlag))

	rootCMD.PersistentFlags().String(privateKeyFlag, "", "The private key")
	_ = viper.BindPFlag(privateKeyFlag, rootCMD.PersistentFlags().Lookup(privateKeyFlag))

	rootCMD.PersistentFlags().String(kmsConfigFileFlag, "", "The path to the KMS config file")
	_ = viper.BindPFlag(kmsConfigFileFlag, rootCMD.PersistentFlags().Lookup(kmsConfigFileFlag))
}

func GlobalFlagValues(cmd *cobra.Command) (string, uint64, *big.Int, *secp256k1.Keypair, kms.KMSSigner, bool, error) {
	url, err := cmd.Flags().GetString(urlFlag)
	if err != nil {
		log.Error().Err(formatFlagError(urlFlag, err))
		return "", DefaultGasLimit, nil, nil, nil, false, err
	}
	ethClient, err := ethclient.Dial(url)
	if err != nil {
		log.Error().Err(fmt.Errorf("cannot dial `%v`", url))
	}

	gasLimitInt, err := cmd.Flags().GetUint64(gasLimitFlag)
	if err != nil {
		log.Error().Err(formatFlagError(gasLimitFlag, err))
		return "", DefaultGasLimit, nil, nil, nil, false, err
	}

	gasPriceInt, err := cmd.Flags().GetUint64(gasPriceFlag)
	if err != nil {
		log.Error().Err(formatFlagError(gasPriceFlag, err))
		return "", DefaultGasLimit, nil, nil, nil, false, err
	}
	var gasPrice *big.Int = nil
	if gasPriceInt != 0 {
		gasPrice = big.NewInt(0).SetUint64(gasPriceInt)
	}

	prepare, err := cmd.Flags().GetBool(prepareFlag)
	if err != nil {
		log.Error().Err(formatFlagError(prepareFlag, err))
		return "", DefaultGasLimit, nil, nil, nil, false, err
	}

	kmsConfigFile, _ := cmd.Flags().GetString(kmsConfigFileFlag)
	if kmsConfigFile == "" {
		senderKeyPair, err := defineSender(cmd)
		if err != nil {
			log.Error().Err(formatFlagError(privateKeyFlag, err))
			return "", DefaultGasLimit, nil, nil, nil, false, err
		}
		return url, gasLimitInt, gasPrice, senderKeyPair, nil, prepare, nil
	} else {
		kmsConfig, err := kms.LoadConfigFromJSONFile(kmsConfigFile)
		if err != nil {
			log.Error().Err(formatFlagError(kmsConfigFileFlag, err))
			return "", DefaultGasLimit, nil, nil, nil, false, err
		}
		kmsSigner, err := getKMSSignerFromConfig(kmsConfig)
		if err != nil {
			log.Error().Err(formatFlagError(kmsConfigFileFlag, err))
			return "", DefaultGasLimit, nil, nil, nil, false, err
		}
		if ethClient != nil {
			chainID, err := ethClient.ChainID(context.Background())
			if err != nil {
				log.Error().Err(fmt.Errorf("fail to retrieve chainID"))
				return "", DefaultGasLimit, nil, nil, nil, false, err
			}

			kmsSigner.WithSigner(types2.NewLondonSigner(chainID))
		}

		return url, gasLimitInt, gasPrice, nil, kmsSigner, prepare, nil
	}
}

func defineSender(cmd *cobra.Command) (*secp256k1.Keypair, error) {
	privateKey, err := cmd.Flags().GetString("private-key")
	if err != nil {
		return nil, err
	}
	if privateKey != "" {
		kp, err := secp256k1.NewKeypairFromString(privateKey)
		if err != nil {
			return nil, err
		}
		return kp, nil
	}
	var AliceKp = keystore.TestKeyRing.EthereumKeys[keystore.AliceKey]
	return AliceKp, nil
}

func ProcessResourceID(resourceID string) (types.ResourceID, error) {
	if resourceID[0:2] == "0x" {
		resourceID = resourceID[2:]
	}
	resourceIdBytes, err := hex.DecodeString(resourceID)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed decoding resourceID hex string: %s", err)
	}
	return calls.SliceTo32Bytes(resourceIdBytes), nil
}

func MarkFlagsAsRequired(cmd *cobra.Command, flags ...string) {
	for _, flag := range flags {
		err := cmd.MarkFlagRequired(flag)
		if err != nil {
			panic(err)
		}
	}
}

func getKMSSignerFromConfig(cfg *kms.Config) (kms.KMSSigner, error) {
	switch strings.ToLower(cfg.Type) {
	case "gcp":
		return gcpkms.NewGoogleKMSClient(context.Background(), cfg.GcpConfig)
	default:
		return nil, fmt.Errorf("KMS `%v` not yet supported", cfg.Type)
	}
}

func formatFlagError(flagName string, err error) error {
	return fmt.Errorf("`%v` error: %v", flagName, err)
}
