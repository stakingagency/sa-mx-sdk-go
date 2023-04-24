package main

import (
	"fmt"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/exchanges/xexchange"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

const (
	proxyAddress = "https://gateway.multiversx.com"
)

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	xex, err := xexchange.NewXExchange(netMan, time.Minute)
	if err != nil {
		fmt.Println(err)
		return
	}

	pairs, err := xex.GetDexPairs()
	if err != nil {
		fmt.Println(err)
		return
	}

	n := len(pairs)
	fmt.Printf("%v pairs listed\n", n)

	xex.SetNewPairCallback(newPair)
	xex.SetPairStateChangedCallback(pairStateChanged)
	xex.SetDexStateChangedCallback(dexStateChanged)

	for {
		time.Sleep(time.Minute)
		pair, err := xex.GetCachedPairByTickers(utils.WEGLD, utils.USDC)
		if err != nil {
			continue
		}

		fmt.Printf("price for pair %v is %.6f\n", pair.ContractAddress, pair.GetPrice())
	}
}

func newPair(ticker1 string, ticker2 string) {
	fmt.Printf("new pair: %s - %s\n", ticker1, ticker2)
}

func pairStateChanged(ticker1 string, ticker2 string, newState bool) {
	if newState {
		fmt.Printf("pair enabled: %s - %s\n", ticker1, ticker2)
	} else {
		fmt.Printf("pair disabled: %s - %s\n", ticker1, ticker2)
	}
}

func dexStateChanged(newState bool) {
	if newState {
		fmt.Println("exchange resumed")
	} else {
		fmt.Println("exchange paused")
	}
}
