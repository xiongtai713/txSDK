package main

import (
"context"
"fmt"
"log"
"pdx-chain"
"pdx-chain/accounts/abi"
"pdx-chain/common"
"pdx-chain/ethclient"
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
}
]`
abi, err := abi.JSON(strings.NewReader(myContractAbi))
if err != nil {
log.Fatalln(err)
}

a := uint8(11)
b := uint8(88)
abiBuf, err := abi.Pack("sum", b, a)
if err != nil {
log.Fatalln(err)
}

client, err := ethclient.Dial("http://39.100.94.15:8545")
if err != nil {
log.Fatalln(err)
}

to := common.HexToAddress("0xfE4f5525b519d74A780aC539464f071600F62A36")
callMsg := ethereum.CallMsg{
To:   &to,
Data: abiBuf,
}
ctx := context.TODO()
result, err := client.CallContract(ctx, callMsg, nil)
if err != nil {
log.Fatalln(err)
}

var r uint8
err = abi.Unpack(&r, "sum", result)
if err != nil {
log.Fatalln(err)
}

fmt.Println("result:", r)
}
