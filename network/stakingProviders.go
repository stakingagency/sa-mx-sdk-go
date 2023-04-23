package network

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	sdkData "github.com/multiversx/mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func (nm *NetworkManager) GetAllProvidersAddresses() ([]string, error) {
	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	query := &sdkData.VmValueRequest{
		Address:  utils.DelegationManagerSC,
		FuncName: "getAllContractAddresses",
	}
	res, err := nm.proxy.ExecuteVMQuery(context.Background(), query)
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

func (nm *NetworkManager) GetMetaData(providerAddress string) (
	name string, website string, identity string, err error,
) {
	query := &sdkData.VmValueRequest{
		Address:  providerAddress,
		FuncName: "getMetaData",
	}
	var res *sdkData.VmValuesResponseData
	res, err = nm.proxy.ExecuteVMQuery(context.Background(), query)
	if err != nil {
		log.Error("can not get contract meta data", "error", err, "function", "GetMetaData")
		return
	}

	if len(res.Data.ReturnData) != 3 {
		err = errors.New("invalid response")
		return
	}

	name = utils.UTF8(string(res.Data.ReturnData[0]))
	website = string(res.Data.ReturnData[1])
	identity = string(res.Data.ReturnData[2])

	return
}

func (nm *NetworkManager) GetUserStakeInfo(address string, provider string) (
	stake *big.Int, reward *big.Int, undelegated *big.Int, unbondable *big.Int, err error,
) {
	conv, _ := pubkeyConverter.NewBech32PubkeyConverter(32, log)
	pubkey, _ := conv.Decode(address)
	sPubKey := hex.EncodeToString(pubkey)
	var stakeInfo []*big.Int
	stakeInfo, err = nm.QueryScMultiIntResult(provider, "getDelegatorFundsData", []string{sPubKey})
	if err != nil {
		return
	}

	if len(stakeInfo) != 4 {
		err = errors.New("invalid response")

		return
	}

	stake = stakeInfo[0]
	reward = stakeInfo[1]
	undelegated = stakeInfo[2]
	unbondable = stakeInfo[3]

	return
}
