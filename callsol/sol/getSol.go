package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"

	"strings"
)

var MyContractAbi = `[
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


func main() {
	//fmt.Println(tset)
	var to = common.HexToAddress("0xDd1588125971b4Ca7a876A4Ea7Ae177e69bf436C")

	abi, err := abi.JSON(strings.NewReader(MyContractAbi))
	if err != nil {
		log.Fatalln("1", err)
	}

	abiBuf, err := abi.Pack("get1")
	if err != nil {
		log.Fatalln("2", err)
	}
	//fmt.Println("abibuf", fmt.Sprintf("%x", abiBuf))

	//To := common.HexToAddress("0x472eAC7e0d57B886A98b1371AC044A4679E41835")

	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: abiBuf,
	}
	ctx := context.TODO()
	client, err := ethclient.Dial("http://127.0.0.1:8547")
	result, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		log.Fatalln("ss", err)
	}

	r := ""
	err = abi.Unpack(&r, "get1", result)
	if err != nil {
		log.Fatalln("Unpack", err)
	}
	fmt.Println(r)
	fmt.Println(result)

}
