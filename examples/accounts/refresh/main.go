package main

import (
	"fmt"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/network"
)

const (
	proxyAddress = "https://gateway.multiversx.com"
	address      = "erd1sdslvlxvfnnflzj42l8czrcngq3xjjzkjp3rgul4ttk6hntr4qdsv6sets"
)

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	acc, err := accounts.NewAccount(address, netMan, time.Second*6)
	if err != nil {
		fmt.Println(err)
		return
	}

	acc.SetEgldBalanceChangedCallback(egldBalanceChanged)
	acc.SetTokenBalanceChangedCallback(tokenBalanceChanged)

	fmt.Printf("watching account's balance for %s\n", address)
	for {
	}
}

func egldBalanceChanged(oldBalance float64, newBalance float64) {
	fmt.Printf("eGLD balance changed with %.4f from %.4f to %.4f\n",
		newBalance-oldBalance, oldBalance, newBalance)
}

func tokenBalanceChanged(ticker string, oldBalance float64, newBalance float64) {
	fmt.Printf("%s balance changed with %.4f from %.4f to %.4f\n",
		ticker, newBalance-oldBalance, oldBalance, newBalance)
}
