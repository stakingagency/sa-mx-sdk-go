package data

type StakingProvider struct {
	ContractAddress  string
	Owner            string
	Name             string
	Website          string
	Identity         string
	ServiceFee       float64
	MaxDelegationCap float64
	HasDelegationCap bool
	ActiveStake      float64
}
