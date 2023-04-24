package xexchange

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/tokens"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type (
	NewPairCallbackFunc          func(ticker1 string, ticker2 string)
	PairStateChangedCallbackFunc func(ticker1 string, ticker2 string, newState bool)
	DexStateChangedCallbackFunc  func(newState bool)
)

type XExchange struct {
	netMan          *network.NetworkManager
	routerScAccount *accounts.Account
	mxTokens        *tokens.Tokens
	refreshInterval time.Duration

	dexState       bool
	cachedPairs    map[string]*data.DexPair
	cachedPairsMut sync.Mutex

	newPairCallback          NewPairCallbackFunc
	pairStateChangedCallback PairStateChangedCallbackFunc
	dexStateChangedCallback  DexStateChangedCallbackFunc
}

var log = logger.GetOrCreate("xexchange")

func NewXExchange(netMan *network.NetworkManager, refreshInterval time.Duration) (*XExchange, error) {
	routerScAccount, err := accounts.NewAccount(utils.DexRouterSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	mxTokens, err := tokens.NewTokens(netMan, refreshInterval)
	if err != nil {
		return nil, err
	}

	xex := &XExchange{
		netMan:          netMan,
		routerScAccount: routerScAccount,
		mxTokens:        mxTokens,
		refreshInterval: refreshInterval,

		cachedPairs: make(map[string]*data.DexPair),

		newPairCallback:          nil,
		pairStateChangedCallback: nil,
		dexStateChangedCallback:  nil,
	}
	xex.startTasks()

	return xex, nil
}

func (xex *XExchange) SetNewPairCallback(f NewPairCallbackFunc) {
	xex.newPairCallback = f
}

func (xex *XExchange) SetPairStateChangedCallback(f PairStateChangedCallbackFunc) {
	xex.pairStateChangedCallback = f
}

func (xex *XExchange) SetDexStateChangedCallback(f DexStateChangedCallbackFunc) {
	xex.dexStateChangedCallback = f
}

func (xex *XExchange) GetCachedDexPairs() (map[string]*data.DexPair, error) {
	if xex.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[string]*data.DexPair)
	xex.cachedPairsMut.Lock()
	for k, v := range xex.cachedPairs {
		res[k] = v
	}
	xex.cachedPairsMut.Unlock()

	return res, nil
}

func (xex *XExchange) GetDexPairs() (map[string]*data.DexPair, error) {
	pairs := make(map[string]*data.DexPair)
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	prefix := hex.EncodeToString([]byte("pair_map.mapped"))
	keys, err := xex.routerScAccount.GetAccountKeys(prefix)
	if err != nil {
		log.Error("get account keys", "error", err, "account", xex.routerScAccount.GetAddress(), "function", "GetDexPairs")
		return nil, err
	}

	for key, value := range keys {
		bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
		if err != nil {
			log.Debug("parse keys", "error", "can not decode key", "key", key, "function", "GetDexPairs")
			continue
		}

		idx := 0
		ticker1, idx, ok := utils.ParseString(bytes, idx)
		allOk := ok
		ticker2, _, ok := utils.ParseString(bytes, idx)
		allOk = allOk && ok
		if !allOk {
			log.Debug("parse keys", "error", "can not decode key", "key", key, "function", "GetDexPairs")
			continue
		}

		contractAddress := conv.Encode(value)
		pairTicker := fmt.Sprintf("%s %s", ticker1, ticker2)
		pair, err := xex.getPairData(ticker1, ticker2, contractAddress)
		if err == nil {
			pairs[pairTicker] = pair
		}
	}

	return pairs, nil
}

