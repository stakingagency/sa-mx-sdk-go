package oneDex

import (
    "github.com/stakingagency/sa-mx-sdk-go/data"
    "strings"
    "math/big"
    "github.com/stakingagency/sa-mx-sdk-go/network"
    "github.com/stakingagency/sa-mx-sdk-go/utils"
    "errors"
    "encoding/binary"
    "encoding/hex"
)

type Address []byte

type TokenIdentifier string

type Pair struct {
    Pair_id uint32
    State State
    Enabled bool
    Owner Address
    First_token_id TokenIdentifier
    Second_token_id TokenIdentifier
    Lp_token_id TokenIdentifier
    Lp_token_decimal uint32
    First_token_reserve *big.Int
    Second_token_reserve *big.Int
    Lp_token_supply *big.Int
    Lp_token_roles_are_set bool
}

type ComplexType1 struct {
    Var0 TokenIdentifier
    Var1 TokenIdentifier
}

type ComplexType2 struct {
    Var0 ComplexType1
    Var1 uint32
}

type ComplexType3 struct {
    Var0 TokenIdentifier
    Var1 uint32
}

type State int

const (
    Inactive State = 0
    Active State = 1
    ActiveButNoSwap State = 2
)

type OneDex struct {
    netMan *network.NetworkManager
    contractAddress string
}

func NewOneDex(contractAddress string, proxyAddress string, indexAddress string) (*OneDex, error) {
    netMan, err := network.NewNetworkManager(proxyAddress, indexAddress)
    if err != nil {
        return nil, err
    }

    contract := &OneDex{
        netMan:          netMan,
        contractAddress: contractAddress,
    }

    return contract, nil
}

func (contract *OneDex) GetNetworkManager() *network.NetworkManager {
  return contract.netMan
}
func (contract *OneDex) GetWegldTokenId() (TokenIdentifier, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getWegldTokenId", nil)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetUsdcTokenId() (TokenIdentifier, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUsdcTokenId", nil)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetBusdTokenId() (TokenIdentifier, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getBusdTokenId", nil)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetUsdtTokenId() (TokenIdentifier, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUsdtTokenId", nil)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetTotalFeePercent() (uint64, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTotalFeePercent", nil)
    if err != nil {
        return 0, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

    return res0, nil
}

func (contract *OneDex) GetSpecialFeePercent() (uint64, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getSpecialFeePercent", nil)
    if err != nil {
        return 0, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

    return res0, nil
}

func (contract *OneDex) GetStakingRewardFeePercent() (uint64, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getStakingRewardFeePercent", nil)
    if err != nil {
        return 0, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64()

    return res0, nil
}

func (contract *OneDex) GetTreasuryAddreess() (Address, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getTreasuryAddreess", nil)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *OneDex) GetStakingRewardAddress() (Address, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getStakingRewardAddress", nil)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *OneDex) GetBurnerAddreess() (Address, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getBurnerAddreess", nil)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *OneDex) GetUnwrapAddreess() (Address, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getUnwrapAddreess", nil)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *OneDex) GetRegisteringCost() (*big.Int, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getRegisteringCost", nil)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPaused() (bool, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPaused", nil)
    if err != nil {
        return false, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64() == 1

    return res0, nil
}

func (contract *OneDex) GetPairIds() ([]ComplexType2, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairIds", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]ComplexType2, 0)
    for i := 0; i < len(res.Data.ReturnData); i+=2 {
        idx := 0
        ok, allOk := true, true
        _Var0, idx, ok := utils.ParseString(res.Data.ReturnData[i+0], idx)
        allOk = allOk && ok
        _Var1, idx, ok := utils.ParseString(res.Data.ReturnData[i+0], idx)
        allOk = allOk && ok
        if !allOk {
            return nil, errors.New("invalid response")
        }

        Var0 := ComplexType1{
            Var0: TokenIdentifier(_Var0),
            Var1: TokenIdentifier(_Var1),
        }
        Var1 := uint32(big.NewInt(0).SetBytes(res.Data.ReturnData[i+1]).Uint64())
        inner := ComplexType2{
            Var0: Var0,
            Var1: Var1,
        }
        res0 = append(res0, inner)
    }

    return res0, nil
}

func (contract *OneDex) GetLastPairId() (uint32, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getLastPairId", nil)
    if err != nil {
        return 0, err
    }

    res0 := uint32(big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64())

    return res0, nil
}

func (contract *OneDex) GetLpTokenPairIdMap() ([]ComplexType3, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getLpTokenPairIdMap", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]ComplexType3, 0)
    for i := 0; i < len(res.Data.ReturnData); i+=2 {
        Var0 := TokenIdentifier(res.Data.ReturnData[i+0])
        Var1 := uint32(big.NewInt(0).SetBytes(res.Data.ReturnData[i+1]).Uint64())
        inner := ComplexType3{
            Var0: Var0,
            Var1: Var1,
        }
        res0 = append(res0, inner)
    }

    return res0, nil
}

func (contract *OneDex) GetPairOwner(pair_id uint32) (Address, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairOwner", _args)
    if err != nil {
        return nil, err
    }

    res0 := res.Data.ReturnData[0]

    return res0, nil
}

func (contract *OneDex) GetPairState(pair_id uint32) (State, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairState", _args)
    if err != nil {
        return 0, err
    }

    res0 := State(big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64())

    return res0, nil
}

func (contract *OneDex) GetPairEnabled(pair_id uint32) (bool, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairEnabled", _args)
    if err != nil {
        return false, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0]).Uint64() == 1

    return res0, nil
}

