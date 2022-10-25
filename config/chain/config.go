package chain

import (
	"fmt"
	kms "github.com/LampardNguyen234/evm-kms"

	"github.com/ChainSafe/chainbridge-core/flags"
	"github.com/spf13/viper"
)

type GeneralChainConfig struct {
	Name           string `mapstructure:"name"`
	Id             *uint8 `mapstructure:"id"`
	Endpoint       string `mapstructure:"endpoint"`
	Type           string `mapstructure:"type"`
	Key            string
	Insecure       bool
	BlockstorePath string
	FreshStart     bool
	LatestBlock    bool
	KmsConfig      kms.Config
}

func (c *GeneralChainConfig) Validate() error {
	// viper defaults to 0 for not specified ints
	if c.Id == nil {
		return fmt.Errorf("required field domain.Id empty for chain %v", c.Id)
	}
	if c.Endpoint == "" {
		return fmt.Errorf("required field chain.Endpoint empty for chain %v", *c.Id)
	}
	if c.Name == "" {
		return fmt.Errorf("required field chain.Name empty for chain %v", *c.Id)
	}
	if c.Key == "" {
		if _, err := c.KmsConfig.IsValid(); err != nil {
			fmt.Println(c.KmsConfig)
			return fmt.Errorf("required either chain.Key non-empty or valid chain.KmsConfig")
		}
	}
	return nil
}

// UseKms indicates whether we should use a KMSSigner instead of private key.
func (c *GeneralChainConfig) UseKms() bool {
	return c.Key == ""
}

func (c *GeneralChainConfig) ParseFlags() {
	c.BlockstorePath = viper.GetString(flags.BlockstoreFlagName)
	c.FreshStart = viper.GetBool(flags.FreshStartFlagName)
	c.LatestBlock = viper.GetBool(flags.LatestBlockFlagName)
}
