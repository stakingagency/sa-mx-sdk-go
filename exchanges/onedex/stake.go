package onedex

import "github.com/stakingagency/sa-mx-sdk-go/data"

type Stake struct {
	ID         uint32
	Token      *data.ESDT
	TotalStake float64
	RewardPool float64
	APR        float64
	Stakers    []*Staker
}

type Staker struct {
	Address    string
	Amount     float64
	Reward     float64
	LastUpdate int64
}

type UserStake struct {
	Token  *data.ESDT
	Amount float64
	Reward float64
}

type BoostedStake struct {
	TotalStake float64
	APR        float64
	Stakers    []*Staker
}
