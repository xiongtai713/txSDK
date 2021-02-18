package main

import (
	"context"
	"fmt"
	"pdx-chain/accounts/abi"
	"pdx-chain/crypto"
	"pdx-chain/utopia/utils/client"

	"log"
	"math/big"
	common2 "pdx-chain/common"
	types2 "pdx-chain/core/types"
	crypto2 "pdx-chain/crypto"
	"time"

	"strings"
)

var To = common2.HexToAddress("0xDa3Ce11D916fFBa4a1289cEf66A7f142eC5A0f74")
var tset = "sss"


func main() {

	myContractAbi := `[
	{
		"constant": false,   
		"inputs": [],
		"name": "get1",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "s",
				"type": "string"
			}
		],
		"name": "put1",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
	/*
	    constant 是否改变合约
	   	inputs  方法参数
	    name 方法名称
	    outputs 返回值
	    payable 是否可以转账
	    type 类型 function，constructor，fallback（缺省方法）

	*/

	abi, err := abi.JSON(strings.NewReader(myContractAbi))
	if err != nil {
		log.Fatalln("1", err)
	}

	abiBuf, err := abi.Pack("put1", "liu")
	if err != nil {
		log.Fatalln("2", err)
	}
	fmt.Println("abibuf", fmt.Sprintf("%x", abiBuf))
	//0x841321fd000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000
	//0x841321fd000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000
	//client, err := client.Connect("http://10.0.0.66:33333")
	client, err := client.Connect("http://127.0.0.1:8547")

	if err != nil {
		log.Fatalln("3", err)
	}

	privKey := "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"
	pri, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal("err", err)
		return
	}
	from := crypto2.PubkeyToAddress(pri.PublicKey)

	//To := common.HexToAddress("0x472eAC7e0d57B886A98b1371AC044A4679E41835")

	fmt.Printf("from:%s\n", from.String())
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	nonce, err := client.EthClient.NonceAt(ctx, from, nil)

	if err != nil {
		fmt.Println("nonce err", err)
		return
	}
	amount := big.NewInt(0)
	fmt.Println(from.String())
	gasLimit := uint64(4712388)
	gasPrice := new(big.Int)
	gasPrice.Mul(big.NewInt(4000), big.NewInt(1e9)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了
	x:=50000
	//nonce=10

	for i:=0;i<x;i++ {
		tx := types2.NewTransaction(uint64(nonce), To, amount, gasLimit, gasPrice, abiBuf)
		signer := types2.NewEIP155Signer(big.NewInt(777))
		signedTx, _ := types2.SignTx(tx, signer, pri)
		//EIP155 signer
		//signer := types.HomesteadSigner{}
		// client.EthClient.SendTransaction(context.TODO(), signedTx)

			if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
				fmt.Println("error", err.Error())

			} else {

				fmt.Println("result:", txHash.String(),"数量",i)
			}


		nonce++
	}
}