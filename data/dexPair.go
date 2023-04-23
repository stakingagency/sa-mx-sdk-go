package data

import "math/big"

type DexPair struct {
	ContractAddress string
	State           bool
	Token1          string
	Token2          string
	Balance1        *big.Int
	Balance2        *big.Int
	Fee             *big.Int
}
