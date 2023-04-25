package converter

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/data"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/utils"
)

func (conv *AbiConverter) convertMutableEndpoints() ([]string, error) {
	lines := make([]string, 0)
	conv.imports["github.com/stakingagency/sa-mx-sdk-go/data"] = true
	for _, endpoint := range conv.abi.Endpoints {
		if endpoint.Mutability != "mutable" {
			continue
		}

		if endpoint.OnlyOwner {
			lines = append(lines, "// only owner")
		}
		line := fmt.Sprintf("func (contract *%s) %s(", conv.abi.Name, utils.ToUpperFirstChar(endpoint.Name))

		defaultInputs := "_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64"
		if len(endpoint.Inputs) > 0 {
			defaultInputs += ", "
		}
		inputs, err := conv.generateInputs(endpoint.Inputs)
		if err != nil {
			return nil, err
		}

		line += defaultInputs + inputs + ") "

		line += "error {"
		lines = append(lines, line)
		body, err := conv.generateMutableBody(endpoint)
		if err != nil {
			return nil, err
		}

		lines = append(lines, body...)
	}

	return lines, nil
}

func (conv *AbiConverter) generateMutableBody(endpoint data.AbiEndpoint) ([]string, error) {
	// generate input arguments
	lines := make([]string, 0)
	inputArgs := make([]string, 0)
	for i, input := range endpoint.Inputs {
		goType, _ := conv.abiType2goType(input.Type) // we don't care for err because it was checked in generateInputs
		inputArg, err := conv.generateInputArg(input.Name, goType, i)
		if err != nil {
			return nil, err
		}

		inputArgs = append(inputArgs, inputArg...)
	}
	if len(inputArgs) > 0 {
		lines = append(lines, "    args := make([]string, 0)")
	}
	for _, arg := range inputArgs {
		lines = append(lines, "    "+arg)
	}

	// generate endpoint sending transaction
	isEsdtTx := len(endpoint.PayableInTokens) == 1 && endpoint.PayableInTokens[0] == "*"
	line := ""
	if isEsdtTx {
		line = fmt.Sprintf("    dataField := hex.EncodeToString([]byte(\"%s\"))", endpoint.Name)
	} else {
		line = fmt.Sprintf("    dataField := \"%s\"", endpoint.Name)
	}
	if len(inputArgs) > 0 {
		conv.imports["strings"] = true
		line += " + \"@\" + strings.Join(args, \"@\")"
	}
	lines = append(lines, line)

	// send transaction
	if isEsdtTx {
		lines = append(lines, "    hash, err := contract.netMan.SendEsdtTransaction(_pk, contract.contractAddress, _value, _gasLimit, _token, dataField, _nonce)")
	} else {
		lines = append(lines, "    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)")
	}

	lines = append(lines, "    if err != nil {")
	lines = append(lines, "        return err")
	lines = append(lines, "    }")
	lines = append(lines, "")

	// watch transaction
	lines = append(lines, "    err = contract.netMan.GetTxResult(hash)")
	lines = append(lines, "    if err != nil {")
	lines = append(lines, "        return err")
	lines = append(lines, "    }")
	lines = append(lines, "")
	lines = append(lines, "    return nil")
	lines = append(lines, "}")
	lines = append(lines, "")

	return lines, nil
}
