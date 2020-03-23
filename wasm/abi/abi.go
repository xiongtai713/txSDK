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
	host     = "http://10.0.0.138:34321"
	privKeys = "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"
)

//solidity调用wasm合约
func main() {
	abi, err := abi.JSON(strings.NewReader(abibyte))
	if err != nil {
		fmt.Println("abi JSON err", err)
	}
	abiBytes, err := abi.Pack("put","pdx","123")
	if err != nil {
		fmt.Println("abi Pack err233", err)
	}
	client, err := eth.Connect(host)
	if err != nil {
		fmt.Println("Connect err", err)
	}
	key, _ := crypto.HexToECDSA(privKeys)

	from := crypto.PubkeyToAddress(key.PublicKey)

	nonce, _ := client.EthClient.PendingNonceAt(context.TODO(), from)

	to := common.HexToAddress("0x6e0d7af8a1291c55297378c085cb92731c4a52f9")

	tx := types.NewTransaction(nonce, to, new(big.Int), 900000, big.NewInt(18).Mul(big.NewInt(18), big.NewInt(1e9)), abiBytes)

	sigtx, _ := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(739)), key)

	hashes, err := client.SendRawTransaction(context.TODO(), sigtx)
	if err != nil {
		fmt.Println("SendRawTransaction err", err)
	}
	fmt.Println("txHash", hashes.Hex())
}
