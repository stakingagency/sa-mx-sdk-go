package tokens

import (
	"encoding/hex"

	"github.com/stakingagency/sa-mx-sdk-go/accounts"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type Tokens struct {
	netMan             *network.NetworkManager
	esdtIssueScAccount *accounts.Account
}

func NewTokens(netMan *network.NetworkManager) (*Tokens, error) {
	esdtIssueScAcount, err := accounts.NewAccount(utils.EsdtIssueSC, netMan)
	if err != nil {
		return nil, err
	}

	t := &Tokens{
		netMan:             netMan,
		esdtIssueScAccount: esdtIssueScAcount,
	}

	return t, nil
}

func (tok *Tokens) GetAllTokens() (map[string]*data.ESDT, error) {
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
		tokens[esdt.Ticker] = esdt
	}

	return tokens, nil
}
