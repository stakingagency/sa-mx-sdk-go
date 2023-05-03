package main

import (
	"fmt"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/exchanges/onedex"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

const (
	proxyAddress = "https://gateway.multiversx.com"
)

var one *onedex.OneDex

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, "")
	if err != nil {
		fmt.Println(err)
		return
	}

	one, err = onedex.NewOneDex(netMan, time.Minute)
	if err != nil {
		fmt.Println(err)
		return
	}

	pools, err := one.GetLiquidityPools()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v liquidity pools\n", len(pools))

	farms, err := one.GetFarms()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v farms\n", len(farms))

	stakes, err := one.GetStakes()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v stakes\n", len(stakes))

	launchpads, err := one.GetLaunchpads()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v launchpads\n", len(launchpads))

	one.SetAnnualReward1ChangedCallback(annualReward1Changed)
	one.SetAnnualReward2ChangedCallback(annualReward2Changed)
	one.SetLaunchpadEndedCallback(launchpadEnded)
	one.SetNewDualFarmCallback(newDualFarm)
	one.SetNewFarmCallback(newFarm)
	one.SetNewLaunchpadCallback(newLaunchpad)
	one.SetNewPairCallback(newPair)
	one.SetNewStakeCallback(newStake)
	one.SetPairStateChangedCallback(pairStateChanged)
	one.SetStakeAprChangedCallback(stakeAprChanged)

	for {
		egldPrice := utils.GetBinancePrice("EGLD")
		price := one.GetCachedTokenPrice(onedex.OneToken, egldPrice)
		fmt.Printf("%s price is %.6f\n", onedex.OneToken, price)
		time.Sleep(time.Minute)
	}
}

func getFarmByID(id uint32) *onedex.Farm {
	farms, err := one.GetCachedFarms()
	if err != nil {
		panic(err)
	}

	return farms[id]
}

func getStakeByID(id uint32) *onedex.Stake {
	stakes, err := one.GetCachedStakes()
	if err != nil {
		panic(err)
	}

	return stakes[id]
}

func newPair(ticker1 string, ticker2 string) {
	fmt.Printf("new pair: %s - %s\n", ticker1, ticker2)
}

func pairStateChanged(ticker1 string, ticker2 string, newState bool) {
	fmt.Printf("pair %s - %s state is now %v\n", ticker1, ticker2, newState)
}

func newStake(ticker string) {
	fmt.Println("new stake: " + ticker)
}

func newFarm(lpTicker string, rewardTicker string) {
	fmt.Printf("new farm for %s with rewards in %s\n", lpTicker, rewardTicker)
}

func newDualFarm(lpTicker string, rewardTicker1 string, rewardTicker2 string) {
	fmt.Printf("new dual farm for %s with rewards in %s and %s\n", lpTicker, rewardTicker1, rewardTicker2)
}

func newLaunchpad(ticker string) {
	fmt.Println("new launchpad: " + ticker)
}

func launchpadEnded(ticker string) {
	fmt.Println("launchpad ended: " + ticker)
}

func annualReward1Changed(farmID uint32, oldReward float64, newReward float64) {
	farm := getFarmByID(farmID)
	if farm == nil {
		return
	}

	fmt.Printf("annual reward changed for farm %s from %.2f to %.2f\n", farm.LpToken.Name, oldReward, newReward)
}

func annualReward2Changed(farmID uint32, oldReward float64, newReward float64) {
	farm := getFarmByID(farmID)
	if farm == nil {
		return
	}

	fmt.Printf("annual reward changed for dual farm %s from %.2f to %.2f\n", farm.LpToken.Name, oldReward, newReward)
}

func stakeAprChanged(stakeID uint32, oldAPR float64, newAPR float64) {
	stake := getStakeByID(stakeID)
	if stake == nil {
		return
	}

	fmt.Printf("APR changed for stake %s from %.2f%% to %.2f%%\n", stake.Token.Name, oldAPR, newAPR)
}
