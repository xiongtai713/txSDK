package main

import (
"context"
"fmt"
"go-eth/eth"
"math/big"
"time"
)

func main() {
if client, err := eth.Connect("http://localhost:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
c, _ := context.WithTimeout(context.Background(), 2000 * time.Millisecond)
//mytx := client.QueryUncompletedTxByAddr(c, common.HexToAddress("0x8000d109DAef5C81799bC01D4d82B0589dEEDb33"))
block, err := client.GetCommitBlockByNum(c, big.NewInt(3))
if err != nil {
fmt.Println("err:", err.Error())
return
}
fmt.Printf("block:%+v", block)

}
}
