package main

import (
	"fmt"

	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/examples/abi/salsaContract"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func main() {
	// instantiate the contract
	contract, err := salsaContract.NewSalsaContract(
		"erd1qqqqqqqqqqqqqpgqpk3qzj86tme9kzxdq87f2rdf5nlwsgvjvcqs5hke3x",
		"https://devnet-gateway.multiversx.com",
		"https://devnet-index.multiversx.com",
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// read the private key from the test wallet provided
	w := interactors.NewWallet()
	privateKey, err := w.LoadPrivateKeyFromPemFile("testWallet.pem")
	if err != nil {
		fmt.Println(err)
		return
	}

	address, _ := w.GetAddressFromPrivateKey(privateKey)

	// instantiate an account object for the wallet address
	addr, _ := address.AddressAsBech32String()
	account, err := accounts.NewAccount(addr, contract.GetNetworkManager(), utils.NoRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}

	// retrieve the wallet's eGLD balance
	balance, err := account.GetEgldBalance()
	if err != nil {
		fmt.Println(err)
		return
	}

	// retrieve the wallet's tokens balances
	tokensBalances, err := account.GetTokensBalances()
	if err != nil {
		fmt.Println(err)
		return
	}

	// get the LEGLD token identifier from the contract
	token, err := contract.GetLiquidTokenId()
	if err != nil {
		fmt.Println(err)
		return
	}

	// get user's reserve from the contract
	reserve, err := contract.GetUserReserve(address.AddressBytes())
	if err != nil {
		fmt.Println(err)
		return
	}

	fReserve := utils.Denominate(reserve, 18)

	// print retrieved info
	fmt.Printf("address %s\n", account.GetAddress())
	fmt.Printf("eGLD balance %.4f\n", balance)
	fmt.Printf("LEGLD balance %.4f\n", tokensBalances[string(token)])
	fmt.Printf("reserve %.2f\n", fReserve)

	// add 1 eGLD reserve to the contract
	fmt.Print("adding 1 eGLD reserve... ")
	err = contract.AddReserve(privateKey, 1, 10000000, nil, utils.AutoNonce)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("done!")
}
