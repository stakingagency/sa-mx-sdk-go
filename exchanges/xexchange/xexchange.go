package xexchange

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type XExchange struct {
	netMan          *network.NetworkManager
	routerScAccount *accounts.Account
}

var log = logger.GetOrCreate("xexchange")

func NewXExchange(netMan *network.NetworkManager) (*XExchange, error) {
	routerScAccount, err := accounts.NewAccount(utils.DexRouterSC, netMan)
	if err != nil {
		return nil, err
	}

	xex := &XExchange{
		netMan:          netMan,
		routerScAccount: routerScAccount,
	}

	return xex, nil
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
	account, err := accounts.NewAccount(contractAddress, xex.netMan)
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

	result.Token1 = ticker1
	result.Token2 = ticker2

	balance, err := utils.GetKey("ELRONDesdt"+result.Token1, keys)
	if err != nil {
		return nil, err
	}

	if len(balance) < 2 {
		return nil, utils.ErrInvalidResponse
	}

	result.Balance1 = big.NewInt(0).SetBytes(balance[2:])

	balance, err = utils.GetKey("ELRONDesdt"+result.Token2, keys)
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
