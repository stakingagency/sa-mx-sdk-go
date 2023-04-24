package main

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/tokens"
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

	tok, err := tokens.NewTokens(netMan, utils.NoRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}

	allTokens, err := tok.GetTokens() // takes a while
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v tokens issued\n", len(allTokens))

	usdcProp, err := tok.GetTokenProperties(utils.USDC)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s token properties:\n", usdcProp.Ticker)
	fmt.Printf("name: %s\n", usdcProp.Name)
	fmt.Printf("short: %s\n", usdcProp.ShortTicker)
	fmt.Printf("decimals: %v\n", usdcProp.Decimals)
	fmt.Printf("type: %s\n", usdcProp.Type)
	fmt.Printf("paused: %v\n", usdcProp.IsPaused)
	fmt.Printf("supply: %v\n", int(usdcProp.Supply))
	fmt.Printf("minted: %v\n", int(usdcProp.Minted))
	fmt.Printf("burned: %v\n", int(usdcProp.Burned))
	fmt.Printf("initial: %v\n", int(usdcProp.InitialMinted))
}