func (contract *OneDex) GetPairFirstTokenId(pair_id uint32) (TokenIdentifier, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairFirstTokenId", _args)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPairSecondTokenId(pair_id uint32) (TokenIdentifier, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairSecondTokenId", _args)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPairFirstTokenReserve(pair_id uint32) (*big.Int, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairFirstTokenReserve", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPairSecondTokenReserve(pair_id uint32) (*big.Int, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairSecondTokenReserve", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPairLpTokenId(pair_id uint32) (TokenIdentifier, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairLpTokenId", _args)
    if err != nil {
        return "", err
    }

    res0 := TokenIdentifier(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetPairLpTokenTotalSupply(pair_id uint32) (*big.Int, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getPairLpTokenTotalSupply", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetEquivalent(token_in TokenIdentifier, token_out TokenIdentifier, amount_in *big.Int) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(token_in)))
    _args = append(_args, hex.EncodeToString([]byte(token_out)))
    _args = append(_args, hex.EncodeToString(amount_in.Bytes()))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getEquivalent", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetMultiPathAmountOut(amount_in_arg *big.Int, path_args []TokenIdentifier) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(amount_in_arg.Bytes()))
    for _, elem := range path_args {
        _args = append(_args, hex.EncodeToString([]byte(elem)))
    }
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getMultiPathAmountOut", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetAmountOut(token_in TokenIdentifier, token_out TokenIdentifier, amount_in *big.Int) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(token_in)))
    _args = append(_args, hex.EncodeToString([]byte(token_out)))
    _args = append(_args, hex.EncodeToString(amount_in.Bytes()))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getAmountOut", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetMultiPathAmountIn(amount_out_wanted *big.Int, path_args []TokenIdentifier) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(amount_out_wanted.Bytes()))
    for _, elem := range path_args {
        _args = append(_args, hex.EncodeToString([]byte(elem)))
    }
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getMultiPathAmountIn", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) GetAmountIn(token_in TokenIdentifier, token_wanted TokenIdentifier, amount_wanted *big.Int) (*big.Int, error) {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(token_in)))
    _args = append(_args, hex.EncodeToString([]byte(token_wanted)))
    _args = append(_args, hex.EncodeToString(amount_wanted.Bytes()))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "getAmountIn", _args)
    if err != nil {
        return nil, err
    }

    res0 := big.NewInt(0).SetBytes(res.Data.ReturnData[0])

    return res0, nil
}

func (contract *OneDex) ViewPairs() ([]Pair, error) {
    res, err := contract.netMan.QuerySC(contract.contractAddress, "viewPairs", nil)
    if err != nil {
        return nil, err
    }

    res0 := make([]Pair, 0)
    for i := 0; i < len(res.Data.ReturnData); i++ {
        idx := 0
        ok, allOk := true, true
        _Pair_id, idx, ok := utils.ParseUint32(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _State, idx, ok := utils.ParseByte(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Enabled, idx, ok := utils.ParseBool(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Owner, idx, ok := utils.ParsePubkey(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _First_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Second_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Lp_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Lp_token_decimal, idx, ok := utils.ParseUint32(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _First_token_reserve, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Second_token_reserve, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Lp_token_supply, idx, ok := utils.ParseBigInt(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        _Lp_token_roles_are_set, idx, ok := utils.ParseBool(res.Data.ReturnData[i], idx)
        allOk = allOk && ok
        if !allOk {
            return nil, errors.New("invalid response")
        }

        _item := Pair{
            Pair_id: _Pair_id,
            State: State(_State),
            Enabled: _Enabled,
            Owner: Address(_Owner),
            First_token_id: TokenIdentifier(_First_token_id),
            Second_token_id: TokenIdentifier(_Second_token_id),
            Lp_token_id: TokenIdentifier(_Lp_token_id),
            Lp_token_decimal: _Lp_token_decimal,
            First_token_reserve: _First_token_reserve,
            Second_token_reserve: _Second_token_reserve,
            Lp_token_supply: _Lp_token_supply,
            Lp_token_roles_are_set: _Lp_token_roles_are_set,
        }
        res0 = append(res0, _item)
    }

    return res0, nil
}

func (contract *OneDex) ViewPair(pair_id uint32) (Pair, error) {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    res, err := contract.netMan.QuerySC(contract.contractAddress, "viewPair", _args)
    if err != nil {
        return Pair{}, err
    }

    idx := 0
    ok, allOk := true, true
    _Pair_id, idx, ok := utils.ParseUint32(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _State, idx, ok := utils.ParseByte(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Enabled, idx, ok := utils.ParseBool(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Owner, idx, ok := utils.ParsePubkey(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _First_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Second_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Lp_token_id, idx, ok := utils.ParseString(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Lp_token_decimal, idx, ok := utils.ParseUint32(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _First_token_reserve, idx, ok := utils.ParseBigInt(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Second_token_reserve, idx, ok := utils.ParseBigInt(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Lp_token_supply, idx, ok := utils.ParseBigInt(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    _Lp_token_roles_are_set, idx, ok := utils.ParseBool(res.Data.ReturnData[0], idx)
    allOk = allOk && ok
    if !allOk {
        return Pair{}, errors.New("invalid response")
    }

    res0 := Pair{
        Pair_id: _Pair_id,
        State: State(_State),
        Enabled: _Enabled,
        Owner: Address(_Owner),
        First_token_id: TokenIdentifier(_First_token_id),
        Second_token_id: TokenIdentifier(_Second_token_id),
        Lp_token_id: TokenIdentifier(_Lp_token_id),
        Lp_token_decimal: _Lp_token_decimal,
        First_token_reserve: _First_token_reserve,
        Second_token_reserve: _Second_token_reserve,
        Lp_token_supply: _Lp_token_supply,
        Lp_token_roles_are_set: _Lp_token_roles_are_set,
    }

    return res0, nil
}

// only owner
func (contract *OneDex) SetConfig(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, wegld_token_id TokenIdentifier, usdc_token_id TokenIdentifier, busd_token_id TokenIdentifier, usdt_token_id TokenIdentifier, total_fee_percent uint64, special_fee_percent uint64, staking_reward_fee_percent uint64, treasury_address Address, staking_reward_address Address, burner_address Address, unwrap_address Address, registering_cost *big.Int) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(wegld_token_id)))
    _args = append(_args, hex.EncodeToString([]byte(usdc_token_id)))
    _args = append(_args, hex.EncodeToString([]byte(busd_token_id)))
    _args = append(_args, hex.EncodeToString([]byte(usdt_token_id)))
    bytes464 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes464, total_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes464))
    bytes564 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes564, special_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes564))
    bytes664 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes664, staking_reward_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes664))
    _args = append(_args, hex.EncodeToString(treasury_address))
    _args = append(_args, hex.EncodeToString(staking_reward_address))
    _args = append(_args, hex.EncodeToString(burner_address))
    _args = append(_args, hex.EncodeToString(unwrap_address))
    _args = append(_args, hex.EncodeToString(registering_cost.Bytes()))
    dataField := "setConfig" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetTotalFeePercent(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, total_fee_percent uint64) error {
    _args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, total_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes064))
    dataField := "setTotalFeePercent" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetSpecialFeePercent(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, special_fee_percent uint64) error {
    _args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, special_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes064))
    dataField := "setSpecialFeePercent" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetStakingRewardFeePercent(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, staking_reward_fee_percent uint64) error {
    _args := make([]string, 0)
    bytes064 := make([]byte, 8)
    binary.BigEndian.PutUint64(bytes064, staking_reward_fee_percent)
    _args = append(_args, hex.EncodeToString(bytes064))
    dataField := "setStakingRewardFeePercent" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetStakingRewardAddress(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, staking_reward_address Address) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(staking_reward_address))
    dataField := "setStakingRewardAddress" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetTreasuryAddress(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, treasury_address Address) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(treasury_address))
    dataField := "setTreasuryAddress" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetBurnerAddress(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, burner_address Address) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(burner_address))
    dataField := "setBurnerAddress" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetUnwrapAddress(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, unwrap_address Address) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(unwrap_address))
    dataField := "setUnwrapAddress" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetRegisteringCost(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, registering_cost *big.Int) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(registering_cost.Bytes()))
    dataField := "setRegisteringCost" + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) EnableSwap(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "enableSwap" + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) CreatePair(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, first_token_id TokenIdentifier, second_token_id TokenIdentifier) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString([]byte(first_token_id)))
    _args = append(_args, hex.EncodeToString([]byte(second_token_id)))
    dataField := "createPair" + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) IssueLpToken(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "issueLpToken" + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) SetLpTokenLocalRoles(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "setLpTokenLocalRoles" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetPairActive(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "setPairActive" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetPairActiveButNoSwap(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "setPairActiveButNoSwap" + "@" + strings.Join(_args, "@")
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
func (contract *OneDex) SetPairInactive(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, pair_id uint32) error {
    _args := make([]string, 0)
    bytes032 := make([]byte, 4)
    binary.BigEndian.PutUint32(bytes032, pair_id)
    _args = append(_args, hex.EncodeToString(bytes032))
    dataField := "setPairInactive" + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) AddInitialLiquidity(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64) error {
    dataField := hex.EncodeToString([]byte("addInitialLiquidity"))
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

func (contract *OneDex) AddLiquidity(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, first_token_amount_min *big.Int, second_token_amount_min *big.Int) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(first_token_amount_min.Bytes()))
    _args = append(_args, hex.EncodeToString(second_token_amount_min.Bytes()))
    dataField := hex.EncodeToString([]byte("addLiquidity")) + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) RemoveLiquidity(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, first_token_amount_min *big.Int, second_token_amount_min *big.Int, unwrap_required bool) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(first_token_amount_min.Bytes()))
    _args = append(_args, hex.EncodeToString(second_token_amount_min.Bytes()))
    if unwrap_required {_args = append(_args, "01") } else {_args = append(_args, "00")}
    dataField := hex.EncodeToString([]byte("removeLiquidity")) + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) SwapMultiTokensFixedInput(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, amount_out_min *big.Int, unwrap_required bool, path_args []TokenIdentifier) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(amount_out_min.Bytes()))
    if unwrap_required {_args = append(_args, "01") } else {_args = append(_args, "00")}
    for _, elem := range path_args {
        _args = append(_args, hex.EncodeToString([]byte(elem)))
    }
    dataField := hex.EncodeToString([]byte("swapMultiTokensFixedInput")) + "@" + strings.Join(_args, "@")
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

func (contract *OneDex) SwapMultiTokensFixedOutput(_pk []byte, _value float64, _gasLimit uint64, _token *data.ESDT, _nonce uint64, amount_out_wanted *big.Int, unwrap_required bool, path_args []TokenIdentifier) error {
    _args := make([]string, 0)
    _args = append(_args, hex.EncodeToString(amount_out_wanted.Bytes()))
    if unwrap_required {_args = append(_args, "01") } else {_args = append(_args, "00")}
    for _, elem := range path_args {
        _args = append(_args, hex.EncodeToString([]byte(elem)))
    }
    dataField := hex.EncodeToString([]byte("swapMultiTokensFixedOutput")) + "@" + strings.Join(_args, "@")
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

