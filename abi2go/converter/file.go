package converter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/data"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/utils"
)

func (conv *AbiConverter) loadAbiFile(fileName string) (*data.ABI, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	abi := &data.ABI{}
	err = json.Unmarshal(bytes, abi)
	if err != nil {
		return nil, err
	}

	return abi, nil
}

func (conv *AbiConverter) saveGoFile(name string, lines []string) error {
	name = utils.ToLowerFirstChar(name)
	_ = os.Mkdir(name, os.ModePerm)
	f, err := os.Create(fmt.Sprintf("%s/%s.go", name, name))
	if err != nil {
		return err
	}

	for _, line := range lines {
		fmt.Fprintln(f, line)
	}

	return f.Close()
}
