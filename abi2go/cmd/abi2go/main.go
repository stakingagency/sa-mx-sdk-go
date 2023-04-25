package main

import (
	"fmt"

	"github.com/stakingagency/sa-mx-sdk-go/abi2go/converter"
)

const (
	abiFileName = "onedex-sc.abi.json"
)

func main() {
	conv, err := converter.NewAbiConverter(abiFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = conv.Convert()
	if err != nil {
		fmt.Println(err)
		return
	}
}
