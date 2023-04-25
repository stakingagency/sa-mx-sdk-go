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
	switch abiType {
	case "u64":
		return "uint64", nil

	case "u32":
		return "uint32", nil

	case "BigUint":
		conv.imports["math/big"] = true

		return "*big.Int", nil

	case "TokenIdentifier":
		conv.customTypes["TokenIdentifier"] = true

		return "TokenIdentifier", nil

	case "Address":
		conv.customTypes["Address"] = true

		return "Address", nil

	case "error":
		return "error", nil
	}

	_, ok := conv.abi.Types[abiType]
	if ok {
		return abiType, nil
	}

	if utils.IsList(abiType) || utils.IsVariadic(abiType) {
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
