package salsaContract

import (
    "errors"
    "github.com/stakingagency/sa-mx-sdk-go/data"
    "strings"
    "encoding/binary"
    "math/big"
    "github.com/stakingagency/sa-mx-sdk-go/network"
    "encoding/hex"
    "github.com/stakingagency/sa-mx-sdk-go/utils"
)

type Address []byte

type TokenIdentifier string

type EsdtTokenPayment struct {
    Token_identifier TokenIdentifier
    Token_nonce uint64
    Amount *big.Int
}

type Undelegation struct {
    Amount *big.Int
    Unbond_epoch uint64
}

type State int

const (
    Inactive State = 0
    Active State = 1
)

type SalsaContract struct {
    netMan *network.NetworkManager
    contractAddress string
}

func NewSalsaContract(contractAddress string, proxyAddress string, indexAddress string) (*SalsaContract, error) {
    netMan, err := network.NewNetworkManager(proxyAddress, indexAddress)
    if err != nil {
        return nil, err
    }

    contract := &SalsaContract{
        netMan:          netMan,
        contractAddress: contractAddress,
    }

    return contract, nil
}

func (contract *SalsaContract) GetNetworkManager() *network.NetworkManager {
  return contract.netMan
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

func (contract *SalsaContract) GetUnbondPeriod() (uint64, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUnbondPeriod", nil)
    if err != nil {
        return 0, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

    return res0, nil
}

func (contract *SalsaContract) GetUserUndelegations(user Address) ([]Undelegation, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserUndelegations", _args)
    if err != nil {
        return nil, err
    }

    res0 := make([]Undelegation, 0)
    for i := 0; i < len(res.Data.ReturnData); i++ {
        idx := 0
        ok, allOk := true, true
        _Amount, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Unbond_epoch, idx, ok := utils.ParseUint64(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        if !allOk {
            return nil, errors.New("invalid response")
        }

        _item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, _item)
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

func (contract *SalsaContract) GetTotalWithdrawnEgld() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTotalWithdrawnEgld", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetTotalUserUndelegations() ([]Undelegation, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTotalUserUndelegations", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]Undelegation, 0)
    for i := 0; i < len(res.Data.ReturnData); i++ {
        idx := 0
        ok, allOk := true, true
        _Amount, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Unbond_epoch, idx, ok := utils.ParseUint64(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        if !allOk {
            return nil, errors.New("invalid response")
        }

        _item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, _item)
    }

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

func (contract *SalsaContract) GetReservePoints() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getReservePoints", nil)
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
    for i := 0; i < len(res.Data.ReturnData); i++ {
        idx := 0
        ok, allOk := true, true
        _Amount, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Unbond_epoch, idx, ok := utils.ParseUint64(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        if !allOk {
            return nil, errors.New("invalid response")
        }

        _item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, _item)
    }

    return res0, nil
}

func (contract *SalsaContract) GetUsersReservePoints(user Address) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUsersReservePoints", _args)
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

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

    return res0, nil
}

func (contract *SalsaContract) GetReservePointsAmount(egld_amount *big.Int) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(egld_amount.Bytes()))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getReservePointsAmount", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetReserveEgldAmount(points_amount *big.Int) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(points_amount.Bytes()))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getReserveEgldAmount", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *SalsaContract) GetUserReserve(user Address) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserReserve", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

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

func (contract *SalsaContract) Delegate(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "delegate"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) UnDelegate(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := hex.EncodeToString([]byte("unDelegate"))
    hash, err := contract.netMan.SendEsdtTransaction(_pk, contract.contractAddress, _value, _gasLimit, _token, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) Withdraw(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "withdraw"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) AddReserve(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "addReserve"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) RemoveReserve(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, amount *big.Int) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(amount.Bytes()))
    dataField := "removeReserve" + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) UnDelegateNow(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, min_amount_out *big.Int) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(min_amount_out.Bytes()))
    dataField := hex.EncodeToString([]byte("unDelegateNow")) + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendEsdtTransaction(_pk, contract.contractAddress, _value, _gasLimit, _token, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) UnDelegateAll(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "unDelegateAll"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) Compound(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "compound"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) WithdrawAll(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "withdrawAll"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

func (contract *SalsaContract) ComputeWithdrawn(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "computeWithdrawn"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) RegisterLiquidToken(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, token_display_name string, token_ticker string, num_decimals uint32) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(token_display_name)))
    _args = append(_args, hex.EncodeToString([]byte(token_ticker)))
    bytes232 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes232, num_decimals)
    _args = append(_args, hex.EncodeToString(bytes232))
    dataField := "registerLiquidToken" + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) SetStateActive(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "setStateActive"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) SetStateInactive(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "setStateInactive"
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) SetProviderAddress(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, address Address) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(address))
    dataField := "setProviderAddress" + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) SetUnbondPeriod(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, period uint64) error {
    _args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, period)
    _args = append(_args, hex.EncodeToString(bytes064))
    dataField := "setUnbondPeriod" + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

// only owner
func (contract *SalsaContract) SetUndelegateNowFee(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, new_fee uint64) error {
    _args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, new_fee)
    _args = append(_args, hex.EncodeToString(bytes064))
    dataField := "setUndelegateNowFee" + "@" + strings.Join(_args, "@")
    hash, err := contract.netMan.SendTransaction(_pk, contract.contractAddress, _value, _gasLimit, dataField, _nonce)
    if err != nil {
        return err
    }

    err = contract.netMan.GetTxResult(hash)
    if err != nil {
        return err
    }

    return nil
}

