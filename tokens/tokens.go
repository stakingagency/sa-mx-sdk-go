package tokens

import (
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type (
	NewTokenIssuedCallbackFunc     func(ticker string)
	TokenStateChangedCallbackFunc  func(ticker string, newState bool)
	TokenSupplyChangedCallbackFunc func(ticker string, oldSupply float64, newSupply float64)
)

type Tokens struct {
	netMan             *network.NetworkManager
	esdtIssueScAccount *accounts.Account
	refreshInterval    time.Duration

	cachedEsdts    map[string]*data.ESDT
	cachedEsdtsMut sync.Mutex

	newTokenIssuedCallback     NewTokenIssuedCallbackFunc
	tokenStateChangedCallback  TokenStateChangedCallbackFunc
	tokenSupplyChangedCallback TokenSupplyChangedCallbackFunc
}

var log = logger.GetOrCreate("tokens")

func NewTokens(netMan *network.NetworkManager, refreshInterval time.Duration) (*Tokens, error) {
	esdtIssueScAcount, err := accounts.NewAccount(utils.EsdtIssueSC, netMan, 0)
	if err != nil {
		return nil, err
	}

	t := &Tokens{
		netMan:             netMan,
		esdtIssueScAccount: esdtIssueScAcount,
		refreshInterval:    refreshInterval,

		cachedEsdts: make(map[string]*data.ESDT),

		newTokenIssuedCallback:     nil,
		tokenStateChangedCallback:  nil,
		tokenSupplyChangedCallback: nil,
	}
	t.startTasks()

	return t, nil
}

func (tok *Tokens) SetNewTokenIssuedCallback(f NewTokenIssuedCallbackFunc) {
	tok.newTokenIssuedCallback = f
}

func (tok *Tokens) SetTokenStateChangedCallback(f TokenStateChangedCallbackFunc) {
	tok.tokenStateChangedCallback = f
}

func (tok *Tokens) SetTokenSupplyChangedCallback(f TokenSupplyChangedCallbackFunc) {
	tok.tokenSupplyChangedCallback = f
}

func (tok *Tokens) GetCachedTokens() (map[string]*data.ESDT, error) {
	if tok.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[string]*data.ESDT)
	tok.cachedEsdtsMut.Lock()
	for k, v := range tok.cachedEsdts {
		res[k] = v
	}
	tok.cachedEsdtsMut.Unlock()

	return res, nil
}

func (tok *Tokens) GetTokens() (map[string]*data.ESDT, error) {
	tokens := make(map[string]*data.ESDT)
	keys, err := tok.esdtIssueScAccount.GetAccountKeys("")
	if err != nil {
		return nil, err
	}

	for bTicker, bytes := range keys {
		bToken, err := hex.DecodeString(bTicker)
		if err != nil {
			continue
		}

		ticker := string(bToken)
		idx := 35
		allOk := true
		bName, idx, ok := utils.ParseByteArray(bytes, idx)
		allOk = allOk && ok
		_, idx, ok = utils.ParseByte(bytes, idx) // dummy 1
		allOk = allOk && ok
		bShort, idx, ok := utils.ParseByteArray(bytes, idx)
		allOk = allOk && ok
		_, idx, ok = utils.ParseByte(bytes, idx) // dummy 2
		allOk = allOk && ok
		bTokenType, idx, ok := utils.ParseByteArray(bytes, idx) // dummy 3 = token type (FungibleESDT)
		tokenType := string(bTokenType)
		if tokenType != "FungibleESDT" && string(tokenType) != "MetaESDT" {
			continue
		}
		allOk = allOk && ok
		_, idx, ok = utils.ParseByte(bytes, idx) // dummy 4
		allOk = allOk && ok
		for allOk {
			var dummy []byte
			dummy, idx, ok = utils.ParseByteArray(bytes, idx) // dummy n
			allOk = allOk && ok
			if len(dummy) > 1 { // reached supply
				break
			}
		}
		_, idx, ok = utils.ParseByte(bytes, idx) // dummy 5
		allOk = allOk && ok
		_, idx, ok = utils.ParseByteArray(bytes, idx) // dummy 6
		allOk = allOk && ok
		_, idx, ok = utils.ParseByte(bytes, idx) // dummy 7
		allOk = allOk && ok
		decimals, _, ok := utils.ParseByte(bytes, idx)
		allOk = allOk && ok

		if !allOk {
			continue
		}

		esdt := &data.ESDT{
			Name:        string(bName),
			Ticker:      ticker,
			ShortTicker: string(bShort),
			Decimals:    uint64(decimals),
			Type:        tokenType,
		}
		err = tok.getTokenMintInfo(esdt)
		if err != nil {
			continue
		}

		esdt.IsPaused, err = tok.IsTokenPaused(esdt.Name)
		if err != nil {
			continue
		}

		tokens[esdt.Ticker] = esdt
	}

	return tokens, nil
}

