package network

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
)

type NetworkManager struct {
	proxy  blockchain.Proxy
	netCfg *sdkData.NetworkConfig

	proxyAddress string
	indexAddress string
}

var (
	log    = logger.GetOrCreate("network")
	suite  = ed25519.NewEd25519()
	keyGen = signing.NewKeyGenerator(suite)
)

func NewNetworkManager(proxyAddress string, indexAddress string) (*NetworkManager, error) {
	args := blockchain.ArgsProxy{
		ProxyURL:            proxyAddress,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}
	proxy, err := blockchain.NewProxy(args)
	if err != nil {
		log.Error("create proxy", "error", err, "function", "NewNetworkManager")
		return nil, err
	}

	netCfg, err := proxy.GetNetworkConfig(context.Background())
	if err != nil {
		log.Error("get network config", "error", err, "function", "NewNetworkManager")
		return nil, err
	}

	nm := &NetworkManager{
		proxy:        proxy,
		netCfg:       netCfg,
		proxyAddress: proxyAddress,
		indexAddress: indexAddress,
	}

	return nm, nil
}
