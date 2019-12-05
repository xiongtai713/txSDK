package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"go-eth/eth"
"time"
)

func main() {
if client, err := eth.Connect("http://10.0.0.148:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
//addr := MKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0")
addr := common.HexToAddress("0x83fe44cb9e75fac28c2560faead245f248dfa5f8")
c, _ := context.WithTimeout(context.Background(), 200 * time.Millisecond)
num, err := client.EthClient.BalanceAt(c, addr, nil)
if err != nil {
fmt.Println("err:", err.Error())
return
}
fmt.Println("num:", num.String())
}
}