package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
	"go-eth/callsol"
)

func main() {
if client, err := eth1.ToolConnect("http://10.0.0.148:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
txHash := common.HexToHash("0x7037b32278f0d7d5cf83ac1bb5365bab2b51df22aa66693bce3aa35b0ddf80e3")
receiptChan := make(chan *types.Receipt)
fmt.Printf("Transaction hash: %s\n", txHash.String())
_, isPending, _ := client.EthClient.TransactionByHash(context.TODO(), txHash)
fmt.Printf("Transaction pending: %v\n", isPending)
// check transaction receipt
client.CheckTransaction(context.TODO(), receiptChan, txHash, 1)
receipt := <-receiptChan
fmt.Printf("receipt:%+v\n", *receipt)
fmt.Printf("Transaction status: %v\n", receipt.Status)
}

}
