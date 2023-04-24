package xexchange

import (
	"math/big"

	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type DexPair struct {
	ContractAddress string
	State           bool
	Token1          *data.ESDT
	Token2          *data.ESDT
	Balance1        *big.Int
	Balance2        *big.Int
	Fee             *big.Int
}

func (pair *DexPair) GetPrice() float64 {
	balance1 := float64(0)
	balance2 := float64(0)
	decimals1 := 18
	decimals2 := 18
	if pair.Token1 != nil {
		decimals1 = int(pair.Token1.Decimals)
	}
	if pair.Token2 != nil {
		decimals2 = int(pair.Token2.Decimals)
	}
	if pair.Balance1 != nil {
		balance1 = utils.Denominate(pair.Balance1, decimals1)
	}
	if pair.Balance2 != nil {
		balance2 = utils.Denominate(pair.Balance2, decimals2)
	}
	if balance1 == 0 {
		return 0
	}

	return balance2 / balance1
}
