package main

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
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

	acc, err := accounts.NewAccount(address, netMan, utils.NoRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err := acc.GetEgldBalance()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("balance of %s is %.4f eGLD\n", address, balance)
	tokens, err := acc.GetTokensBalances()
	if err != nil {
		fmt.Println(err)
		return
	}

	for token, balance := range tokens {
		fmt.Printf("token balance of %s is %.4f %s\n", address, balance, token)
	}
}
