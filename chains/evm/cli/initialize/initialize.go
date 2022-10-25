package initialize

import (
	"fmt"
	kms "github.com/LampardNguyen234/evm-kms"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmclient"
	evmgaspricer "github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmgaspricer"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor/prepare"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor/signAndSend"
	"github.com/ChainSafe/chainbridge-core/crypto/secp256k1"
	"github.com/rs/zerolog/log"
)

func InitializeClient(
	url string,
	senderKeyPair *secp256k1.Keypair, kmsSigner kms.KMSSigner,
) (*evmclient.EVMClient, error) {
	var ethClient *evmclient.EVMClient
	var err error
	if senderKeyPair != nil {
		ethClient, err = evmclient.NewEVMClient(
			url, senderKeyPair.PrivateKey())

	} else if kmsSigner != nil {
		ethClient, err = evmclient.NewEVMClientWithKMSSigner(url, kmsSigner)
	} else {
		err = fmt.Errorf("either `senderKeyPair` or `kmsSigner` must be set")
	}
	if err != nil {
		log.Error().Err(fmt.Errorf("eth client initialization error: %v", err))
		return nil, err
	}

	return ethClient, nil
}

// Initialize transactor which is used for contract calls
// if --prepare flag value is set as true (from CLI) call data is outputted to stdout
// which can be used for multisig contract calls
func InitializeTransactor(
	gasPrice *big.Int,
	txFabric calls.TxFabric,
	client *evmclient.EVMClient,
	prepareFlag bool,
) (transactor.Transactor, error) {
	var trans transactor.Transactor
	if prepareFlag {
		trans = prepare.NewPrepareTransactor()
	} else {
		gasPricer := evmgaspricer.NewLondonGasPriceClient(
			client,
			&evmgaspricer.GasPricerOpts{UpperLimitFeePerGas: gasPrice},
		)
		trans = signAndSend.NewSignAndSendTransactor(txFabric, gasPricer, client)
	}

	return trans, nil
}
