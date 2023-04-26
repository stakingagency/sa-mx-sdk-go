package staking

import (
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

var initialized = false

func (st *Staking) startTasks() {
	if st.refreshInterval == utils.NoRefresh {
		return
	}

	go func() {
		for {
			startTime := time.Now().UnixNano()

			st.refreshProviders()

			endTime := time.Now().UnixNano()
			waitTime := st.refreshInterval - time.Duration(endTime-startTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			initialized = true
		}
	}()
}

func (st *Staking) refreshProviders() {
	newProviders, err := st.GetProvidersConfigs()
	if err != nil {
		log.Error("get providers configs", "error", err, "function", "refreshProviders")
		return
	}

	st.cachedProvidersMut.Lock()
	if initialized {
		for address, newCfg := range newProviders {
			oldCfg := st.cachedProviders[address]
			st.cachedProvidersMut.Unlock()
			if oldCfg == nil {
				if st.newProviderCallback != nil {
					st.newProviderCallback(address)
				}
			} else {
				if newCfg.Owner != oldCfg.Owner && st.providerOwnerChangedCallback != nil {
					st.providerOwnerChangedCallback(address, oldCfg.Owner, newCfg.Owner)
				}
				if newCfg.Name != oldCfg.Name && st.providerNameChangedCallback != nil {
					st.providerNameChangedCallback(address, oldCfg.Name, newCfg.Name)
				}
				if newCfg.ServiceFee != oldCfg.ServiceFee && st.providerFeeChangedCallback != nil {
					st.providerFeeChangedCallback(address, oldCfg.ServiceFee, newCfg.ServiceFee)
				}
				if newCfg.MaxDelegationCap != oldCfg.MaxDelegationCap {
					if st.providerCapChangedCallback != nil {
						st.providerCapChangedCallback(address, oldCfg.MaxDelegationCap, newCfg.MaxDelegationCap)
					}
				}
				hadSpace := oldCfg.HasDelegationCap && oldCfg.ActiveStake <= oldCfg.MaxDelegationCap
				hasSpace := newCfg.HasDelegationCap && newCfg.ActiveStake <= newCfg.MaxDelegationCap
				if !hadSpace && hasSpace && newCfg.HasDelegationCap && st.providerSpaceAvailableCallback != nil {
					st.providerSpaceAvailableCallback(address, newCfg.MaxDelegationCap-newCfg.ActiveStake)
				}
			}
			st.cachedProvidersMut.Lock()
			st.cachedProviders[address] = newCfg
		}
		for address := range st.cachedProviders {
			if newProviders[address] == nil && st.providerClosedCallback != nil {
				st.cachedProvidersMut.Unlock()
				st.providerClosedCallback(address)
				st.cachedProvidersMut.Lock()
			}
		}
	}
	st.cachedProviders = newProviders
	st.cachedProvidersMut.Unlock()
}
