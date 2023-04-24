package accounts

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	logger "github.com/multiversx/mx-chain-logger-go"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type Account struct {
	netMan  *network.NetworkManager
	address string
}

var log = logger.GetOrCreate("accounts")

func NewAccount(address string, nm *network.NetworkManager) (*Account, error) {
	acc := &Account{
		netMan:  nm,
		address: address,
	}

	return acc, nil
}

func (acc *Account) GetAddress() string {
	return acc.address
}

func (acc *Account) GetAccountKeys(prefix string) (map[string][]byte, error) {
	endpoint := fmt.Sprintf("%s/address/%s/keys", acc.netMan.GetProxyAddress(), acc.address)
	bytes, err := utils.GetHTTP(endpoint, "")
	if err != nil {
		log.Error("get http", "error", err, "endpoint", endpoint, "function", "GetAccountKeys")
		return nil, err
	}

	response := &data.AccountKeys{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		log.Error("unmarshal http response (get)", "error", err, "endpoint", endpoint, "function", "GetAccountKeys")
		return nil, err
	}

	if response.Error != "" {
		log.Error("http response (get)", "error", err, "endpoint", endpoint, "function", "GetAccountKeys")
		return nil, errors.New(response.Error)
	}

	result := make(map[string][]byte)
	for key, value := range response.Data.Pairs {
		bv, err := hex.DecodeString(value)
		if err != nil {
			log.Error("decode string", "error", err, "endpoint", endpoint, "function", "GetAccountKeys")
			return nil, err
		}

		if strings.HasPrefix(key, prefix) {
			result[key] = bv
		}
	}

	return result, nil
}

func (acc *Account) GetAccountKey(key string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/address/%s/key/%s", acc.netMan.GetProxyAddress(), acc.address, key)
	bytes, err := utils.GetHTTP(endpoint, "")
	if err != nil {
		log.Error("get http", "error", err, "endpoint", endpoint, "function", "GetAccountKey")
		return nil, err
	}

	response := &data.AccountKey{}
	err = json.Unmarshal(bytes, response)
	if err != nil {
		log.Error("unmarshal http response (get)", "error", err, "endpoint", endpoint, "function", "GetAccountKey")
		return nil, err
	}

	if response.Error != "" {
		log.Error("http response (get)", "error", err, "endpoint", endpoint, "function", "GetAccountKey")
		return nil, errors.New(response.Error)
	}

	return hex.DecodeString(response.Data.Value)
}

func (acc *Account) DNSResolve(herotag string) (string, error) {
	scAddress := utils.GetDNSAddress(herotag)
	args := []string{hex.EncodeToString([]byte(herotag))}
	address, err := acc.netMan.QueryScAddressResult(scAddress, "resolve", args)
	if err != nil {
		log.Error("query vm", "error", err, "function", "DNSResolve")
		return "", err
	}

	return address, nil
}

func (acc *Account) GetEgldBalance() (float64, error) {
	addr, _ := sdkData.NewAddressFromBech32String(acc.address)
	account, err := acc.netMan.GetProxy().GetAccount(context.Background(), addr)
	if err != nil {
		log.Error("get account", "error", err, "address", acc.address, "function", "GetEgldBalance")
		return 0, err
	}

	balance, err := account.GetBalance(acc.netMan.GetNetworkConfig().Denomination)
	if err != nil {
		log.Error("get balance", "error", err, "address", acc.address, "function", "GetEgldBalance")
		return 0, err
	}

	return balance, nil
}
