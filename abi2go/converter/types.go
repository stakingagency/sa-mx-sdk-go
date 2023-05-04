package converter

import (
	"fmt"
	"strings"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/errors"
	"github.com/stakingagency/sa-mx-sdk-go/abi2go/utils"
)

func (conv *AbiConverter) convertTypes() ([]string, error) {
	lines := make([]string, 0)
	for abiTypeName, abiType := range conv.abi.Types {
		switch abiType.Type {
		case "struct":
			for i, field := range abiType.Fields {
				// capitalize first letter of struct fields to export them
				conv.abi.Types[abiTypeName].Fields[i].Name = utils.ToUpperFirstChar(field.Name)
			}
			complexType := make([][2]string, 0)
			for _, field := range abiType.Fields {
				goType, err := conv.abiType2goType(field.Type)
				if err != nil {
					return nil, err
				}

				complexType = append(complexType, [2]string{field.Name, goType})
			}
			conv.complexTypes[abiTypeName] = complexType

		case "enum":
			lines = append(lines, fmt.Sprintf("type %s int", abiTypeName))
			lines = append(lines, "")
			lines = append(lines, "const (")
			for _, variant := range abiType.Variants {
				lines = append(lines, fmt.Sprintf("    %s %s = %v", variant.Name, abiTypeName, variant.Discriminant))
			}
			lines = append(lines, ")")
			lines = append(lines, "")

		default:
			return nil, errors.ErrUnknownAbiFieldType
		}
	}
	conv.appendContractType(&lines)

	return lines, nil
}

func (conv *AbiConverter) abiType2goType(abiType string, multiResult ...bool) (string, error) {
	// TODO: what is Option ?
	if strings.HasPrefix(abiType, "Option<") {
		abiType = strings.TrimPrefix(abiType, "Option<")
		abiType = strings.TrimSuffix(abiType, ">")
	}
	if strings.HasPrefix(abiType, "optional<") {
		abiType = strings.TrimPrefix(abiType, "optional<")
		abiType = strings.TrimSuffix(abiType, ">")
	}

	switch abiType {
	case "bool":
		return abiType, nil

	case "u64":
		return "uint64", nil

	case "u32":
		return "uint32", nil

	case "u8":
		return "byte", nil

	case "BigUint":
		conv.imports["math/big"] = true

		return "*big.Int", nil

	case "BigInt":
		conv.imports["math/big"] = true

		return "*big.Int", nil

	case "TokenIdentifier":
		conv.customTypes[abiType] = true

		return abiType, nil

	case "EgldOrEsdtTokenIdentifier":
		conv.customTypes[abiType] = true

		return abiType, nil

	case "Address":
		conv.customTypes[abiType] = true

		return abiType, nil

	case "error":
		return abiType, nil

	case "bytes":
		return "string", nil
	}

	_, ok := conv.abi.Types[abiType]
	if ok {
		return abiType, nil
	}

	_, ok = conv.complexTypes[abiType]
	if ok {
		return abiType, nil
	}

	if utils.IsList(abiType) || utils.IsSimpleVariadic(abiType) {
		idx := strings.Index(abiType, "<")
		innerType := string([]byte(abiType)[idx+1:])
		innerType = strings.TrimSuffix(innerType, ">")
		innerGoType, err := conv.abiType2goType(innerType)
		if err != nil {
			return "", err
		}

		innerType = "[]" + innerGoType

		return innerType, nil
	}

	if utils.IsMultiVariadic(abiType, multiResult...) || utils.IsMulti(abiType) {
		idx := strings.Index(abiType, "<")
		innerTypes := string([]byte(abiType)[idx+1:])
		innerTypes = strings.TrimSuffix(innerTypes, ">")
		if utils.IsMultiVariadic(abiType) {
			idx = strings.Index(innerTypes, "<")
			innerTypes = string([]byte(innerTypes)[idx+1:])
			innerTypes = strings.TrimSuffix(innerTypes, ">")
		}

		mapTypes := utils.SplitTypes(innerTypes)
		for i := 0; i < len(mapTypes); i++ {
			mapType := mapTypes[i]
			if strings.Contains(mapType, "<") && strings.HasSuffix(mapType, ">") {
				idx = strings.Index(mapType, "<")
				mapType = string([]byte(mapType)[idx+1:])
				mapType = strings.TrimSuffix(mapType, ">")
				innerMapTypes := utils.SplitTypes(mapType)
				innerTypeName, err := conv.getOrCreateComplexType(innerMapTypes)
				if err != nil {
					return "", errors.ErrUnknownAbiFieldType
				}

				mapTypes[i] = innerTypeName
			}
		}
		if len(mapTypes) < 2 {
			return "[]" + mapTypes[0], nil
		}

		typeName, err := conv.getOrCreateComplexType(mapTypes)
		if err != nil {
			return "", err
		}

		return "[]" + typeName, nil
	}

	if utils.IsTuple(abiType) {
		idx := strings.Index(abiType, "<")
		innerTypes := string([]byte(abiType)[idx+1:])
		innerTypes = strings.TrimSuffix(innerTypes, ">")

		tupleMembers := utils.SplitTypes(innerTypes)
		typeName, err := conv.getOrCreateComplexType(tupleMembers)
		if err != nil {
			return "", err
		}

		return typeName, nil
	}

	return "", errors.ErrUnknownAbiFieldType
}

func (conv *AbiConverter) getOrCreateComplexType(tupleMembers []string) (string, error) {
	complexType := make([][2]string, 0)
	for i, tupleMember := range tupleMembers {
		innerGoType, err := conv.abiType2goType(tupleMember)
		if err != nil {
			return "", err
		}

		variable := fmt.Sprintf("Var%v", i)
		complexType = append(complexType, [2]string{variable, innerGoType})
	}
	for existingName, existingType := range conv.complexTypes {
		if len(complexType) == len(existingType) {
			identical := true
			for i, newType := range complexType {
				if existingType[i][0] != newType[0] || existingType[i][1] != newType[1] {
					identical = false
				}
			}
			if identical {
				return existingName, nil
			}
		}
	}
	typeName := fmt.Sprintf("ComplexType%v", len(conv.complexTypes))
	conv.complexTypes[typeName] = complexType

	return typeName, nil
}

func (conv *AbiConverter) appendContractType(lines *[]string) {
	const (
		contractType = "type %s struct {\n" +
			"    netMan *network.NetworkManager\n" +
			"    contractAddress string\n" +
			"}\n" +
			"\n" +
			"func New%s(contractAddress string, proxyAddress string, indexAddress string) (*%s, error) {\n" +
			"    netMan, err := network.NewNetworkManager(proxyAddress, indexAddress)\n" +
			"    if err != nil {\n" +
			"        return nil, err\n" +
			"    }\n" +
			"\n" +
			"    contract := &%s{\n" +
			"        netMan:          netMan,\n" +
			"        contractAddress: contractAddress,\n" +
			"    }\n" +
			"\n" +
			"    return contract, nil\n" +
			"}\n"
	)
	line := fmt.Sprintf(contractType, conv.abi.Name, conv.abi.Name, conv.abi.Name, conv.abi.Name)
	*lines = append(*lines, strings.Split(line, "\n")...)
	conv.imports["github.com/stakingagency/sa-mx-sdk-go/network"] = true
}
