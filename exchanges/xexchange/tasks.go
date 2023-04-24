package xexchange

import (
	"math/big"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

var initialized = false

func (xex *XExchange) startTasks() {
	if xex.refreshInterval == utils.NoRefresh {
		return
	}

	go func() {
		for {
			startTime := time.Now().UnixNano()

			xex.refreshPairs()

			endTime := time.Now().UnixNano()
			waitTime := xex.refreshInterval - time.Duration(endTime-startTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			initialized = true
		}
	}()
}

func (xex *XExchange) refreshPairs() {
	bNewState, err := xex.routerScAccount.GetAccountKey("state")
	if err != nil {
		log.Error("get dex state", "error", err, "function", "refreshPairs")
		return
	}

	newState := big.NewInt(0).SetBytes(bNewState).Uint64() == 1
	if initialized && xex.dexStateChangedCallback != nil {
		xex.dexStateChangedCallback(newState)
	}
	xex.dexState = newState

	newPairs, err := xex.GetDexPairs()
	if err != nil {
		log.Error("get dex pairs", "error", err, "function", "refreshPairs")
		return
	}

	xex.cachedPairsMut.Lock()
	if initialized {
		for pairTicker, newPair := range newPairs {
			oldPair := xex.cachedPairs[pairTicker]
			if oldPair == nil {
				if xex.newPairCallback != nil {
					xex.newPairCallback(newPair.Token1.Ticker, newPair.Token2.Ticker)
				}
			} else {
				if newPair.State != oldPair.State && xex.pairStateChangedCallback != nil {
					xex.pairStateChangedCallback(newPair.Token1.Ticker, newPair.Token2.Ticker, newPair.State)
				}
			}
			xex.cachedPairs[pairTicker] = newPair
		}
	}
	xex.cachedPairs = newPairs
	xex.cachedPairsMut.Unlock()
}
