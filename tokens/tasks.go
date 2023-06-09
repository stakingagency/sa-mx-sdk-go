package tokens

import (
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

var initialized = false

func (tok *Tokens) startTasks() {
	if tok.refreshInterval == utils.NoRefresh {
		return
	}

	go func() {
		for {
			startTime := time.Now().UnixNano()

			tok.refreshTokens()

			endTime := time.Now().UnixNano()
			waitTime := tok.refreshInterval - time.Duration(endTime-startTime)
			if waitTime > 0 {
				time.Sleep(waitTime)
			}
			initialized = true
		}
	}()
}

func (tok *Tokens) refreshTokens() {
	newTokens, err := tok.GetTokens()
	if err != nil {
		log.Error("get all tokens", "error", err, "function", "refreshTokens")
		return
	}

	tok.cachedEsdtsMut.Lock()
	if initialized {
		for ticker, newEsdt := range newTokens {
			oldEsdt := tok.cachedEsdts[ticker]
			tok.cachedEsdtsMut.Unlock()
			if oldEsdt == nil {
				if tok.newTokenIssuedCallback != nil {
					tok.newTokenIssuedCallback(ticker)
				}
			} else {
				if newEsdt.Supply != oldEsdt.Supply {
					if tok.tokenSupplyChangedCallback != nil {
						tok.tokenSupplyChangedCallback(ticker, oldEsdt.Supply, newEsdt.Supply)
					}
				}
				if newEsdt.IsPaused != oldEsdt.IsPaused {
					if tok.tokenStateChangedCallback != nil {
						tok.tokenStateChangedCallback(ticker, !newEsdt.IsPaused)
					}
				}
			}
			tok.cachedEsdtsMut.Lock()
			tok.cachedEsdts[ticker] = newEsdt
		}
	}
	tok.cachedEsdtsMut.Unlock()
}
