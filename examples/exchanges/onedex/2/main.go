package main

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/examples/exchanges/onedex/2/oneDex"
)

func main() {
	// instantiate the contract
	contract, err := oneDex.NewOneDex(
		"erd1qqqqqqqqqqqqqpgqqz6vp9y50ep867vnr296mqf3dduh6guvmvlsu3sujc",
		"https://gateway.multiversx.com",
		"https://index.multiversx.com",
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// get all the listed pairs IDs
	pairs, err := contract.GetPairIds()
	if err != nil {
		fmt.Println(err)
		return
	}

	// get the lp token identifier for each pair
	for _, pair := range pairs {
		lpToken, err := contract.GetPairLpTokenId(pair.Var1)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("pair id %v token1 %s token2 %s lp %s\n", pair.Var1, pair.Var0.Var0, pair.Var0.Var1, lpToken)
	}
}
