package staking

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/network"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

type Staking struct {
	netMan *network.NetworkManager
}

var log = logger.GetOrCreate("staking")

func NewStaking(netMan *network.NetworkManager) (*Staking, error) {
	st := &Staking{
		netMan: netMan,
	}

	return st, nil
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

func (st *Staking) GetUserStakeInfo(address string, provider string) (
	stake *big.Int, reward *big.Int, undelegated *big.Int, unbondable *big.Int, err error,
) {
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	pubkey, _ := conv.Decode(address)
	sPubKey := hex.EncodeToString(pubkey)
	var stakeInfo []*big.Int
	stakeInfo, err = st.netMan.QueryScMultiIntResult(provider, "getDelegatorFundsData", []string{sPubKey})
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
