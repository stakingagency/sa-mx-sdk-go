package salsaContract

import (
    "github.com/stakingagency/sa-mx-sdk-go/utils"
    "encoding/binary"
    "math/big"
    "github.com/stakingagency/sa-mx-sdk-go/network"
    "encoding/hex"
)

type TokenIdentifier string

type Address []byte

type State int

const (
    Inactive State = 0
    Active State = 1
)

type Undelegation struct {
    Amount *big.Int
    Unbond_epoch uint64
}

type EsdtTokenPayment struct {
    Token_identifier TokenIdentifier
    Token_nonce uint64
    Amount *big.Int
}

type SalsaContract struct {
    netMan *network.NetworkManager
    contractAddress string
}

func NewSalsaContract(contractAddress string, proxyAddress string) (*SalsaContract, error) {
    netMan, err := network.NewNetworkManager(proxyAddress, "")
    if err != nil {
        return nil, err
    }

    contract := &SalsaContract{
        netMan:          netMan,
        contractAddress: contractAddress,
    }

    return contract, nil
}

func (contract *SalsaContract) GetLiquidTokenId() (TokenIdentifier, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getLiquidTokenId", nil)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetLiquidTokenSupply() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getLiquidTokenSupply", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetState() (State, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getState", nil)
    if err != nil {
        return 0, err
    }

    res0 := State(big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64())

    return res0, nil
}

func (contract *SalsaContract) GetProviderAddress() (Address, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getProviderAddress", nil)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *SalsaContract) GetUserUndelegations(user Address) ([]Undelegation, error) {
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserUndelegations", args)
    if err != nil {
        return nil, err
    }

    res0 := make([]Undelegation, 0)
    idx := 0
    allOk, ok := true, true
    var _Amount *big.Int
    var _Unbond_epoch uint64
    for {
        _Amount, idx, ok = utils.ParseBigInt(res.Data.ReturnData[0], idx)
        allOk = allOk && ok
        _Unbond_epoch, idx, ok = utils.ParseUint64(res.Data.ReturnData[0], idx)
        allOk = allOk && ok
        if !allOk {
            break
        }
        item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, item)
    }

    return res0, nil
}

func (contract *SalsaContract) GetTotalEgldStaked() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTotalEgldStaked", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetUserWithdrawnEgld() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserWithdrawnEgld", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetEgldReserve() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getEgldReserve", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetAvailableEgldReserve() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getAvailableEgldReserve", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetReserveUndelegations() ([]Undelegation, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getReserveUndelegations", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]Undelegation, 0)
    idx := 0
    allOk, ok := true, true
    var _Amount *big.Int
    var _Unbond_epoch uint64
    for {
        _Amount, idx, ok = utils.ParseBigInt(res.Data.ReturnData[0], idx)
        allOk = allOk && ok
        _Unbond_epoch, idx, ok = utils.ParseUint64(res.Data.ReturnData[0], idx)
        allOk = allOk && ok
        if !allOk {
            break
        }
        item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, item)
    }

    return res0, nil
}

func (contract *SalsaContract) GetReserverID(user Address) (uint32, error) {
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getReserverID", args)
    if err != nil {
        return 0, err
    }

    res0 := binary.BigEndian.Uint32(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetUsersReserves() ([]*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUsersReserves", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]*big.Int, 0)
    for i := 0; i < len(res.Data.ReturnData); i++ {
        res0 = append(res0, big.NewInt(0).SetBytes(res.Data.ReturnData[i]))
    }

    return res0, nil
}

func (contract *SalsaContract) GetUserReserveByAddress(user Address) (*big.Int, error) {
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserReserveByAddress", args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetUndelegateNowFee() (uint64, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUndelegateNowFee", nil)
    if err != nil {
        return 0, err
    }

    res0 := binary.BigEndian.Uint64(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetTokenPrice() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTokenPrice", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

