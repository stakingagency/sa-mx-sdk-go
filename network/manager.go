package network

import (
	"context"
	"math/big"
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

func (nm *NetworkManager) QueryScIntResult(scAddress, funcName string, args []string) (*big.Int, error) {
	if args == nil {
		args = make([]string, 0)
	}
	request := &sdkData.VmValueRequest{
		Address:  scAddress,
		FuncName: funcName,
		Args:     args,
	}
	res, err := nm.proxy.ExecuteVMQuery(context.Background(), request)
	if err != nil {
		return nil, err
	}

	if len(res.Data.ReturnData) == 0 {
		return big.NewInt(0), nil
	}

	return big.NewInt(0).SetBytes(res.Data.ReturnData[0]), nil
}

func (nm *NetworkManager) QueryScMultiIntResult(scAddress, funcName string, args []string) ([]*big.Int, error) {
	if args == nil {
		args = make([]string, 0)
	}
	request := &sdkData.VmValueRequest{
		Address:  scAddress,
		FuncName: funcName,
		Args:     args,
	}
	res, err := nm.proxy.ExecuteVMQuery(context.Background(), request)
	if err != nil {
		return nil, err
	}

	ints := make([]*big.Int, 0)
	for _, b := range res.Data.ReturnData {
		ints = append(ints, big.NewInt(0).SetBytes(b))
	}

	return ints, nil
}
