package accounts

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type (
	EgldBalanceChangedCallbackFunc  func(oldBalance float64, newBalance float64)
	TokenBalanceChangedCallbackFunc func(ticker string, oldBalance float64, newBalance float64)
)

type Account struct {
	netMan          *network.NetworkManager
	address         string
	refreshInterval time.Duration

	cachedEgldBalance       float64
	cachedTokensBalances    map[string]float64
	cachedTokensBalancesMut sync.Mutex
	cachedEsdts             map[string]*data.ESDT

	egldBalanceChangedCallback  EgldBalanceChangedCallbackFunc
	tokenBalanceChangedCallback TokenBalanceChangedCallbackFunc
}

var log = logger.GetOrCreate("accounts")

func NewAccount(address string, nm *network.NetworkManager, refreshInterval time.Duration) (*Account, error) {
	acc := &Account{
		netMan:          nm,
		address:         address,
		refreshInterval: refreshInterval,

		cachedTokensBalances: make(map[string]float64),
		cachedEsdts:          make(map[string]*data.ESDT),

		egldBalanceChangedCallback:  nil,
		tokenBalanceChangedCallback: nil,
	}
	acc.startTasks()

	return acc, nil
}

func (acc *Account) SetEgldBalanceChangedCallback(f EgldBalanceChangedCallbackFunc) {
	acc.egldBalanceChangedCallback = f
}

func (acc *Account) SetTokenBalanceChangedCallback(f TokenBalanceChangedCallbackFunc) {
	acc.tokenBalanceChangedCallback = f
}

func (acc *Account) GetAddress() string {
	return acc.address
}

func (acc *Account) GetAccountKeys(prefix string) (map[string][]byte, error) {
	endpoint := fmt.Sprintf("address/%s/keys", acc.address)
	response := &data.AccountKeys{}
	err := acc.netMan.QueryProxy(endpoint, response)
	if err != nil {
		log.Error("query proxy", "error", err, "endpoint", endpoint, "function", "GetAccountKeys")
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
	endpoint := fmt.Sprintf("address/%s/key/%s", acc.address, key)
	response := &data.AccountKey{}
	err := acc.netMan.QueryProxy(endpoint, response)
	if err != nil {
		log.Error("query proxy", "error", err, "endpoint", endpoint, "function", "GetAccountKey")
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

func (acc *Account) GetCachedEgldBalance() (float64, error) {
	if acc.refreshInterval == utils.NoRefresh {
		return 0, utils.ErrRefreshIntervalNotSet
	}

	return acc.cachedEgldBalance, nil
}

func (acc *Account) GetTokensBalances() (map[string]float64, error) {
	prefix := hex.EncodeToString([]byte("ELRONDesdt"))
	keys, err := acc.GetAccountKeys(prefix)
	if err != nil {
		return nil, err
	}

	res := make(map[string]float64)
	for key, value := range keys {
		ticker := strings.TrimPrefix(key, prefix)
		decimals, err := acc.GetTokenDecimals(ticker)
		if err != nil {
			decimals = 18
		}
		res[ticker] = utils.Denominate(big.NewInt(0).SetBytes(value), decimals)
	}

	return res, nil
}

func (acc *Account) GetCachedTokensBalances() (map[string]float64, error) {
	if acc.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[string]float64)
	acc.cachedTokensBalancesMut.Lock()
	for k, v := range acc.cachedTokensBalances {
		res[k] = v
	}
	acc.cachedTokensBalancesMut.Unlock()

	return res, nil
}

func (acc *Account) GetTokenDecimals(ticker string) (int, error) {
	args := []string{hex.EncodeToString([]byte(ticker))}
	res, err := acc.netMan.QueryScMultiIntResult(utils.EsdtIssueSC, "getTokenProperties", args)
	if err != nil {
		return 0, err
	}

	if len(res) < 6 {
		return 0, utils.ErrInvalidResponse
	}

	sDecimals := strings.TrimPrefix(res[5].String(), "NumDecimals-")
	decimals, err := strconv.ParseUint(sDecimals, 10, 64)
	if err != nil {
		return 0, utils.ErrInvalidResponse
	}

	return int(decimals), nil

}
