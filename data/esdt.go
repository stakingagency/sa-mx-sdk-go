package data

type ESDT struct {
	Name        string
	Ticker      string
	ShortTicker string
	Decimals    uint64
	Type        string
	IsPaused    bool

	Supply        float64
	Minted        float64
	Burned        float64
	InitialMinted float64
}

type EsdtMintInfoResponse struct {
	Data  EsdtMintInfo `json:"data"`
	Error string       `json:"error"`
	Code  string       `json:"code"`
}

type EsdtMintInfo struct {
	Supply        string `json:"supply"`
	Minted        string `json:"minted"`
	Burned        string `json:"burned"`
	InitialMinted string `json:"initialMinted"`
}
