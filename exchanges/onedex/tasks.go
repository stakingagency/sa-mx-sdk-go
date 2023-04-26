package onedex

import (
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

var initialized = false

func (one *OneDex) startTasks() {
	if one.refreshInterval == utils.NoRefresh {
		return
	}

	go func() {
		for {
			startTime := time.Now().UnixNano()

			one.refreshLiquidityPools()
			one.refreshFarms()
			one.refreshStakes()
			one.refreshLaunchpads()

			endTime := time.Now().UnixNano()
			waitTime := one.refreshInterval - time.Duration(endTime-startTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			initialized = true
		}
	}()
}

func (one *OneDex) refreshLiquidityPools() {
	newLiquidityPools, err := one.GetLiquidityPools()
	if err != nil {
		log.Error("get liquidity pools", "error", err, "function", "refreshLiquidityPools")
		return
	}

	one.liquidityPoolsMut.Lock()
	if initialized {
		for id, newPool := range newLiquidityPools {
			oldPool := one.liquidityPools[id]
			one.liquidityPoolsMut.Unlock()
			if oldPool == nil {
				if one.newPairCallback != nil {
					one.newPairCallback(newPool.Token1.Ticker, newPool.Token2.Ticker)
				}
			} else {
				if oldPool.Enabled != newPool.Enabled && one.pairStateChangedCallback != nil {
					one.pairStateChangedCallback(newPool.Token1.Ticker, newPool.Token2.Ticker, newPool.Enabled)
				}
			}
			one.liquidityPoolsMut.Lock()
		}
	}
	one.liquidityPools = newLiquidityPools
	one.liquidityPoolsMut.Unlock()
}

func (one *OneDex) refreshFarms() {
	newFarms, err := one.GetFarms()
	if err != nil {
		log.Error("get farms", "error", err, "function", "refreshFarns")
		return
	}

	one.farmsMut.Lock()
	if initialized {
		for id, newFarm := range newFarms {
			oldFarm := one.farms[id]
			one.farmsMut.Unlock()
			if oldFarm == nil {
				if !newFarm.IsDual() {
					if one.newFarmCallback != nil {
						one.newFarmCallback(newFarm.LpToken.Ticker, newFarm.RewardToken1.Ticker)
					}
				} else {
					if one.newDualFarmCallback != nil {
						one.newDualFarmCallback(newFarm.LpToken.Ticker, newFarm.RewardToken1.Ticker, newFarm.RewardToken2.Ticker)
					}
				}
			} else {
				if oldFarm.AnnualRewardPerLP1 != newFarm.AnnualRewardPerLP1 && one.annualReward1ChangedCallback != nil {
					one.annualReward1ChangedCallback(id, oldFarm.AnnualRewardPerLP1, newFarm.AnnualRewardPerLP1)
				}
				if oldFarm.AnnualRewardPerLP2 != newFarm.AnnualRewardPerLP2 && one.annualReward2ChangedCallback != nil {
					one.annualReward2ChangedCallback(id, oldFarm.AnnualRewardPerLP2, newFarm.AnnualRewardPerLP2)
				}
			}
			one.farmsMut.Lock()
		}
	}
	one.farms = newFarms
	one.farmsMut.Unlock()
}

func (one *OneDex) refreshStakes() {
	newStakes, err := one.GetStakes()
	if err != nil {
		log.Error("get stakes", "error", err, "function", "refreshStakes")
		return
	}

	one.stakesMut.Lock()
	if initialized {
		for id, newStake := range newStakes {
			oldStake := one.stakes[id]
			one.stakesMut.Unlock()
			if oldStake == nil {
				if one.newStakeCallback != nil {
					one.newStakeCallback(newStake.Token.Ticker)
				}
			} else {
				if oldStake.APR != newStake.APR && one.stakeAprChangedCallback != nil {
					one.stakeAprChangedCallback(id, oldStake.APR, newStake.APR)
				}
			}
			one.stakesMut.Lock()
		}
	}
	one.stakes = newStakes
	one.stakesMut.Unlock()
}

func (one *OneDex) refreshLaunchpads() {
	newLaunchpads, err := one.GetLaunchpads()
	if err != nil {
		log.Error("get launchpads", "error", err, "function", "refreshLaunchpads")
		return
	}

	one.launchpadsMut.Lock()
	if initialized {
		for id, newLaunchpad := range newLaunchpads {
			oldLaunchpad := one.launchpads[id]
			one.launchpadsMut.Unlock()
			if oldLaunchpad == nil {
				if one.newLaunchpadCallback != nil {
					one.newLaunchpadCallback(newLaunchpad.Token)
				}
			} else {
				if !newLaunchpad.IsLive && oldLaunchpad.IsLive && one.launchpadEndedCallback != nil {
					one.launchpadEndedCallback(newLaunchpad.Token)
				}
			}
			one.launchpadsMut.Lock()
		}
	}
	one.launchpads = newLaunchpads
	one.launchpadsMut.Unlock()
}
