package network

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func (nm *NetworkManager) SearchIndexer(index string, query interface{}, sort []interface{}) ([]*data.IndexerEntry, error) {
	bytes, err := utils.PostHTTP(fmt.Sprintf("%s/%s/_pit?keep_alive=2m", nm.indexAddress, index), "")
	if err != nil {
		log.Error("post http", "error", err, "function", "searchIndexer")
		return nil, err
	}

	pit := &data.IndexerPitResponse{}
	err = json.Unmarshal(bytes, pit)
	if err != nil {
		log.Error("unmarshal http response (post)", "error", err, "function", "searchIndexer")
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/_search", nm.indexAddress)
	body := &data.IndexerPitSearch{
		Size: 10000,
		PIT: data.IndexerPit{
			ID:        pit.ID,
			KeepAlive: "2m",
		},
		Query: query,
		Sort:  sort,
	}
	sBody, _ := json.Marshal(body)

	res := make([]*data.IndexerEntry, 0)
	pits := make([]string, 0)
	for {
		bytes, err := utils.GetHTTP(endpoint, string(sBody))
		if err != nil {
			log.Error("get http", "error", err, "endpoint", endpoint, "body", string(sBody), "function", "searchIndexer")
			return nil, err
		}

		list := &data.IndexerResult{}
		err = json.Unmarshal(bytes, list)
		if err != nil {
			return nil, err
		}

		if list.Shards.Failed > 0 {
			log.Error("indexer error", "endpoint", endpoint, "result", string(bytes), "function", "searchIndexer")
			return nil, utils.ErrFailedIndexerShard
		}

		pits = append(pits, list.PitId)
		res = append(res, list.Hits.Hits...)
		var lastHit *data.IndexerEntry = nil
		if len(list.Hits.Hits) > 0 {
			lastHit = list.Hits.Hits[len(list.Hits.Hits)-1]
		}
		if len(list.Hits.Hits) < 10000 || lastHit == nil {
			break
		} else {
			body = &data.IndexerPitSearch{
				Size: 10000,
				PIT: data.IndexerPit{
					ID:        pit.ID,
					KeepAlive: "2m",
				},
				Query:       query,
				Sort:        sort,
				SearchAfter: lastHit.Sort,
			}
			sBody, _ = json.Marshal(body)
		}
	}

	for _, id := range pits {
		p := &data.IndexerPitResponse{
			ID: id,
		}
		sPit, _ := json.Marshal(p)
		_, err := utils.DeleteHTTP(fmt.Sprintf("%s/_pit", nm.indexAddress), string(sPit))
		if err != nil {
			log.Warn("delete pit id", "error", err, "function", "searchIndexer")
		}
	}

	return res, nil
}

func (nm *NetworkManager) GetTxInfo(hash string) (*data.IndexerEntry, error) {
	query := make(map[string]map[string]string)
	query["match"] = make(map[string]string)
	query["match"]["_id"] = hash
	res, err := nm.SearchIndexer("transactions", query, nil)
	if err != nil {
		return nil, err
	}

	if len(res) != 1 {
		return nil, utils.ErrTxNotFound
	}

	return res[0], nil
}

func (nm *NetworkManager) GetTxLogs(hash string) ([]*data.IndexerEntry, error) {
	query := make(map[string]map[string][]map[string]map[string]interface{})
	query["bool"] = make(map[string][]map[string]map[string]interface{})
	query["bool"]["should"] = make([]map[string]map[string]interface{}, 2)
	query["bool"]["should"][0] = make(map[string]map[string]interface{})
	query["bool"]["should"][1] = make(map[string]map[string]interface{})
	query["bool"]["should"][0]["term"] = make(map[string]interface{})
	query["bool"]["should"][1]["term"] = make(map[string]interface{})
	query["bool"]["should"][0]["term"]["originalTxHash"] = hash
	query["bool"]["should"][1]["term"]["_id"] = hash
	res, err := nm.SearchIndexer("logs", query, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (nm *NetworkManager) GetTxOperations(hash string) ([]*data.IndexerEntry, error) {
	query := make(map[string]map[string]string)
	query["match"] = make(map[string]string)
	query["match"]["originalTxHash"] = hash
	res, err := nm.SearchIndexer("operations", query, nil)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (nm *NetworkManager) GetTxResult(hash string) error {
	start := time.Now().Unix()
	for {
		if time.Now().Unix()-start > 120 {
			return utils.ErrTimeout
		}

		time.Sleep(time.Second * 12)
		tx, err := nm.GetTxInfo(hash)
		if err != nil {
			return err
		}

		if tx.Source.Status == "pending" {
			continue
		}

		if tx.Source.Status == "fail" {
			message := "tx error"
			logErr, _ := nm.GetLogErrors(hash)
			if logErr != "" {
				message = logErr
			}
			return errors.New(message)
		}

		if tx.Source.HasOperations {
			ops, err := nm.GetTxOperations(hash)
			if err != nil {
				return err
			}

			pending := false
			for _, op := range ops {
				if op.Source.Status == "pending" {
					pending = true
				}
			}
			if pending {
				continue
			}
		}

		break
	}

	errText, _ := nm.GetLogErrors(hash)
	if errText != "" {
		return errors.New(errText)
	}

	log.Debug("tx sent successfully", "hash", hash)

	return nil
}

func (nm *NetworkManager) GetLogErrors(hash string) (string, error) {
	message := ""
	logs, err := nm.GetTxLogs(hash)
	if err != nil {
		return "", err
	}

	txErrors := make([]string, 0)
	for _, log := range logs {
		for _, event := range log.Source.Events {
			if event.Identifier == "signalError" || event.Identifier == "internalVMErrors" {
				eventData := utils.Base64Decode(event.Data)
				hexEvent := strings.ReplaceAll(eventData, "@", "")
				hb, err := hex.DecodeString(hexEvent)
				if err == nil {
					eventData = string(hb)
				}
				list := strings.Split(eventData, "\n")
				for i := 0; i < len(list); i++ {
					if strings.Contains(list[i], "runtime.go") {
						l := strings.Split(list[i], " ")
						list[i] = strings.Join(l[1:], " ")
					}
					if strings.Contains(list[i], "error signalled by smartcontract") {
						list = append(list[:i], list[i+1:]...)
						i--
					}
				}
				eventData = strings.Join(list, "\n")
				txErrors = append(txErrors, "`"+event.Identifier+":` "+eventData)
			}
		}
	}
	if len(txErrors) > 0 {
		message = strings.Join(txErrors, "\n")
	}

	return message, nil
}
