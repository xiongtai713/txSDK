package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"math/big"
)

func sendTestTx(privKey1 string,  client *Client, nonce uint64) error {
	to1:="0x69633dbb9ebc560ca28f94e18d743f774b8ef01c"



	//amount := big.NewInt(0).Mul(big.NewInt(1000), big.NewInt(1e18))
	amount := big.NewInt(1)
	gasLimit := uint64(12021000)
	gasPrice := big.NewInt(7000 * 1e10) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

	privKey3, _ := crypto.HexToECDSA(privKey1)

	data := make([]byte, 1000*300)
	//n1 := rand.Int31n(9)
	//n2 := rand.Int31n(9)
	//n3 := rand.Int31n(9)
	//n4 := rand.Int31n(9)
	//to := common.HexToAddress(fmt.Sprintf("0x08b299d855734914cd719ea60dc84b25680f%d%d%d%d", n1, n2, n3, n4))
	to := common.HexToAddress(to1)

	//fmt.Printf("to:%s  \n", to.String())
	//from := crypto.PubkeyToAddress(privKey1.PublicKey)
	//
	//nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)

	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)

	signer := types.NewEIP155Signer(big.NewInt(739))
	signedTx, err := types.SignTx(tx, signer, privKey3)
	_, err = client.SendRawTransaction(context.TODO(), signedTx)
	if err != nil {
		fmt.Println("SendRawTransactionErr", err)
		fmt.Println("SendRawTransactionErr nonce",nonce,)
		return err
	}
	//fmt.Printf("Transaction hash: %s, nonce %d\n", txHash.String(), nonce, )
	return nil
}
