package main

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stakingagency/sa-mx-sdk-go/examples/abi/salsaContract"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func main() {
	contract, err := salsaContract.NewSalsaContract("erd1qqqqqqqqqqqqqpgqpk3qzj86tme9kzxdq87f2rdf5nlwsgvjvcqs5hke3x", "https://devnet-gateway.multiversx.com")
	if err != nil {
		fmt.Println(err)
		return
	}

	reserve, err := contract.GetAvailableEgldReserve()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("available reserve: %.4f\n", utils.Denominate(reserve, 18))

	user := "erd1nh838fctgya24c0y4taf2cmqr0zwg47drvj7jszm8afrce8q0j9q80cl8x"
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, logger.GetOrCreate("abi"))
	pubKey, _ := conv.Decode(user)
	userUndelegations, err := contract.GetUserUndelegations(pubKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, undelegation := range userUndelegations {
		fmt.Printf("amount %.4f in epoch %v\n", utils.Denominate(undelegation.Amount, 18), undelegation.Unbond_epoch)
	}
}
