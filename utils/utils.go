package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/hashing/keccak"
	logger "github.com/multiversx/mx-chain-logger-go"
	"golang.org/x/crypto/sha3"
	"golang.org/x/net/html/charset"
)

var log = logger.GetOrCreate("network")

func Base64Decode(s string) string {
	res, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}

	return string(res)
}

func UTF8(s string) string {
	r, err := charset.NewReader(strings.NewReader(s), "latin1")
	if err != nil {
		return ""
	}

	result, err := ioutil.ReadAll(r)
	if err != nil {
		return ""
	}

	return string(result)
}

func GetDNSAddress(username string) string {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write([]byte(username))
	hash := h.Sum(nil)
	shardId := hash[len(hash)-1]
	var initialDNSAddress = bytes.Repeat([]byte{1}, 32)
	shardInBytes := []byte{0, shardId}
	newDNSPk := string(initialDNSAddress[:(30)]) + string(shardInBytes)
	creatorAddress := []byte(newDNSPk)

	buffNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(buffNonce, 0)
	adrAndNonce := append([]byte(newDNSPk), buffNonce...)
	base := keccak.NewKeccak().Compute(string(adrAndNonce))

	prefixMask := make([]byte, 8)
	prefixMask = append(prefixMask, []byte{5, 0}...)
	suffixMask := creatorAddress[len(creatorAddress)-2:]

	copy(base[:10], prefixMask)
	copy(base[len(base)-2:], suffixMask)

	converter, _ := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	dnsAddress, _ := converter.Encode(base)

	return dnsAddress
}

func Denominate(iValue *big.Int, decimals int) float64 {
	fValue := big.NewFloat(0).SetInt(iValue)
	ten := big.NewFloat(10)
	for i := 0; i < decimals; i++ {
		fValue.Quo(fValue, ten)
	}
	res, _ := fValue.Float64()

	return res
}

func Renominate(value float64, decimals int) *big.Int {
	fValue := big.NewFloat(value)
	ten := big.NewFloat(10)
	for i := 0; i < decimals; i++ {
		fValue.Mul(fValue, ten)
	}
	res, _ := fValue.Int(nil)

	return res
}

type priceType struct {
	Symbol string `json:"symbol"`
	Price  string `price:"price"`
}

func GetBinancePrice(symbol string) float64 {
	body, err := GetHTTP(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol), "")
	if err != nil {
		return 0
	}
	var pt priceType
	json.Unmarshal(body, &pt)
	p, err := strconv.ParseFloat(pt.Price, 64)
	if err != nil {
		return 0
	}
	return p
}
