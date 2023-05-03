package salsaContract

import (
    "math/big"
    "github.com/stakingagency/sa-mx-sdk-go/network"
    "encoding/hex"
    "github.com/stakingagency/sa-mx-sdk-go/utils"
    "github.com/stakingagency/sa-mx-sdk-go/data"
    "strings"
    "encoding/binary"
)

type TokenIdentifier string

type Address []byte

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

func (contract *SalsaContract) GetUserUndelegations(user Address) ([]Undelegation, error) {
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(user))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUserUndelegations", args)
    if err != nil {
        return nil, err
    }

    res0 := make([]Undelegation, 0)
        idx := 0
        ok, allOk := true, true
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
        ok, allOk := true, true
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
        _item := Undelegation{
            Amount: _Amount,
            Unbond_epoch: _Unbond_epoch,
        }
        res0 = append(res0, _item)
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

    res0 := uint32(big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64())

    return res0, nil
}

func (contract *SalsaContract) GetUsersReserves() ([]*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUsersReserves", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]*big.Int, 0)
    for i := 0; i < len(res.Data.ReturnData); i++ {
        _item := big.NewInt(0).SetBytes(res.Data.ReturnData[i])
        res0 = append(res0, _item)
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

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

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
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(amount.Bytes()))
    dataField := "removeReserve" + "@" + strings.Join(args, "@")
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

func (contract *SalsaContract) UnDelegateNow(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := hex.EncodeToString([]byte("unDelegateNow"))
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

func (contract *SalsaContract) UndelegateReserves(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := "undelegateReserves"
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

// only owner
func (contract *SalsaContract) RegisterLiquidToken(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, token_display_name string, token_ticker string, num_decimals uint32) error {
    args := make([]string, 0)
    args = append(args, hex.EncodeToString([]byte(token_display_name)))
    args = append(args, hex.EncodeToString([]byte(token_ticker)))
    bytes232 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes232, num_decimals)
    args = append(args, hex.EncodeToString(bytes232))
    dataField := "registerLiquidToken" + "@" + strings.Join(args, "@")
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
    args := make([]string, 0)
    args = append(args, hex.EncodeToString(address))
    dataField := "setProviderAddress" + "@" + strings.Join(args, "@")
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
    args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, new_fee)
    args = append(args, hex.EncodeToString(bytes064))
    dataField := "setUndelegateNowFee" + "@" + strings.Join(args, "@")
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

