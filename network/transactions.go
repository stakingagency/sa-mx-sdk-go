package network

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/multiversx/mx-sdk-go/blockchain"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/builders"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/stakingagency/sa-mx-sdk-go/data"
	"github.com/stakingagency/sa-mx-sdk-go/utils"
)

func (nm *NetworkManager) SendTransaction(privateKey []byte, receiver string, value float64, gasLimit uint64, dataField string, nonce uint64) (string, error) {
	w := interactors.NewWallet()
	sender, err := w.GetAddressFromPrivateKey(privateKey)
	if err != nil {
		return "", err
	}

	args := blockchain.ArgsProxy{
		ProxyURL:            nm.proxyAddress,
		Client:              nil,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       false,
		CacheExpirationTime: time.Minute,
		EntityType:          core.Proxy,
	}
	proxy, err := blockchain.NewProxy(args)
	if err != nil {
		return "", err
	}

	txArgs, _, err := proxy.GetDefaultTransactionArguments(context.Background(), sender, nm.netCfg)
	if err != nil {
		return "", err
	}

	txArgs.Receiver = receiver
	txArgs.Value = utils.Renominate(value, 18).String()
	txArgs.Data = []byte(dataField)

	if gasLimit != utils.AutoGasLimit {
		txArgs.GasLimit = gasLimit
	} else {
		txArgs.GasLimit += uint64(len(dataField)) * 1500
	}

	if nonce != utils.AutoNonce {
		txArgs.Nonce = nonce
	}

	holder, _ := cryptoProvider.NewCryptoComponentsHolder(keyGen, privateKey)
	txBuilder, err := builders.NewTxBuilder(cryptoProvider.NewSigner())
	if err != nil {
		return "", err
	}

	ti, err := interactors.NewTransactionInteractor(proxy, txBuilder)
	if err != nil {
		return "", err
	}

	err = ti.ApplySignature(holder, &txArgs)
	if err != nil {
		return "", err
	}

	hash, err := ti.SendTransaction(context.Background(), &txArgs)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func (nm *NetworkManager) SendEsdtTransaction(privateKey []byte, receiver string, value float64, gasLimit uint64, token *data.ESDT, function string, nonce uint64) (string, error) {
	iValue := utils.Renominate(value, int(token.Decimals))
	sValue := hex.EncodeToString(iValue.Bytes())
	sTicker := hex.EncodeToString([]byte(token.Ticker))
	dataField := fmt.Sprintf("ESDTTransfer@%s@%s", sTicker, sValue)
	if function != "" {
		dataField += "@" + function
	}
	if gasLimit == utils.AutoGasLimit {
		gasLimit = 500000
	}

	return nm.SendTransaction(privateKey, receiver, 0, gasLimit, dataField, nonce)
}
