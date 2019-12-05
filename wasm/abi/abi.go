package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/accounts/abi"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"go-eth/eth"
"math/big"
"strings"
)

var abibyte = `[
{
"constant": false,
"inputs": [
{
"name": "key",
"type": "string"
},
{
"name": "val",
"type": "string"
}
],
"name": "put",
"outputs": [],
"payable": true,
"stateMutability": "payable",
"type": "function"
},
{
"constant": true,
"inputs": [
{
"name": "key",
"type": "string"
}
],
"name": "get",
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
"inputs": [],
"name": "getcounter",
"outputs": [
{
"name": "",
"type": "uint256"
}
],
"payable": false,
"stateMutability": "view",
"type": "function"
}
]`

var (
host     = "http://127.0.0.1:8546"
privKeys = "5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15"
)


//solidity调用wasm合约
func main() {
abi, err := abi.JSON(strings.NewReader(abibyte))
if err != nil {
fmt.Println("abi JSON err", err)
}
abiBytes, err := abi.Pack("put", "k", "v")
if err != nil {
fmt.Println("abi Pack err", err)
}
client, err := eth.Connect(host)
if err != nil {
fmt.Println("Connect err", err)
}
key, _ := crypto.HexToECDSA(privKeys)

from := crypto.PubkeyToAddress(key.PublicKey)

nonce, _ := client.EthClient.PendingNonceAt(context.TODO(), from)

to := common.HexToAddress("0x52dEf9374CA449FB65D8F85Bcea73826D81EA227")

tx := types.NewTransaction(nonce, to, new(big.Int), 900000, big.NewInt(18).Mul(big.NewInt(18), big.NewInt(1e9)), abiBytes)

sigtx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(738)), key)

hashes, err := client.SendRawTransaction(context.TODO(), sigtx)
if err != nil {
fmt.Println("SendRawTransaction err", err)
}
fmt.Println("txHash", hashes.Hex())
}
