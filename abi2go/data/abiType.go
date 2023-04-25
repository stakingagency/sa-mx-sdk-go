package data

type ABI struct {
	BuildInfo struct {
		ContractCrate struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"contractCrate"`
		Framework struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"framework"`
	} `json:"buildInfo"`
	Name        string `json:"name"`
	Constructor struct {
		Inputs  []AbiEndpointIO `json:"inputs"`
		Outputs []AbiEndpointIO `json:"outputs"`
	} `json:"constructor"`
	Endpoints    []AbiEndpoint       `json:"endpoints"`
	Events       []AbiEvent          `json:"events"`
	HasCallbacks bool                `json:"hasCallbacks"`
	Types        map[string]*AbiType `json:"types"`
}

type AbiEndpoint struct {
	Name            string          `json:"name"`
	OnlyOwner       bool            `json:"onlyOwner"`
	Mutability      string          `json:"mutability"`
	PayableInTokens []string        `json:"payableInTokens"`
	Inputs          []AbiEndpointIO `json:"inputs"`
	Outputs         []AbiEndpointIO `json:"outputs"`
}

type AbiEndpointIO struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	MultiResult bool   `json:"multi_result"`
}

type AbiEvent struct {
	Identifier string          `json:"identifier"`
	Inputs     []AbiEventInput `json:"inputs"`
}

type AbiEventInput struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Indexed bool   `json:"indexed"`
}

type AbiType struct {
	Type     string           `json:"type"`
	Fields   []AbiTypeField   `json:"fields"`
	Variants []AbiTypeVariant `json:"variants"`
}

type AbiTypeField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type AbiTypeVariant struct {
	Name         string `json:"name"`
	Discriminant uint64 `json:"discriminant"`
}
