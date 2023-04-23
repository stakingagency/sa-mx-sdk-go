package network

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/core"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/sa-mx-sdk-go/data"
	"github.com/sa-mx-sdk-go/utils"
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

func (nm *NetworkManager) GetAccountKeys(address string, prefix string) (map[string][]byte, error) {
	endpoint := fmt.Sprintf("%s/address/%s/keys", nm.proxyAddress, address)
	bytes, err := utils.GetHTTP(endpoint, "")
	if err != nil {
		log.Error("get http", "error", err, "endpoint", endpoint, "function", "getAccountKeys")
		return nil, err
	}

	response := &data.AccountKeys{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		log.Error("unmarshal http response (get)", "error", err, "endpoint", endpoint, "function", "getAccountKeys")
		return nil, err
	}

	if response.Error != "" {
		log.Error("http response (get)", "error", err, "endpoint", endpoint, "function", "getAccountKeys")
		return nil, errors.New(response.Error)
	}

	result := make(map[string][]byte)
	for key, value := range response.Data.Pairs {
		bv, err := hex.DecodeString(value)
		if err != nil {
			log.Error("decode string", "error", err, "endpoint", endpoint, "function", "getAccountKeys")
			return nil, err
		}

		if strings.HasPrefix(key, prefix) {
			result[key] = bv
		}
	}

	return result, nil
}

func (nm *NetworkManager) GetAccountKey(address string, key string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/address/%s/key/%s", nm.proxyAddress, address, key)
	bytes, err := utils.GetHTTP(endpoint, "")
	if err != nil {
		log.Error("get http", "error", err, "endpoint", endpoint, "function", "getAccountKey")
		return nil, err
	}

	response := &data.AccountKey{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		log.Error("unmarshal http response (get)", "error", err, "endpoint", endpoint, "function", "getAccountKey")
		return nil, err
	}

	if response.Error != "" {
		log.Error("http response (get)", "error", err, "endpoint", endpoint, "function", "getAccountKey")
		return nil, errors.New(response.Error)
	}

	return hex.DecodeString(response.Data.Value)
}
