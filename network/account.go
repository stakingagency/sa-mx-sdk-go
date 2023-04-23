package network

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

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
