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
		"string":                    "\"\"",
		"TokenIdentifier":           "\"\"",
		"EgldOrEsdtTokenIdentifier": "\"\"",
		"uint32":                    "0",
		"uint64":                    "0",
		"bool":                      "false",
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
			if strings.Contains(imp, "\"") {
				lines = append(lines, fmt.Sprintf("    %s", imp))
			} else {
				lines = append(lines, fmt.Sprintf("    \"%s\"", imp))
			}
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
			case "EgldOrEsdtTokenIdentifier":
				lines = append(lines, "type EgldOrEsdtTokenIdentifier string")
			case "EsdtLocalRole":
				lines = append(lines, "type EsdtLocalRole byte")
			}
			lines = append(lines, "")
		}
	}

	if len(conv.complexTypes) > 0 {
		for name, fields := range conv.complexTypes {
			lines = append(lines, fmt.Sprintf("type %s struct {", name))
			for _, field := range fields {
				fieldName := field[0]
				fieldType := field[1]
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
		oLines, err := conv.setOutput(i, output)
		if err != nil {
			return nil, err
		}

		lines = append(lines, oLines...)
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

func (conv *AbiConverter) setOutput(i int, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	goType, _ := conv.abiType2goType(output.Type)
	isArray := strings.HasPrefix(goType, "[]")
	goType = strings.TrimPrefix(goType, "[]")
	complexType, isComplexType := conv.complexTypes[goType]

	varName := fmt.Sprintf("res%v", i)
	if isArray {
		lines = append(lines, fmt.Sprintf("    %s := make([]%s, 0)", varName, goType))
	}

	if utils.IsMultiVariadic(output.Type) || utils.IsMulti(output.Type) {
		if !isComplexType || !isArray {
			return nil, errors.ErrUnknownGoFieldType
		}

		noOfFields := len(complexType)
		lines = append(lines, fmt.Sprintf("    for i := 0; i < len(res.Data.ReturnData); i+=%v {", noOfFields))
		for fieldIdx := 0; fieldIdx < noOfFields; fieldIdx++ {
			typeName := complexType[fieldIdx][0]
			innerType := complexType[fieldIdx][1]
			dataSource := fmt.Sprintf("res.Data.ReturnData[i+%v]", fieldIdx)
			parsedLines, err := conv.parseAnyType(typeName, innerType, typeName, dataSource, "        ", output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, parsedLines...)
		}
		lines = append(lines, fmt.Sprintf("        inner := %s{", goType))
		for fieldIdx := 0; fieldIdx < noOfFields; fieldIdx++ {
			typeName := complexType[fieldIdx][0]
			lines = append(lines, fmt.Sprintf("            %s: %s,", typeName, typeName))
		}
		lines = append(lines, "        }")
		lines = append(lines, fmt.Sprintf("        %s = append(%s, inner)", varName, varName))
		lines = append(lines, "    }")

		return lines, nil
	}

	if utils.IsSimpleVariadic(output.Type) {
		if !isArray {
			return nil, errors.ErrUnknownGoFieldType
		}

		lines = append(lines, "    for i := 0; i < len(res.Data.ReturnData); i++ {")
		dataSource := "res.Data.ReturnData[i]"
		parsedLines, err := conv.parseAnyType(output.Name, goType, "_item", dataSource, "        ", output)
		if err != nil {
			return nil, err
		}

		lines = append(lines, parsedLines...)
		lines = append(lines, fmt.Sprintf("        %s = append(%s, _item)", varName, varName))
		lines = append(lines, "    }")

		return lines, nil
	}

	if utils.IsList(output.Type) {
		if !isArray || !isComplexType {
			return nil, errors.ErrUnknownGoFieldType
		}

		dataSource := "res.Data.ReturnData[0]"
		iLines, err := conv.instantiateParser(goType, true, "        ")
		if err != nil {
			return nil, err
		}

		lines = append(lines, iLines...)
		lines = append(lines, "    for {")

		cLines, err := conv.parseComplexType(output.Name, goType, "_item", dataSource, "        ", false, "break", output)
		if err != nil {
			return nil, err
		}

		lines = append(lines, cLines...)
		lines = append(lines, fmt.Sprintf("        %s = append(%s, _item)", varName, varName))
		lines = append(lines, "    }")

		return lines, nil
	}

	dataSource := fmt.Sprintf("res.Data.ReturnData[%v]", i)
	parsedLines, err := conv.parseAnyType(output.Name, goType, varName, dataSource, "    ", output)
	if err != nil {
		return nil, err
	}

	lines = append(lines, parsedLines...)

	return lines, nil
}

func (conv *AbiConverter) parseAnyType(typeName string, goType string, varName string, dataSource string, indent string, output data.AbiEndpointIO) ([]string, error) {
	lines := make([]string, 0)
	switch goType {
	case "uint64":
		lines = append(lines, fmt.Sprintf("%s%s := big.NewInt(0).SetBytes(%s).Uint64()", indent, varName, dataSource))

	case "uint32":
		lines = append(lines, fmt.Sprintf("%s%s := uint32(big.NewInt(0).SetBytes(%s).Uint64())", indent, varName, dataSource))

	case "*big.Int":
		lines = append(lines, fmt.Sprintf("%s%s := big.NewInt(0).SetBytes(%s)", indent, varName, dataSource))

	case "bool":
		lines = append(lines, fmt.Sprintf("%s%s := big.NewInt(0).SetBytes(%s).Uint64() == 1", indent, varName, dataSource))

	case "Address":
		lines = append(lines, fmt.Sprintf("%s%s := %s", indent, varName, dataSource))

	case "TokenIdentifier":
		lines = append(lines, fmt.Sprintf("%s%s := TokenIdentifier(%s)", indent, varName, dataSource))

	case "EgldOrEsdtTokenIdentifier":
		lines = append(lines, fmt.Sprintf("%s%s := EgldOrEsdtTokenIdentifier(%s)", indent, varName, dataSource))

	case "EsdtLocalRole":
		lines = append(lines, fmt.Sprintf("%s%s := EsdtLocalRole(byte(big.NewInt(0).SetBytes(%s).Uint64()))", indent, varName, dataSource))

	case "string":
		lines = append(lines, fmt.Sprintf("%s%s := string(%s)", indent, varName, dataSource))

	default:
		_, isComplexType := conv.complexTypes[goType]
		if isComplexType {
			iLines, err := conv.instantiateParser(goType, false, indent)
			if err != nil {
				return nil, err
			}

			lines = append(lines, iLines...)
			cLines, err := conv.parseComplexType(typeName, goType, varName, dataSource, indent, true, "return", output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, cLines...)

			return lines, nil
		}

		abiType, isAbiType := conv.abi.Types[goType]
		if isAbiType && abiType.Type == "enum" {
			lines = append(lines, fmt.Sprintf("%s%s := %s(big.NewInt(0).SetBytes(%s).Uint64())", indent, varName, goType, dataSource))

			return lines, nil
		}

		return nil, errors.ErrNotImplemented
	}

	return lines, nil
}

func (conv *AbiConverter) instantiateParser(goType string, allVars bool, indent string) ([]string, error) {
	conv.imports["github.com/stakingagency/sa-mx-sdk-go/utils"] = true
	lines := make([]string, 0)
	lines = append(lines, fmt.Sprintf("%sidx := 0", indent))
	lines = append(lines, fmt.Sprintf("%sok, allOk := true, true", indent))
	complexType, isComplexType := conv.complexTypes[goType]
	if allVars && isComplexType {
		for _, ct := range complexType {
			fieldName := ct[0]
			fieldType := ct[1]
			if fieldType == "TokenIdentifier" || fieldType == "EgldOrEsdtTokenIdentifier" {
				fieldType = "string"
			}
			lines = append(lines, fmt.Sprintf("%svar _%s %s", indent, fieldName, fieldType))
		}
	}

	return lines, nil
}

func (conv *AbiConverter) parseComplexType(typeName string, goType string, varName string, dataSource string, indent string, newVars bool, notAllOk string, output data.AbiEndpointIO) ([]string, error) {
	complexType, isComplexType := conv.complexTypes[goType]
	if !isComplexType {
		return nil, errors.ErrNotImplemented
	}

	attribute := ""
	if newVars {
		attribute = ":"
	}
	lines := make([]string, 0)
	for _, ct := range complexType {
		fieldName := ct[0]
		fieldType := ct[1]
		fieldLine := fmt.Sprintf("%s_%s, idx, ok %s= utils.Parse", indent, fieldName, attribute)
		parseType, err := conv.getParseType(fieldType)
		if err != nil {
			return nil, err
		}

		if parseType == "ComplexType" {
			ctLines, err := conv.parseComplexType(fieldName, fieldType, "_"+fieldName, dataSource, indent, newVars, notAllOk, output)
			if err != nil {
				return nil, err
			}

			lines = append(lines, ctLines...)
		} else {
			fieldLine += parseType + fmt.Sprintf("(%s, idx)", dataSource)
			lines = append(lines, fieldLine)
			lines = append(lines, fmt.Sprintf("%sallOk = allOk && ok", indent))
		}
	}
	lines = append(lines, fmt.Sprintf("%sif !allOk {", indent))
	switch notAllOk {
	case "continue":
		lines = append(lines, fmt.Sprintf("%s    continue", indent))
	case "break":
		lines = append(lines, fmt.Sprintf("%s    break", indent))
	case "return":
		errReturn, err := conv.generateErrorReturn([]data.AbiEndpointIO{output})
		if err != nil {
			return nil, err
		}

		conv.imports["errors"] = true
		lines = append(lines, indent+"    return "+errReturn+"ors.New(\"invalid response\")")
	}
	lines = append(lines, fmt.Sprintf("%s}", indent))
	lines = append(lines, fmt.Sprintf("%s%s := %s{", indent, varName, goType))
	for _, ct := range complexType {
		fieldName := ct[0]
		fieldType := ct[1]
		abiType, isAbiType := conv.abi.Types[fieldName]
		if (isAbiType && abiType.Type == "enum") || conv.customTypes[fieldType] {
			lines = append(lines, fmt.Sprintf("%s    %s: %s(_%s),", indent, fieldName, fieldType, fieldName))
			continue
		}
		lines = append(lines, fmt.Sprintf("%s    %s: _%s,", indent, fieldName, fieldName))
	}
	lines = append(lines, indent+"}")

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

func (conv *AbiConverter) generateInputArg(name string, goType string, i int) ([]string, error) {
	switch goType {
	case "uint64":
		conv.imports["encoding/binary"] = true
		conv.imports["encoding/hex"] = true

		return []string{
			fmt.Sprintf("bytes%v64 := make([]byte, 8)", i),
			fmt.Sprintf("binary.BigEndian.PutUint64(bytes%v64, "+name+")", i),
			fmt.Sprintf("args = append(args, hex.EncodeToString(bytes%v64))", i),
		}, nil

	case "uint32":
		conv.imports["encoding/binary"] = true
		conv.imports["encoding/hex"] = true

		return []string{
			fmt.Sprintf("bytes%v32 := make([]byte, 4)", i),
			fmt.Sprintf("binary.BigEndian.PutUint32(bytes%v32, "+name+")", i),
			fmt.Sprintf("args = append(args, hex.EncodeToString(bytes%v32))", i),
		}, nil

	case "Address":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString(" + name + "))"}, nil

	case "*big.Int":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString(" + name + ".Bytes()))"}, nil

	case "bool":
		return []string{"if " + name + " {args = append(args, \"01\") } else {args = append(args, \"00\")}"}, nil

	case "TokenIdentifier":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte(" + name + ")))"}, nil

	case "EgldOrEsdtTokenIdentifier":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte(" + name + ")))"}, nil

	case "EsdtLocalRole":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte{byte(" + name + ")}))"}, nil

	case "string":
		conv.imports["encoding/hex"] = true

		return []string{"args = append(args, hex.EncodeToString([]byte(" + name + ")))"}, nil
	}

	if strings.HasPrefix(goType, "[]") {
		lines := make([]string, 0)
		lines = append(lines, fmt.Sprintf("for _, elem := range %s {", name))
		args, err := conv.generateInputArg("elem", strings.TrimPrefix(goType, "[]"), i)
		if err != nil {
			return nil, err
		}

		for i, arg := range args {
			args[i] = "    " + arg
		}
		lines = append(lines, args...)
		lines = append(lines, "}")

		return lines, nil
	}

	complexType, isComplexType := conv.complexTypes[goType]
	if isComplexType {
		lines := make([]string, 0)
		for i, ct := range complexType {
			n := ct[0]
			t := ct[1]
			args, err := conv.generateInputArg(name+"."+n, t, i)
			if err != nil {
				return nil, err
			}

			for i, arg := range args {
				args[i] = "    " + arg
			}
			lines = append(lines, args...)
		}

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
	_, ok = conv.complexTypes[goType]
	if ok {
		res = goType + "{}"
	}
	if res == "" {
		res = "nil"
	}

	return res
}

func (conv *AbiConverter) getParseType(goType string) (string, error) {
	switch goType {
	case "uint64":
		return "Uint64", nil

	case "uint32":
		return "Uint32", nil

	case "*big.Int":
		return "BigInt", nil

	case "bool":
		return "Bool", nil

	case "Address":
		return "Pubkey", nil

	case "TokenIdentifier":
		return "String", nil

	case "EgldOrEsdtTokenIdentifier":
		return "String", nil

	case "EsdtLocalRole":
		return "Byte", nil

	case "string":
		return "String", nil

	default:
		if conv.abi.Types[goType] != nil && conv.abi.Types[goType].Type == "enum" {
			return "Byte", nil
		}

		_, isComplexType := conv.complexTypes[goType]
		if isComplexType {
			return "ComplexType", nil
		}
	}

	return "", errors.ErrNotImplemented
}
