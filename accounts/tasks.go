package accounts

import (
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

var initialized = false

func (acc *Account) startTasks() {
	if acc.refreshInterval == utils.NoRefresh {
		return
	}

	go func() {
		for {
			startTime := time.Now().UnixNano()

			acc.refreshEgldBalance()
			acc.refreshTokensBalances()

			endTime := time.Now().UnixNano()
			waitTime := acc.refreshInterval - time.Duration(endTime-startTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			initialized = true
		}
	}()
}

func (acc *Account) refreshEgldBalance() {
	newEgldBalance, err := acc.GetEgldBalance()
	if err != nil {
		log.Error("get egld balance", "error", err, "address", acc.address, "function", "refreshEgldBalance")
		return
	}

	if newEgldBalance != acc.cachedEgldBalance && acc.egldBalanceChangedCallback != nil && initialized {
		acc.egldBalanceChangedCallback(acc.cachedEgldBalance, newEgldBalance)
	}
	acc.cachedEgldBalance = newEgldBalance
}

func (acc *Account) refreshTokensBalances() {
	newTokensBalances, err := acc.GetTokensBalances()
	if err != nil {
		log.Error("get tokens balances", "error", err, "address", acc.address, "function", "refreshTokensBalances")
		return
	}

	acc.cachedTokensBalancesMut.Lock()
	if acc.tokenBalanceChangedCallback != nil && initialized {
		for ticker, newBalance := range newTokensBalances {
			oldBalance := acc.cachedTokensBalances[ticker]
			acc.cachedTokensBalancesMut.Unlock()
			if oldBalance != newBalance {
				acc.tokenBalanceChangedCallback(ticker, oldBalance, newBalance)
				acc.cachedTokensBalances[ticker] = newBalance
			}
			acc.cachedTokensBalancesMut.Lock()
		}
		for ticker, oldBalance := range acc.cachedTokensBalances {
			newBalance := newTokensBalances[ticker]
			if oldBalance != newBalance {
				acc.cachedTokensBalancesMut.Unlock()
				acc.tokenBalanceChangedCallback(ticker, oldBalance, newBalance)
				acc.cachedTokensBalancesMut.Lock()
			}
		}
	}
	acc.cachedTokensBalances = newTokensBalances
	acc.cachedTokensBalancesMut.Unlock()
}
