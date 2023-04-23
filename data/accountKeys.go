package data

type AccountKeys struct {
	Data struct {
		BlockInfo struct {
			Hash     string `json:"hash"`
			Nonce    uint64 `json:"nonce"`
			RootHash string `json:"rootHash"`
		} `json:"blockInfo"`
		Pairs map[string]string `json:"pairs"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type AccountKey struct {
	Data struct {
		Value string `json:"value"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
