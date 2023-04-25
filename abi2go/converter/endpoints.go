package converter

import (
	"fmt"
	"strings"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/data"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/errors"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/utils"
)

var (
	defaultTypeValues = map[string]string{
		"TokenIdentifier": "\"\"",
		"uint32":          "0",
		"uint64":          "0",
	}
)

func (conv *AbiConverter) convertReadonlyEndpoints() ([]string, error) {
	lines := make([]string, 0)

	for _, endpoint := range conv.abi.Endpoints {
		if endpoint.Mutability != "readonly" {
			continue
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
		body, err := conv.generateBody(endpoint)
		if err != nil {
			return nil, err
		}

		lines = append(lines, body...)
	}

	return lines, nil
}

func (conv *AbiConverter) generateImports() ([]string, error) {
	lines := make([]string, 0)
	name := utils.ToLowerFirstChar(conv.abi.Name)
	lines = append(lines, fmt.Sprintf("package %s", name))
	lines = append(lines, "")
	if len(conv.imports) > 0 {
		lines = append(lines, "import (")
		for imp := range conv.imports {
			lines = append(lines, fmt.Sprintf("    \"%s\"", imp))
		}
		lines = append(lines, ")")
		lines = append(lines, "")
	}
	if len(conv.customTypes) > 0 {
		for customType := range conv.customTypes {
			switch customType {
			case "Address":
				lines = append(lines, "type Address []byte")
			case "TokenIdentifier":
				lines = append(lines, "type TokenIdentifier string")
			}
			lines = append(lines, "")
		}
	}

	return lines, nil
}

func (conv *AbiConverter) generateBody(endpoint data.AbiEndpoint) ([]string, error) {
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

	lines = append(lines, line+errReturn+"err")
	lines = append(lines, "    }")
	lines = append(lines, "")

	// set output values
	for i, output := range endpoint.Outputs {
		if output.MultiResult || utils.IsVariadic(output.Type) {
			setVariadic, err := conv.setVariadicOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setVariadic...)
		} else if utils.IsList(output.Type) {
			setList, err := conv.setListOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setList...)
		} else {
			setSimpleValue, err := conv.setSimpleOutput(i, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, setSimpleValue)
		}
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

func (conv *AbiConverter) setVariadicOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	goType, _ := conv.abiType2goType(output.Type)
	goType = strings.TrimPrefix(goType, "[]")
	lines = append(lines, fmt.Sprintf("    res%v := make([]%s, 0)", i, goType))
	lines = append(lines, "    for i := 0; i < len(res.Data.ReturnData); i++ {")
	line := fmt.Sprintf("        res%v = append(res%v, ", i, i)
	switch goType {
	case "*big.Int":
		line += "big.NewInt(0).SetBytes(res.Data.ReturnData[i])"
	default:
		return nil, errors.ErrNotImplemented
	}
	line += ")"
	lines = append(lines, line)
	lines = append(lines, "    }")

	return lines, nil
}

func (conv *AbiConverter) setListOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	goType, _ := conv.abiType2goType(output.Type)
	goType = strings.TrimPrefix(goType, "[]")
	lines = append(lines, fmt.Sprintf("    res%v := make([]%s, 0)", i, goType))
	lines = append(lines, "    idx := 0")
	lines = append(lines, "    allOk, ok := true, true")
	switch goType {
	default:
		if conv.abi.Types[goType] != nil {
			for _, field := range conv.abi.Types[goType].Fields {
				fieldType, _ := conv.abiType2goType(field.Type)
				lines = append(lines, fmt.Sprintf("    var %s %s", field.Name, fieldType))
			}
		}
	}
	lines = append(lines, "    for {")
	switch goType {
	default:
		if conv.abi.Types[goType] != nil {
			for _, field := range conv.abi.Types[goType].Fields {
				conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
				line := fmt.Sprintf("        %s, idx, ok = utils.Parse", field.Name)
				fieldGoType, _ := conv.abiType2goType(field.Type)
				switch fieldGoType {
				case "*big.Int":
					line += "BigInt"

				case "uint64":
					line += "Uint64"

				case "uint32":
					line += "Uint32"

				default:
					return nil, errors.ErrNotImplemented
				}
				line += "(res.Data.ReturnData[0], idx)"
				lines = append(lines, line)
				lines = append(lines, "        allOk = allOk && ok")
			}
			lines = append(lines, "        if !allOk {")
			lines = append(lines, "            break")
			lines = append(lines, "        }")
			lines = append(lines, fmt.Sprintf("        item := %s{", goType))
			for _, field := range conv.abi.Types[goType].Fields {
				lines = append(lines, fmt.Sprintf("            %s: %s,", field.Name, field.Name))
			}
			lines = append(lines, "        }")
			lines = append(lines, fmt.Sprintf("        res%v = append(res%v, item)", i, i))
		} else {
			return nil, errors.ErrNotImplemented
		}
	}
	lines = append(lines, "    }")

	return lines, nil
}

