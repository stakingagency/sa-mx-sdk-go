package main

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/tokens"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

const (
	proxyAddress = "https://devnet-gateway.multiversx.com"

	pemFile  = "alice.pem"
	receiver = "erd1l453hd0gt5gzdp7czpuall8ggt2dcv5zwmfdf3sd3lguxseux2fsmsgldz"
	ticker   = "ONE-429c7d"
)

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	privateKey, err := interactors.NewWallet().LoadPrivateKeyFromPemFile(pemFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := netMan.SendTransaction(privateKey, receiver, 1, utils.AutoGasLimit, "Hello !", utils.AutoNonce)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("move balance tx hash: " + hash)
	time.Sleep(time.Second * 6)

	tok, err := tokens.NewTokens(netMan, utils.NoRefresh)
	if err != nil {
		fmt.Println(err)
		return
	}

	token, err := tok.GetTokenProperties(ticker)
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err = netMan.SendEsdtTransaction(privateKey, receiver, 1, utils.AutoGasLimit, token, "", utils.AutoNonce)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("esdt transfer tx hash: " + hash)
}