func (xex *XExchange) getPairData(ticker1 string, ticker2 string, contractAddress string) (*data.DexPair, error) {
	account, err := accounts.NewAccount(contractAddress, xex.netMan, 0)
	if err != nil {
		return nil, err
	}

	keys, err := account.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	result := &data.DexPair{
		ContractAddress: contractAddress,
	}

	if xex.refreshInterval == utils.NoRefresh {
		result.Token1, err = xex.mxTokens.GetTokenProperties(ticker1)
		if err != nil {
			return nil, err
		}

		result.Token2, err = xex.mxTokens.GetTokenProperties(ticker2)
		if err != nil {
			return nil, err
		}
	} else {
		result.Token1, err = xex.mxTokens.GetCachedTokenProperties(ticker1)
		if err != nil {
			return nil, err
		}

		result.Token2, err = xex.mxTokens.GetCachedTokenProperties(ticker2)
		if err != nil {
			return nil, err
		}
	}

	balance, err := utils.GetKey("ELRONDesdt"+result.Token1.Ticker, keys)
	if err != nil {
		return nil, err
	}

	if len(balance) < 2 {
		return nil, utils.ErrInvalidResponse
	}

	result.Balance1 = big.NewInt(0).SetBytes(balance[2:])

	balance, err = utils.GetKey("ELRONDesdt"+result.Token2.Ticker, keys)
	if err != nil {
		return nil, err
	}

	if len(balance) < 2 {
		return nil, utils.ErrInvalidResponse
	}

	result.Balance2 = big.NewInt(0).SetBytes(balance[2:])

	result.Fee, err = utils.GetBigIntKey("total_fee_percent", keys)
	if err != nil {
		return nil, err
	}

	result.State, err = utils.GetBoolKey("state", keys)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (xex *XExchange) GetPairByTickers(ticker1 string, ticker2 string) (*data.DexPair, error) {
	searchBytes := []byte("pair_map.mapped")
	t1 := utils.EncodeString(ticker1)
	searchBytes = append(searchBytes, t1...)
	t2 := utils.EncodeString(ticker2)
	searchBytes = append(searchBytes, t2...)
	searchKey := hex.EncodeToString(searchBytes)
	acc, err := accounts.NewAccount(utils.DexRouterSC, xex.netMan, utils.NoRefresh)
	if err != nil {
		return nil, err
	}

	key, err := acc.GetAccountKey(searchKey)
	if err != nil {
		return nil, err
	}

	contractPubkey, _, ok := utils.ParsePubkey(key, 0)
	if !ok {
		return nil, utils.ErrInvalidResponse
	}

	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	contractAddress := conv.Encode(contractPubkey)

	pair, err := xex.getPairData(ticker1, ticker2, contractAddress)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (xex *XExchange) GetPairByContractAddress(contractAddress string) (*data.DexPair, error) {
	acc, err := accounts.NewAccount(contractAddress, xex.netMan, utils.NoRefresh)
	if err != nil {
		return nil, err
	}

	firstTokenKey, err := acc.GetAccountKey(hex.EncodeToString([]byte("first_token_id")))
	if err != nil {
		return nil, err
	}

	secondTokenKey, err := acc.GetAccountKey(hex.EncodeToString([]byte("second_token_id")))
	if err != nil {
		return nil, err
	}

	pair, err := xex.getPairData(string(firstTokenKey), string(secondTokenKey), contractAddress)
	if err != nil {
		return nil, err
	}

	return pair, nil
}

func (xex *XExchange) GetCachedPairByTickers(ticker1 string, ticker2 string) (*data.DexPair, error) {
	if xex.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	pairTicker := fmt.Sprintf("%s %s", ticker1, ticker2)
	xex.cachedPairsMut.Lock()
	pair := xex.cachedPairs[pairTicker]
	xex.cachedPairsMut.Unlock()

	if pair == nil {
		var err error
		pair, err = xex.GetPairByTickers(ticker1, ticker2)
		if err != nil {
			return nil, err
		}
	}

	return pair, nil
}

func (xex *XExchange) GetCachedPairByContractAddress(contractAddress string) (*data.DexPair, error) {
	if xex.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	var pair *data.DexPair
	xex.cachedPairsMut.Lock()
	for _, cachedPair := range xex.cachedPairs {
		if cachedPair.ContractAddress == contractAddress {
			pair = cachedPair
			break
		}
	}
	xex.cachedPairsMut.Unlock()

	if pair == nil {
		var err error
		pair, err = xex.GetPairByContractAddress(contractAddress)
		if err != nil {
			return nil, err
		}
	}

	return pair, nil
}
