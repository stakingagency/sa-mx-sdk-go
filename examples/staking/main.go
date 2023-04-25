package main

import (
	"fmt"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/staking"
)

const (
	proxyAddress    = "https://gateway.multiversx.com"
	providerAddress = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"
)

var stake *staking.Staking

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	stake, err = staking.NewStaking(netMan, time.Second*6)
	if err != nil {
		fmt.Println(err)
		return
	}

	stake.SetNewProviderCallback(newProvider)
	stake.SetProviderCapChangedCallback(capChanged)
	stake.SetProviderClosedCallback(closedProvider)
	stake.SetProviderFeeChangedCallback(feeChanged)
	stake.SetProviderNameChangedCallback(nameChanged)
	stake.SetProviderOwnerChangedCallback(ownerChanged)
	stake.SetProviderSpaceAvailableCallback(spaceAvailable)

	fmt.Println("watching providers info changes")
	lastAvailable := 0
	for {
		time.Sleep(time.Minute)
		provider, err := stake.GetCachedProviderConfig(providerAddress)
		if err != nil {
			continue
		}

		newAvailable := int(provider.MaxDelegationCap - provider.ActiveStake)
		if lastAvailable != newAvailable {
			fmt.Printf("%s has %v eGLD available\n", provider.Name, newAvailable)
		}
		lastAvailable = newAvailable
	}
}

func getProviderName(providerAddress string) string {
	name, _, _, err := stake.GetCachedMetaData(providerAddress)
	if err != nil || name == "" {
		return providerAddress
	}

	return name
}

func ownerChanged(providerAddress string, oldOwner string, newOwner string) {
	fmt.Printf("provider %s changed owner from %s to %s\n",
		getProviderName(providerAddress), oldOwner, newOwner)
}

func nameChanged(providerAddress string, oldName string, newName string) {
	fmt.Printf("provider %s changed name from %s to %s\n",
		getProviderName(providerAddress), oldName, newName)
}

func feeChanged(providerAddress string, oldFee float64, newFee float64) {
	fmt.Printf("provider %s changed fee from %.2f to %.2f\n",
		getProviderName(providerAddress), oldFee, newFee)
}

func capChanged(providerAddress string, oldCap float64, newCap float64) {
	fmt.Printf("provider %s changed max cap from %v to %v\n",
		getProviderName(providerAddress), int(oldCap), int(newCap))
}

func spaceAvailable(providerAddress string, spaceAvailable float64) {
	fmt.Printf("provider %s has %.2f space available\n",
		getProviderName(providerAddress), spaceAvailable)
}

func newProvider(providerAddress string) {
	fmt.Printf("new provider: %s\n", getProviderName(providerAddress))
}

func closedProvider(providerAddress string) {
	fmt.Printf("provider closed: %s\n", getProviderName(providerAddress))
}
