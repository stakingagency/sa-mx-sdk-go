package staking

import (
	"context"
	"encoding/hex"
	"math/big"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type (
	ProviderOwnerChangedCallbackFunc   func(providerAddress string, oldOwner string, newOwner string)
	ProviderNameChangedCallbackFunc    func(providerAddress string, oldName string, newName string)
	ProviderFeeChangedCallbackFunc     func(providerAddress string, oldFee float64, newFee float64)
	ProviderCapChangedCallbackFunc     func(providerAddress string, oldCap float64, newCap float64)
	ProviderSpaceAvailableCallbackFunc func(providerAddress string, spaceAvailable float64)
	NewProviderCallbackFunc            func(providerAddress string)
	ProviderClosedCallbackFunc         func(providerAddress string)
)

type Staking struct {
	netMan          *network.NetworkManager
	refreshInterval time.Duration

	cachedProviders    map[string]*data.StakingProvider
	cachedProvidersMut sync.Mutex

	providerOwnerChangedCallback   ProviderOwnerChangedCallbackFunc
	providerNameChangedCallback    ProviderNameChangedCallbackFunc
	providerFeeChangedCallback     ProviderFeeChangedCallbackFunc
	providerCapChangedCallback     ProviderCapChangedCallbackFunc
	providerSpaceAvailableCallback ProviderSpaceAvailableCallbackFunc
	newProviderCallback            NewProviderCallbackFunc
	providerClosedCallback         ProviderClosedCallbackFunc
}

var log = logger.GetOrCreate("staking")

func NewStaking(netMan *network.NetworkManager, refreshInterval time.Duration) (*Staking, error) {
	st := &Staking{
		netMan:          netMan,
		refreshInterval: refreshInterval,

		cachedProviders: make(map[string]*data.StakingProvider),

		providerOwnerChangedCallback:   nil,
		providerNameChangedCallback:    nil,
		providerFeeChangedCallback:     nil,
		providerCapChangedCallback:     nil,
		providerSpaceAvailableCallback: nil,
		newProviderCallback:            nil,
		providerClosedCallback:         nil,
	}
	st.startTasks()

	return st, nil
}

func (st *Staking) SetProviderOwnerChangedCallback(f ProviderOwnerChangedCallbackFunc) {
	st.providerOwnerChangedCallback = f
}

func (st *Staking) SetProviderNameChangedCallback(f ProviderNameChangedCallbackFunc) {
	st.providerNameChangedCallback = f
}

func (st *Staking) SetProviderFeeChangedCallback(f ProviderFeeChangedCallbackFunc) {
	st.providerFeeChangedCallback = f
}

func (st *Staking) SetProviderCapChangedCallback(f ProviderCapChangedCallbackFunc) {
	st.providerCapChangedCallback = f
}

func (st *Staking) SetNewProviderCallback(f NewProviderCallbackFunc) {
	st.newProviderCallback = f
}

func (st *Staking) SetProviderClosedCallback(f ProviderClosedCallbackFunc) {
	st.providerClosedCallback = f
}

func (st *Staking) SetProviderSpaceAvailableCallback(f ProviderSpaceAvailableCallbackFunc) {
	st.providerSpaceAvailableCallback = f
}

func (st *Staking) GetAllProvidersAddresses() ([]string, error) {
	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	query := &sdkData.VmValueRequest{
		Address:  utils.DelegationManagerSC,
		FuncName: "getAllContractAddresses",
	}
	res, err := st.netMan.GetProxy().ExecuteVMQuery(context.Background(), query)
	if err != nil {
		log.Error("can not get contract info", "error", err, "function", "GetAllContracts")
		return nil, err
	}

	addresses := make([]string, 0)
	for _, pubKey := range res.Data.ReturnData {
		address := converter.Encode(pubKey)
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (st *Staking) GetCachedMetaData(providerAddress string) (
	name string, website string, identity string, err error,
) {
	if st.refreshInterval == utils.NoRefresh {
		return "", "", "", utils.ErrRefreshIntervalNotSet
	}

	st.cachedProvidersMut.Lock()
	cfg := st.cachedProviders[providerAddress]
	st.cachedProvidersMut.Unlock()
	if cfg == nil {
		name, website, identity, err = st.GetMetaData(providerAddress)
		if err != nil {
			return "", "", "", err
		}

		return
	}

	return cfg.Name, cfg.Website, cfg.Identity, nil
}

func (st *Staking) GetMetaData(providerAddress string) (
	name string, website string, identity string, err error,
) {
	query := &sdkData.VmValueRequest{
		Address:  providerAddress,
		FuncName: "getMetaData",
	}
	var res *sdkData.VmValuesResponseData
	res, err = st.netMan.GetProxy().ExecuteVMQuery(context.Background(), query)
	if err != nil {
		log.Error("can not get contract meta data", "error", err, "function", "GetMetaData")
		return
	}

	if len(res.Data.ReturnData) != 3 {
		err = utils.ErrInvalidResponse
		return
	}

	name = utils.UTF8(string(res.Data.ReturnData[0]))
	website = string(res.Data.ReturnData[1])
	identity = string(res.Data.ReturnData[2])

	return
}

func (st *Staking) GetUserStakeInfo(address string, providerAddress string) (
	stake *big.Int, reward *big.Int, undelegated *big.Int, unbondable *big.Int, err error,
) {
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	pubkey, _ := conv.Decode(address)
	sPubKey := hex.EncodeToString(pubkey)
	var stakeInfo []*big.Int
	stakeInfo, err = st.netMan.QueryScMultiIntResult(providerAddress, "getDelegatorFundsData", []string{sPubKey})
	if err != nil {
		log.Error("query vm", "error", err, "function", "GetUserStakeInfo")
		return
	}

	if len(stakeInfo) != 4 {
		err = utils.ErrInvalidResponse

		return
	}

	stake = stakeInfo[0]
	reward = stakeInfo[1]
	undelegated = stakeInfo[2]
	unbondable = stakeInfo[3]

	return
}

func (st *Staking) GetCachedProviderConfig(providerAddress string) (*data.StakingProvider, error) {
	if st.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	st.cachedProvidersMut.Lock()
	cfg := st.cachedProviders[providerAddress]
	st.cachedProvidersMut.Unlock()
	if cfg == nil {
		var err error
		cfg, err = st.GetProviderConfig(providerAddress)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (st *Staking) GetProviderConfig(providerAddress string) (*data.StakingProvider, error) {
	res, err := st.netMan.QueryScMultiIntResult(providerAddress, "getContractConfig", nil)
	if err != nil {
		return nil, err
	}

	if len(res) != 10 {
		return nil, utils.ErrInvalidResponse
	}

	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	owner := conv.Encode(res[0].Bytes())
	cfg := &data.StakingProvider{
		ContractAddress:  providerAddress,
		Owner:            owner,
		ServiceFee:       utils.Denominate(res[1], 18) / 100,
		MaxDelegationCap: utils.Denominate(res[2], 18),
		HasDelegationCap: res[5].String() == "true",
	}
	cfg.Name, cfg.Website, cfg.Identity, err = st.GetMetaData(providerAddress)
	if err != nil {
		return nil, err
	}

	iActiveStake, err := st.netMan.QueryScIntResult(providerAddress, "getTotalActiveStake", nil)
	if err != nil {
		return nil, err
	}

	cfg.ActiveStake = utils.Denominate(iActiveStake, 18)

	return cfg, nil
}

func (st *Staking) GetCachedProvidersConfigs() (map[string]*data.StakingProvider, error) {
	if st.refreshInterval == utils.NoRefresh {
		return nil, utils.ErrRefreshIntervalNotSet
	}

	res := make(map[string]*data.StakingProvider)
	st.cachedProvidersMut.Lock()
	for k, v := range st.cachedProviders {
		res[k] = v
	}
	st.cachedProvidersMut.Unlock()

	return res, nil
}

func (st *Staking) GetProvidersConfigs() (map[string]*data.StakingProvider, error) {
	addresses, err := st.GetAllProvidersAddresses()
	if err != nil {
		return nil, err
	}

	res := make(map[string]*data.StakingProvider, 0)
	for _, address := range addresses {
		cfg, err := st.GetProviderConfig(address)
		if err != nil {
			continue
		}

		res[address] = cfg
	}

	return res, nil
}
