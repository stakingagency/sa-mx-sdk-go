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
			lines = append(lines, fmt.Sprintf("type %s struct {", abiTypeName))
			for i, field := range abiType.Fields {
				// capitalize first letter of struct fields to export them
				conv.abi.Types[abiTypeName].Fields[i].Name = utils.ToUpperFirstChar(field.Name)
			}
			for _, field := range abiType.Fields {
				goType, err := conv.abiType2goType(field.Type)
				if err != nil {
					return nil, err
				}

				lines = append(lines, fmt.Sprintf("    %s %s", field.Name, goType))
			}
			lines = append(lines, "}")
			lines = append(lines, "")

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

func (conv *AbiConverter) abiType2goType(abiType string) (string, error) {
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

	case "BigUint":
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

	if utils.IsMultiVariadic(abiType) {
		idx := strings.Index(abiType, "<")
		innerTypes := string([]byte(abiType)[idx+1:])
		innerTypes = strings.TrimSuffix(innerTypes, ">")
		idx = strings.Index(innerTypes, "<")
		innerTypes = string([]byte(innerTypes)[idx+1:])
		innerTypes = strings.TrimSuffix(innerTypes, ">")

		mapTypes := strings.Split(innerTypes, ",")
		n := len(mapTypes)
		if n < 2 {
			return "", errors.ErrUnknownAbiFieldType
		}

		idx = 1
		for {
			if strings.Contains(mapTypes[idx], ">") {
				mapTypes[idx-1] += "," + mapTypes[idx]
				mapTypes = append(mapTypes[:idx], mapTypes[idx+1:]...)
				idx = 0
				n--
			}
			idx++
			if idx >= n {
				break
			}
		}
		if n != 2 {
			return "", errors.ErrUnknownAbiFieldType
		}

		innerGoType1, err := conv.abiType2goType(mapTypes[0])
		if err != nil {
			return "", errors.ErrUnknownAbiFieldType
		}

		innerGoType2, err := conv.abiType2goType(mapTypes[1])
		if err != nil {
			return "", errors.ErrUnknownAbiFieldType
		}

		innerType := "map[" + innerGoType1 + "]" + innerGoType2

		return innerType, nil
	}

	if utils.IsTuple(abiType) {
		idx := strings.Index(abiType, "<")
		innerTypes := string([]byte(abiType)[idx+1:])
		innerTypes = strings.TrimSuffix(innerTypes, ">")

		tupleMembers := strings.Split(innerTypes, ",")
		complexType := make(map[string]string)
		for i, tupleMember := range tupleMembers {
			innerGoType, err := conv.abiType2goType(tupleMember)
			if err != nil {
				return "", err
			}

			variable := fmt.Sprintf("var%v", i)
			complexType[variable] = innerGoType
		}
		found := false
		typeName := fmt.Sprintf("ComplexType%v", len(conv.complexTypes))
		for existingName, existingType := range conv.complexTypes {
			if len(complexType) == len(existingType) {
				identical := true
				for n, t := range complexType {
					if existingType[n] != t {
						identical = false
					}
				}
				if identical {
					found = true
					typeName = existingName
				}
			}
		}
		if !found {
			conv.complexTypes[typeName] = complexType
		}

		return typeName, nil
	}

	return "", errors.ErrUnknownAbiFieldType
}

func (conv *AbiConverter) appendContractType(lines *[]string) {
	const (
		contractType = "type %s struct {\n" +
			"    netMan *network.NetworkManager\n" +
			"    contractAddress string\n" +
			"}\n" +
			"\n" +
			"func New%s(contractAddress string, proxyAddress string) (*%s, error) {\n" +
			"    netMan, err := network.NewNetworkManager(proxyAddress, \"\")\n" +
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
