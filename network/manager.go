package network

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
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

func (nm *NetworkManager) GetProxy() blockchain.Proxy {
	return nm.proxy
}

func (nm *NetworkManager) GetProxyAddress() string {
	return nm.proxyAddress
}

func (nm *NetworkManager) GetIndexAddress() string {
	return nm.indexAddress
}

func (nm *NetworkManager) GetNetworkConfig() *sdkData.NetworkConfig {
	return nm.netCfg
}

func (nm *NetworkManager) querySC(scAddress, funcName string, args []string) (*sdkData.VmValuesResponseData, error) {
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

	return res, nil
}

func (nm *NetworkManager) QueryScIntResult(scAddress, funcName string, args []string) (*big.Int, error) {
	res, err := nm.querySC(scAddress, funcName, args)
	if err != nil {
		return nil, err
	}

	if len(res.Data.ReturnData) == 0 {
		return big.NewInt(0), nil
	}

	return big.NewInt(0).SetBytes(res.Data.ReturnData[0]), nil
}

func (nm *NetworkManager) QueryScMultiIntResult(scAddress, funcName string, args []string) ([]*big.Int, error) {
	res, err := nm.querySC(scAddress, funcName, args)
	if err != nil {
		return nil, err
	}

	ints := make([]*big.Int, 0)
	for _, b := range res.Data.ReturnData {
		ints = append(ints, big.NewInt(0).SetBytes(b))
	}

	return ints, nil
}

func (nm *NetworkManager) QueryScAddressResult(scAddress, funcName string, args []string) (string, error) {
	res, err := nm.querySC(scAddress, funcName, args)
	if err != nil {
		return "", err
	}

	if len(res.Data.ReturnData) == 0 {
		return "", nil
	}

	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	address := converter.Encode(res.Data.ReturnData[0])

	return address, nil
}

func (nm *NetworkManager) QueryProxy(path string, value interface{}) error {
	endpoint := fmt.Sprintf("%s/%s", nm.proxyAddress, path)
	res, err := utils.GetHTTP(endpoint, "")
	if err != nil {
		return err
	}

	err = json.Unmarshal(res, value)
	if err != nil {
		return err
	}

	return nil
}
