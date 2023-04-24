package onedex

type Launchpad struct {
	ID          uint32
	IsLive      bool
	CreateTime  int64
	StartTime   int64
	EndTime     int64
	Description string
	Telegram    string
	Twitter     string
	Website     string
	HardCap     float64
	TotalBought float64
	FundAmount  float64
	Token       string
	FundToken   string
	Rate        float64
	Buyers      []*LaunchpadBuyer
}

type LaunchpadBuyer struct {
	Address string
	Amount  float64
}