func (tok *Tokens) IsTokenPaused(ticker string) (bool, error) {
	args := []string{hex.EncodeToString([]byte(ticker))}
	res, err := tok.netMan.QueryScMultiIntResult(utils.EsdtIssueSC, "getTokenProperties", args)
	if err != nil {
		return false, err
	}

	if len(res) < 7 {
		return false, utils.ErrInvalidResponse
	}

	isPaused := res[6].String() == "IsPaused-true"

	return isPaused, nil
}

func (tok *Tokens) GetCachedTokenProperties(ticker string) (*data.ESDT, error) {
	if tok.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	tok.cachedEsdtsMut.Lock()
	token := tok.cachedEsdts[ticker]
	tok.cachedEsdtsMut.Unlock()

	if token == nil {
		var err error
		token, err = tok.GetTokenProperties(ticker)
		if err != nil {
			return nil, err
		}
	}

	return token, nil
}

func (tok *Tokens) GetTokenProperties(ticker string) (*data.ESDT, error) {
	args := []string{hex.EncodeToString([]byte(ticker))}
	res, err := tok.netMan.QueryScMultiIntResult(utils.EsdtIssueSC, "getTokenProperties", args)
	if err != nil {
		return nil, err
	}

	if len(res) < 7 {
		return nil, utils.ErrInvalidResponse
	}

	sDecimals := strings.TrimPrefix(res[5].String(), "NumDecimals-")
	decimals, err := strconv.ParseUint(sDecimals, 10, 64)
	if err != nil {
		return nil, utils.ErrInvalidResponse
	}

	esdt := &data.ESDT{
		Name:        res[0].String(),
		Type:        res[1].String(),
		Ticker:      ticker,
		ShortTicker: strings.Split(ticker, "-")[0],
		Decimals:    decimals,
		IsPaused:    res[6].String() == "IsPaused-true",
	}

	err = tok.getTokenMintInfo(esdt)
	if err != nil {
		return nil, err
	}

	return esdt, nil
}

func (tok *Tokens) getTokenMintInfo(token *data.ESDT) error {
	mintInfoResponse := &data.EsdtMintInfoResponse{}
	err := tok.netMan.QueryProxy("network/esdt/supply/"+token.Ticker, mintInfoResponse)
	if err != nil {
		return err
	}

	iSupply, _ := big.NewInt(0).SetString(mintInfoResponse.Data.Supply, 10)
	iMinted, _ := big.NewInt(0).SetString(mintInfoResponse.Data.Minted, 10)
	iBurned, _ := big.NewInt(0).SetString(mintInfoResponse.Data.Burned, 10)
	iInitialMinted, _ := big.NewInt(0).SetString(mintInfoResponse.Data.InitialMinted, 10)

	token.Supply = utils.Denominate(iSupply, int(token.Decimals))
	token.Minted = utils.Denominate(iMinted, int(token.Decimals))
	token.Burned = utils.Denominate(iBurned, int(token.Decimals))
	token.InitialMinted = utils.Denominate(iInitialMinted, int(token.Decimals))

	return nil
}
