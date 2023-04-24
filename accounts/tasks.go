package accounts

import "time"

var initialized = false

func (acc *Account) startTasks() {
	if acc.refreshInterval == 0 {
		return
	}

	go func() {
		startTime := time.Now().UnixNano()

		acc.refreshEgldBalance()
		acc.refreshTokensBalances()

		endTime := time.Now().UnixNano()
		waitTime := acc.refreshInterval - time.Duration(endTime-startTime)
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
		initialized = true
	}()
}

func (acc *Account) refreshEgldBalance() {
	newEgldBalance, err := acc.GetEgldBalance()
	if err != nil {
		log.Error("get egld balance", "error", err, "address", acc.address, "function", "refreshEgldBalance")
		return
	}

	if acc.egldBalanceChangedCallback != nil && initialized {
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
			if oldBalance != newBalance {
				acc.tokenBalanceChangedCallback(ticker, oldBalance, newBalance)
				acc.cachedTokensBalances[ticker] = newBalance
			}
		}
		for ticker, oldBalance := range acc.cachedTokensBalances {
			newBalance := newTokensBalances[ticker]
			if oldBalance != newBalance {
				acc.tokenBalanceChangedCallback(ticker, oldBalance, newBalance)
			}
		}
	}
	acc.cachedTokensBalances = newTokensBalances
	acc.cachedTokensBalancesMut.Unlock()
}
