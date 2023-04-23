package data

type ESDT struct {
	Name        string
	Ticker      string
	ShortTicker string
	Decimals    uint64
	Type        string
}

type TokensList struct {
	Data struct {
		Tokens []string `json:"tokens"`
	} `json:"data"`
}

type WalletEsdts struct {
	Data struct {
		Esdts map[string]*WalletEsdtEntry `json:"esdts"`
	} `json:"data"`
}

type WalletEsdt struct {
	Data struct {
		TokenData WalletEsdtEntry `json:"tokenData"`
	} `json:"data"`
}

type WalletEsdtEntry struct {
	Identifier string `json:"tokenIdentifier"`
	Balance    string `json:"balance"`
}
