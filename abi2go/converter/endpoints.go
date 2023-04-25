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
		"bool":            "false",
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

	if len(conv.complexTypes) > 0 {
		for name, fields := range conv.complexTypes {
			lines = append(lines, fmt.Sprintf("type %s struct {", name))
			for fieldName, fieldType := range fields {
				lines = append(lines, fmt.Sprintf("    %s %s", fieldName, fieldType))
			}
			lines = append(lines, "}")
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

func (conv *AbiConverter) setMultiVariadicOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	goType, _ := conv.abiType2goType(output.Type)
	if !strings.HasPrefix(goType, "map[") || !strings.Contains(goType, "]") {
		return nil, errors.ErrUnknownGoFieldType
	}

	innerType := strings.TrimPrefix(goType, "map[")
	innerType = strings.Split(innerType, "]")[0]
	outerType := strings.Split(goType, "]")[1]

	lines = append(lines, fmt.Sprintf("    res%v := make(%s)", i, goType))
	lines = append(lines, "    for i := 0; i < len(res.Data.ReturnData); i+=2 {")

	// generate inner object
	switch innerType {
	case "TokenIdentifier":
		lines = append(lines, fmt.Sprintf("        inner%v := TokenIdentifier(res.Data.ReturnData[i])", i))

	default:
		complexType, ok := conv.complexTypes[innerType]
		if ok {
			conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
			lines = append(lines, "        idx := 0")
			lines = append(lines, "        ok, allOk := true, true")
			for n, t := range complexType {
				fieldLine := fmt.Sprintf("        _%s, idx, ok := utils.Parse", n)
				switch t {
				case "uint64":
					fieldLine += "Uint64"

				case "uint32":
					fieldLine += "Uint32"

				case "*big.Int":
					fieldLine += "BigInt"

				case "TokenIdentifier":
					fieldLine += "String"

				default:
					return nil, errors.ErrNotImplemented
				}
				fieldLine += "(res.Data.ReturnData[i], idx)"
				lines = append(lines, fieldLine)
				lines = append(lines, "        allOk = allOk && ok")
			}
			lines = append(lines, "        if !allOk {")
			lines = append(lines, "            continue")
			lines = append(lines, "        }")
			lines = append(lines, fmt.Sprintf("        inner%v := %s{", i, innerType))
			for n, t := range complexType {
				abiType, ok := conv.abi.Types[t]
				if (ok && abiType.Type == "enum") || conv.customTypes[t] {
					lines = append(lines, fmt.Sprintf("            %s: %s(_%s),", n, t, n))
					continue
				}
				lines = append(lines, fmt.Sprintf("            %s: _%s,", n, n))
			}
			lines = append(lines, "        }")
		} else {
			return nil, errors.ErrNotImplemented
		}
	}

	// generate outer object
	switch outerType {
	case "uint64":
		lines = append(lines, fmt.Sprintf("        outer%v := big.NewInt(0).SetBytes(res.Data.ReturnData[i+1]).Uint64()", i))

	case "uint32":
		conv.imports["encoding/binary"] = true
		lines = append(lines, fmt.Sprintf("        outer%v := binary.BigEndian.Uint32(res.Data.ReturnData[i+1])", i))

	case "*big.Int":
		lines = append(lines, fmt.Sprintf("        outer%v := big.NewInt(0).SetBytes(res.Data.ReturnData[i+1])", i))

	case "TokenIdentifier":
		lines = append(lines, fmt.Sprintf("        outer%v := TokenIdentifier(res.Data.ReturnData[i+1])", i))

	default:
		return nil, errors.ErrNotImplemented
	}

	lines = append(lines, fmt.Sprintf("        res%v[inner%v] = outer%v", i, i, i))
	lines = append(lines, "    }")

	return lines, nil
}

func (conv *AbiConverter) setVariadicOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	goType, _ := conv.abiType2goType(output.Type)
	goType = strings.TrimPrefix(goType, "[]")
	lines = append(lines, fmt.Sprintf("    res%v := make([]%s, 0)", i, goType))
	lines = append(lines, "    for i := 0; i < len(res.Data.ReturnData); i++ {")
	switch goType {
	case "uint64":
		lines = append(lines, fmt.Sprintf("        res%v = append(res%v, big.NewInt(0).SetBytes(res.Data.ReturnData[i]).Uint64())", i, i))

	case "uint32":
		lines = append(lines, fmt.Sprintf("        res%v = append(res%v, uint32(big.NewInt(0).SetBytes(res.Data.ReturnData[i]).Uint64()))", i, i))

	case "TokenIdentifier":
		lines = append(lines, fmt.Sprintf("        res%v = append(res%v, string(res.Data.ReturnData[i]))", i, i))

	case "*big.Int":
		lines = append(lines, fmt.Sprintf("        res%v = append(res%v, big.NewInt(0).SetBytes(res.Data.ReturnData[i]))", i, i))

	default:
		abiType, ok := conv.abi.Types[goType]
		if ok {
			if abiType.Type == "enum" {
				lines = append(lines, fmt.Sprintf("        res%v = append(res%v, %s(big.NewInt(0).SetBytes(res.Data.ReturnData[i])))", i, i, goType))
			} else {
				conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
				lines = append(lines, "        idx := 0")
				lines = append(lines, "        ok, allOk := true, true")
				for _, field := range abiType.Fields {
					innerGoType, err := conv.abiType2goType(field.Type)
					if err != nil {
						return nil, err
					}

					line := fmt.Sprintf("        _%s, idx, ok := utils.Parse", field.Name)
					switch innerGoType {
					case "uint64":
						line += "Uint64"

					case "uint32":
						line += "Uint32"

					case "*big.Int":
						line += "BigInt"

					case "bool":
						line += "Bool"

					case "Address":
						line += "Pubkey"

					case "TokenIdentifier":
						line += "String"

					default:
						if conv.abi.Types[field.Type] != nil && conv.abi.Types[field.Type].Type == "enum" {
							line += "Byte"
						} else {
							return nil, errors.ErrNotImplemented
						}
					}
					line += "(res.Data.ReturnData[i], idx)"
					lines = append(lines, line)
					lines = append(lines, "        allOk = allOk && ok")
				}
				lines = append(lines, "        if !allOk {")
				lines = append(lines, "            continue")
				lines = append(lines, "        }")
				lines = append(lines, fmt.Sprintf("        item := %s{", goType))
				for _, field := range abiType.Fields {
					abiType, ok := conv.abi.Types[field.Type]
					if (ok && abiType.Type == "enum") || conv.customTypes[field.Type] {
						lines = append(lines, fmt.Sprintf("            %s: %s(_%s),", field.Name, field.Type, field.Name))
						continue
					}
					lines = append(lines, fmt.Sprintf("            %s: _%s,", field.Name, field.Name))
				}
				lines = append(lines, "        }")
				lines = append(lines, fmt.Sprintf("        res%v = append(res%v, item)", i, i))
			}
		} else {
			return nil, errors.ErrNotImplemented
		}
	}
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
			conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
			for _, field := range conv.abi.Types[goType].Fields {
				conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
				line := fmt.Sprintf("        _%s, idx, ok = utils.Parse", field.Name)
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
				abiType, ok := conv.abi.Types[field.Type]
				if (ok && abiType.Type == "enum") || conv.customTypes[field.Type] {
					lines = append(lines, fmt.Sprintf("            %s: %s(_%s),", field.Name, field.Type, field.Name))
					continue
				}
				lines = append(lines, fmt.Sprintf("            %s: _%s,", field.Name, field.Name))
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

func (conv *AbiConverter) setSimpleOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	goType, _ := conv.abiType2goType(output.Type)
	goType = strings.TrimPrefix(goType, "[]")
	lines := make([]string, 0)
	switch goType {
	case "bool":
		lines = append(lines, fmt.Sprintf("    res%v := big.NewInt(0).SetBytes(res.Data.ReturnData[%v]).Uint64() == 1", i, i))
	case "uint64":
		conv.imports["encoding/binary"] = true
		lines = append(lines, fmt.Sprintf("    res%v := binary.BigEndian.Uint64(res.Data.ReturnData[%v])", i, i))

	case "uint32":
		conv.imports["encoding/binary"] = true
		lines = append(lines, fmt.Sprintf("    res%v := binary.BigEndian.Uint32(res.Data.ReturnData[%v])", i, i))

	case "*big.Int":
		lines = append(lines, fmt.Sprintf("    res%v := big.NewInt(0).SetBytes(res.Data.ReturnData[%v])", i, i))

	case "TokenIdentifier":
		lines = append(lines, fmt.Sprintf("    res%v := TokenIdentifier(res.Data.ReturnData[%v])", i, i))

	case "Address":
		lines = append(lines, fmt.Sprintf("    res%v := res.Data.ReturnData[%v]", i, i))

	default:
		abiType, ok := conv.abi.Types[goType]
		if ok {
			if abiType.Type == "enum" {
				lines = append(lines, fmt.Sprintf("    res%v := %s(big.NewInt(0).SetBytes(res.Data.ReturnData[%v]).Uint64())", i, output.Type, i))
			} else {
				conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
				lines = append(lines, "        idx := 0")
				lines = append(lines, "        ok, allOk := true, true")
				for _, field := range abiType.Fields {
					innerGoType, err := conv.abiType2goType(field.Type)
					if err != nil {
						return nil, err
					}

					line := fmt.Sprintf("        _%s, idx, ok := utils.Parse", field.Name)
					switch innerGoType {
					case "uint64":
						line += "Uint64"

					case "uint32":
						line += "Uint32"

					case "*big.Int":
						line += "BigInt"

					case "bool":
						line += "Bool"

					case "Address":
						line += "Pubkey"

					case "TokenIdentifier":
						line += "String"

					default:
						if conv.abi.Types[field.Type] != nil && conv.abi.Types[field.Type].Type == "enum" {
							line += "Byte"
						} else {
							return nil, errors.ErrNotImplemented
						}
					}
					line += fmt.Sprintf("(res.Data.ReturnData[%v], idx)", i)
					lines = append(lines, line)
					lines = append(lines, "        allOk = allOk && ok")
				}
				lines = append(lines, "        if !allOk {")
				errReturn, err := conv.generateErrorReturn([]data.AbiEndpointIO{output})
				if err != nil {
					return nil, err
				}

				conv.imports["errors"] = true
				lines = append(lines, "            return "+errReturn+"ors.New(\"invalid response\")")
				lines = append(lines, "        }")
				lines = append(lines, fmt.Sprintf("        res%v := %s{", i, goType))
				for _, field := range abiType.Fields {
					abiType, ok := conv.abi.Types[field.Type]
					if (ok && abiType.Type == "enum") || conv.customTypes[field.Type] {
						lines = append(lines, fmt.Sprintf("            %s: %s(_%s),", field.Name, field.Type, field.Name))
						continue
					}
					lines = append(lines, fmt.Sprintf("            %s: _%s,", field.Name, field.Name))
				}
				lines = append(lines, "        }")
			}
		} else {
			return nil, errors.ErrUnknownGoFieldType
		}
	}

	return lines, nil
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

	return line + "err", nil
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

		return []string{"args = append(args, hex.EncodeToString(" + name + ".Bytes()))"}, nil

	case "TokenIdentifier":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte(" + name + ")))"}, nil
	}

	if strings.HasPrefix(goType, "[]") {
		lines := make([]string, 0)
		lines = append(lines, fmt.Sprintf("for _, elem := range %s {", name))
		args, err := conv.generateInputArg("elem", strings.TrimPrefix(goType, "[]"))
		if err != nil {
			return nil, err
		}

		lines = append(lines, args...)
		lines = append(lines, "}")

		return lines, nil
	}

	return nil, errors.ErrUnknownGoFieldType
}

func (conv *AbiConverter) getDefaultTypeValue(goType string) string {
	res := defaultTypeValues[goType]
	abiType, ok := conv.abi.Types[goType]
	if ok {
		if abiType.Type == "enum" {
			res = "0"
		} else {
			res = goType + "{}"
		}
	}
	if res == "" {
		res = "nil"
	}

	return res
}