func (conv *AbiConverter) setSimpleOutput(i int, output data.AbiEndpointIO) (string, error) {
	goType, _ := conv.abiType2goType(output.Type)
	goType = strings.TrimPrefix(goType, "[]")
	line := fmt.Sprintf("    res%v := ", i)
	switch goType {
	case "uint64":
		conv.imports["encoding/binary"] = true
		line += fmt.Sprintf("binary.BigEndian.Uint64(res.Data.ReturnData[%v])", i)

	case "uint32":
		conv.imports["encoding/binary"] = true
		line += fmt.Sprintf("binary.BigEndian.Uint32(res.Data.ReturnData[%v])", i)

	case "*big.Int":
		line += fmt.Sprintf("big.NewInt(0).SetBytes(res.Data.ReturnData[%v])", i)

	case "TokenIdentifier":
		line += fmt.Sprintf("TokenIdentifier(res.Data.ReturnData[%v])", i)

	case "Address":
		line += fmt.Sprintf("res.Data.ReturnData[%v]", i)

	default:
		if conv.abi.Types[output.Type] != nil && conv.abi.Types[output.Type].Type == "enum" {
			line += fmt.Sprintf("%s(big.NewInt(0).SetBytes(res.Data.ReturnData[%v]).Uint64())", output.Type, i)
		} else {
			return "", errors.ErrUnknownGoFieldType
		}
	}

	return line, nil
}

func (conv *AbiConverter) generateErrorReturn(outputs []data.AbiEndpointIO) (string, error) {
	line := ""
	for i := 0; i < len(outputs); i++ {
		output := outputs[i]
		goType, _ := conv.abiType2goType(output.Type)
		def := conv.getDefaultTypeValue(goType)
		if conv.abi.Types[output.Type] != nil && conv.abi.Types[output.Type].Type == "enum" {
			def = "0"
		}
		line += fmt.Sprintf("%s, ", def)
	}

	return line, nil
}

func (conv *AbiConverter) generateInputs(inputs []data.AbiEndpointIO) (string, error) {
	res := ""
	n := len(inputs)
	for i, input := range inputs {
		if input.Name == "" {
			return "", errors.ErrUnnamedInput
		}

		goType, err := conv.abiType2goType(input.Type)
		if err != nil {
			return "", err
		}

		res += fmt.Sprintf("%s %s", input.Name, goType)
		if i < n-1 {
			res += ", "
		}
	}

	return res, nil
}

func (conv *AbiConverter) generateOutputs(outputs []data.AbiEndpointIO) (string, error) {
	res := ""
	named, unnamed := 0, 0
	outputs = append(outputs, data.AbiEndpointIO{Type: "error"})
	n := len(outputs)
	if n > 1 {
		res += "("
	}
	for i, output := range outputs {
		goType, err := conv.abiType2goType(output.Type)
		if err != nil {
			return "", err
		}

		if output.Name != "" {
			res += fmt.Sprintf("%s ", output.Name)
			named++
		} else {
			unnamed++
		}
		res += goType
		if i < n-1 {
			res += ", "
		}
	}
	if n > 1 {
		res += ")"
	}

	if named > 0 && unnamed > 0 {
		return "", errors.ErrMixedNamedAndUnnamedOutputs
	}

	return res, nil
}

func (conv *AbiConverter) generateInputArg(name string, goType string) ([]string, error) {
	switch goType {
	case "uint64":
		conv.imports["encoding/binary"] = true
		conv.imports["encoding/hex"] = true

		return []string{
			"bytes := make([]byte, 8)",
			"binary.BigEndian.PutUint64(bytes, " + name + ")",
			"args = append(args, hex.EncodeToString(bytes))",
		}, nil

	case "uint32":
		conv.imports["encoding/binary"] = true
		conv.imports["encoding/hex"] = true

		return []string{
			"bytes := make([]byte, 4)",
			"binary.BigEndian.PutUint32(bytes, " + name + ")",
			"args = append(args, hex.EncodeToString(bytes))",
		}, nil

	case "Address":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString(" + name + "))"}, nil

	case "*big.Int":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString(" + name + ".Bytes())"}, nil

	case "TokenIdentifier":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte(" + name + ")))"}, nil
	}

	return nil, errors.ErrUnknownGoFieldType
}

func (conv *AbiConverter) getDefaultTypeValue(goType string) string {
	res := defaultTypeValues[goType]
	if res == "" {
		res = "nil"
	}

	return res
}
