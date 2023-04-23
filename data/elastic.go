package data

type IndexerResult struct {
	ScrollId string `json:"_scroll_id"`
	PitId    string `json:"pit_id"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Hits []*IndexerEntry `json:"hits"`
	} `json:"hits"`
}

type IndexerEntry struct {
	Hash   string `json:"_id"`
	Source struct {
		// transactions
		Nonce         uint64 `json:"nonce"`
		Value         string `json:"value"`
		Receiver      string `json:"receiver"`
		Sender        string `json:"sender"`
		Data          []byte `json:"data"`
		Status        string `json:"status"`
		Operation     string `json:"operation"`
		Function      string `json:"function"`
		IsScCall      bool   `json:"isScCall"`
		HasScResults  bool   `json:"hasScResults"`
		HasLogs       bool   `json:"hasLogs"`
		HasOperations bool   `json:"hasOperations"`
		Timestamp     int64  `json:"timestamp"`

		// scresults
		Tokens     []string `json:"tokens"`
		EsdtValues []string `json:"esdtValues"`

		// logs
		Events []*IndexerEvent `json:"events"`
	} `json:"_source"`
	Sort []interface{} `json:"sort"`
}

type IndexerPitResponse struct {
	ID string `json:"id"`
}

type IndexerPitSearch struct {
	Size        uint16        `json:"size"`
	PIT         IndexerPit    `json:"pit"`
	Query       interface{}   `json:"query,omitempty"`
	Sort        []interface{} `json:"sort,omitempty"`
	SearchAfter []interface{} `json:"search_after,omitempty"`
}

type IndexerPit struct {
	ID        string `json:"id"`
	KeepAlive string `json:"keep_alive"`
}

type IndexerEvent struct {
	Identifier string   `json:"identifier"` // signalError, internalVMErrors
	Address    string   `json:"address"`
	Data       string   `json:"data"`
	Topics     []string `json:"topics"`
	Order      uint64   `json:"order"`
	Timestamp  int64    `json:"timestamp"`
}
