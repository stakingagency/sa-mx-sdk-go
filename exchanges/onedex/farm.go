package onedex

import "github.com/stakingagency/sa-mx-sdk-go/data"

type Farm struct {
	ID                 uint32
	LpToken            *data.ESDT
	RewardToken1       *data.ESDT
	RewardToken2       *data.ESDT // for dual farms - nil if simple farm
	RewardPool1        float64
	RewardPool2        float64
	AnnualRewardPerLP1 float64
	AnnualRewardPerLP2 float64
	TotalStake         float64
	Farmers            []*Farmer
}

type Farmer struct {
	Address    string
	Amount     float64
	LastUpdate int64
}

type UserFarm struct {
	Farm    *Farm
	Amount  float64
	Reward1 float64
	Reward2 float64
}

func (f *Farm) IsDual() bool {
	return f.RewardToken2 != nil
}
