package network

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func (nm *NetworkManager) GetDexPairs() (map[string]*data.DexPair, error) {
	pairs := make(map[string]*data.DexPair)
	conv, err := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	if err != nil {
		log.Error("GetDexPairs", "step", "create bech32 converter", "error", err)
		panic(nil)
	}

	prefix := hex.EncodeToString([]byte("pair_map.mapped"))
	keys, err := nm.GetAccountKeys(utils.DexRouterSC, prefix)
	if err != nil {
		return nil, err
	}

	for key, value := range keys {
		bytes, err := hex.DecodeString(strings.TrimPrefix(key, prefix))
		if err != nil {
			log.Debug("GetDexPairs", "step", "parse keys", "error", "can not decode key", "key", key)
			continue
		}

		idx := 0
		ticker1, idx, ok := utils.ParseString(bytes, idx)
		allOk := ok
		ticker2, _, ok := utils.ParseString(bytes, idx)
		allOk = allOk && ok
		if !allOk {
			log.Debug("GetDexPairs", "step", "parse keys", "error", "can not decode key", "key", key)
			continue
		}

		contractAddress := conv.Encode(value)
		pairTicker := fmt.Sprintf("%s %s", ticker1, ticker2)
		pair, err := nm.getPairData(ticker1, ticker2, contractAddress)
		if err == nil {
			pairs[pairTicker] = pair
		}
	}

	return pairs, nil
}

func (nm *NetworkManager) getPairData(ticker1 string, ticker2 string, contractAddress string) (*data.DexPair, error) {
	keys, err := nm.GetAccountKeys(contractAddress, "")
	if err != nil {
		return nil, err
	}

	result := &data.DexPair{
		ContractAddress: contractAddress,
	}

	result.Token1 = ticker1
	result.Token2 = ticker2

	balance, err := nm.getKey("ELRONDesdt"+result.Token1, keys)
	if err != nil || len(balance) < 2 {
		return nil, err
	}

	result.Balance1 = big.NewInt(0).SetBytes(balance[2:])

	balance, err = nm.getKey("ELRONDesdt"+result.Token2, keys)
	if err != nil || len(balance) < 2 {
		return nil, err
	}

	result.Balance2 = big.NewInt(0).SetBytes(balance[2:])

	result.Fee, err = nm.getBigIntKey("total_fee_percent", keys)
	if err != nil {
		return nil, err
	}

	result.State, err = nm.getBoolKey("state", keys)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (nm *NetworkManager) getKey(key string, keys map[string][]byte) ([]byte, error) {
	result, ok := keys[hex.EncodeToString([]byte(key))]
	if !ok {
		return nil, fmt.Errorf("%v key not found", key)
	}

	return result, nil
}

func (nm *NetworkManager) getBigIntKey(key string, keys map[string][]byte) (*big.Int, error) {
	result, err := nm.getKey(key, keys)
	if err != nil {
		return nil, err
	}

	return big.NewInt(0).SetBytes(result), nil
}

func (nm *NetworkManager) getBoolKey(key string, keys map[string][]byte) (bool, error) {
	result, err := nm.getKey(key, keys)
	if err != nil {
		return false, err
	}

	return len(result) == 1 && result[0] == 1, nil
}
