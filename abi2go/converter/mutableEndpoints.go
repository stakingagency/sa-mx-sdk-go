package converter

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/data"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/utils"
)

func (conv *AbiConverter) convertMutableEndpoints() ([]string, error) {
	lines := make([]string, 0)

	for _, endpoint := range conv.abi.Endpoints {
		if endpoint.Mutability != "mutable" {
			continue
		}

		if endpoint.OnlyOwner {
			lines = append(lines, "// only owner")
		}
		line := fmt.Sprintf("func (contract *%s) %s(", conv.abi.Name, utils.ToUpperFirstChar(endpoint.Name))
		inputs, err := conv.generateInputs(endpoint.Inputs)
		if err != nil {
			return nil, err
		}

		line += inputs + ") "
		outputs, err := conv.generateOutputs(endpoint.Outputs)
		if err != nil {
			return nil, err
		}

		line += outputs + " {"
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
	for _, input := range endpoint.Inputs {
		goType, _ := conv.abiType2goType(input.Type) // we don't care for err because it was checked in generateInputs
		inputArg, err := conv.generateInputArg(input.Name, goType)
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

	// generate endpoint fetching
	line := fmt.Sprintf("    res, err := contract.netMan.QuerySC(contract.contractAddress, \"%s\", ", endpoint.Name)
	if len(inputArgs) > 0 {
		line += "args)"
	} else {
		line += "nil)"
	}
	lines = append(lines, line)
	lines = append(lines, "    if err != nil {")
	line = "        return "
	errReturn, err := conv.generateErrorReturn(endpoint.Outputs)
	if err != nil {
		return nil, err
	}

	lines = append(lines, line+errReturn)
	lines = append(lines, "    }")
	lines = append(lines, "")

	// set output values
	for i, output := range endpoint.Outputs {
		if utils.IsMultiVariadic(output.Type) {
			setMultiVariadic, err := conv.setMultiVariadicOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setMultiVariadic...)
			continue
		}

		if utils.IsSimpleVariadic(output.Type) {
			setVariadic, err := conv.setVariadicOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setVariadic...)
			continue
		}

		if utils.IsList(output.Type) {
			setList, err := conv.setListOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setList...)
			continue
		}

		setSimpleValue, err := conv.setSimpleOutput(i, output)
		if err != nil {
			return nil, err
		}

		lines = append(lines, setSimpleValue...)
	}
	lines = append(lines, "")
	line = "    return "
	for i := 0; i < len(endpoint.Outputs); i++ {
		line += fmt.Sprintf("res%v, ", i)
	}
	line += "nil"
	lines = append(lines, line)
	lines = append(lines, "}")
	lines = append(lines, "")

	return lines, nil
}
