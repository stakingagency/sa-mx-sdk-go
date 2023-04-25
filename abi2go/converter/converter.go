package converter

import (
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/data"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/errors"
)

type AbiConverter struct {
	abi          *data.ABI
	imports      map[string]bool
	customTypes  map[string]bool
	complexTypes map[string]map[string]string
}

func NewAbiConverter(fileName string) (*AbiConverter, error) {
	converter := &AbiConverter{
		imports:      make(map[string]bool),
		customTypes:  make(map[string]bool),
		complexTypes: make(map[string]map[string]string),
	}
	abi, err := converter.loadAbiFile(fileName)
	if err != nil {
		return nil, err
	}

	if abi.BuildInfo.Framework.Name != "multiversx-sc" {
		return nil, errors.ErrNotMultiversX
	}

	converter.abi = abi

	return converter, nil
}

func (conv *AbiConverter) Convert() error {
	lines := make([]string, 0)
	typesLines, err := conv.convertTypes()
	if err != nil {
		return err
	}

	lines = append(lines, typesLines...)

	readonlyEndpointsLines, err := conv.convertReadonlyEndpoints()
	if err != nil {
		return err
	}

	lines = append(lines, readonlyEndpointsLines...)

	mutableEndpointsLines, err := conv.convertMutableEndpoints()
	if err != nil {
		return err
	}

	lines = append(lines, mutableEndpointsLines...)

	imports, err := conv.generateImports()
	if err != nil {
		return err
	}

	lines = append(imports, lines...)

	return conv.saveGoFile(conv.abi.Name, lines)
}
