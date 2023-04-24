package onedex

import "github.com/stakingagency/sa-mx-sdk-go/data"

type LiquidityPool struct {
	ID            uint32
	Token1        *data.ESDT
	Token1Reserve float64
	Token2        *data.ESDT
	Token2Reserve float64
	LpToken       *data.ESDT
	LpTokenSupply float64
	Token1Price   float64
	Enabled       bool
	State         byte
}
