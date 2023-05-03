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
	pairIDs, err := contract.GetPairIds()
	if err != nil {
		fmt.Println(err)
		return
	}

	// get for each pair, the first, second and lp tokens
	for _, pairID := range pairIDs {
		firstToken, err := contract.GetPairFirstTokenId(pairID)
		if err != nil {
			fmt.Println(err)
			return
		}

		secondToken, err := contract.GetPairSecondTokenId(pairID)
		if err != nil {
			fmt.Println(err)
			return
		}

		lpToken, err := contract.GetPairLpTokenId(pairID)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("pair id %v token1 %s token2 %s lp %s\n", pairID, firstToken, secondToken, lpToken)
	}
}
