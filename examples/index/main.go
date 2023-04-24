package main

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/network"
)

const (
	proxyAddress = "https://gateway.multiversx.com"
	indexAddress = "https://index.multiversx.com"

	receiver = "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqhllllsajxzat"
	function = "delegate"
)

func main() {
	netMan, err := network.NewNetworkManager(proxyAddress, indexAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	query := make(map[string]map[string][]map[string]map[string]interface{})
	query["bool"] = make(map[string][]map[string]map[string]interface{})
	query["bool"]["filter"] = make([]map[string]map[string]interface{}, 2)
	query["bool"]["filter"][0] = make(map[string]map[string]interface{})
	query["bool"]["filter"][0]["term"] = make(map[string]interface{})
	query["bool"]["filter"][1] = make(map[string]map[string]interface{})
	query["bool"]["filter"][1]["term"] = make(map[string]interface{})

	query["bool"]["filter"][0]["term"]["receiver"] = receiver
	query["bool"]["filter"][1]["term"]["function"] = function

	sort := make([]interface{}, 0)
	sortDoc := make(map[string]string)
	sortDoc["_shard_doc"] = "desc"
	sort = append(sort, sortDoc)

	txs, err := netMan.SearchIndexer("transactions", query, sort)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v %s transactions sent to %s\n", len(txs), function, receiver)
}
