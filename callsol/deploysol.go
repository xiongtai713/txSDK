package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"strings"
)

func main() {
	//h := "0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000043535353500000000000000000000000000000000000000000000000000000000"
	//b, err := hex.DecodeString(h[2:])
	//if err != nil {
	//fmt.Println("decode err:", err.Error())
	//return
	//}

	// load contract ABI
	myContractAbi := `[
{
"constant": false,
"inputs": [],
"name": "kill",
"outputs": [],
"payable": false,
"stateMutability": "nonpayable",
"type": "function"
},
{
"constant": false,
"inputs": [
{
"name": "_newgreeting",
"type": "string"
}
],
"name": "setGreeting",
"outputs": [],
"payable": false,
"stateMutability": "nonpayable",
"type": "function"
},
{
"inputs": [
{
"name": "_greeting",
"type": "string"
}
],
"payable": false,
"stateMutability": "nonpayable",
"type": "constructor"
},
{
"constant": true,
"inputs": [],
"name": "greet",
"outputs": [
{
"name": "",
"type": "string"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
},
{
"constant": true,
"inputs": [
{
"name": "a",
"type": "uint8"
},
{
"name": "b",
"type": "uint8"
}
],
"name": "sum",
"outputs": [
{
"name": "",
"type": "uint8"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
}
]`
	abi, err := abi.JSON(strings.NewReader(myContractAbi))
	if err != nil {
		log.Fatalln("1", err)
	}

	abiBuf, err := abi.Pack("sum", 2, 1)
	if err != nil {
		log.Fatalln("2", err)
	}

	client, err := ethclient.Dial("http://127.0.0.1:8547")
	if err != nil {
		log.Fatalln("3", err)
	}

	//to := common.HexToAddress("0xEA1E9A0ab182Cf473EA2E027ae735C6352eb8098")
	callMsg := ethereum.CallMsg{
		To:   nil,
		Data: abiBuf,
	}
	ctx := context.TODO()
	result, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		log.Fatalln("4", err)
	}

	//r := ""
	//err = abi.Unpack(&r, "greet", result)
	//if err != nil {
	//log.Fatalln("5",err)
	//}

	fmt.Println("result:", result)
}
