package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
	"go-eth/callsol"
"time"
)

func main() {
if client, err := eth1.ToolConnect("http://10.0.0.110:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
c, _ := context.WithTimeout(context.Background(), 800 * time.Millisecond)
start := time.Now()
withDrawTx, err := client.QueryTxByHash(c, common.HexToHash("0x83db1744f2f9e4733188c0c1ccd59cc3b68b951f9e2351e6231eec657eaa8830"))
if err != nil {
fmt.Println("query tx error", "err", err)
return
}
fmt.Printf("ellips:%s\n", time.Now().Sub(start).String())

fmt.Println("withDraw tx", "resp", withDrawTx)
fmt.Println("withDraw rpctx", "resp", *withDrawTx.RPCTx)
fmt.Println("withDraw status", "resp", withDrawTx.Status)
fmt.Println("withDraw currentCommit", "resp", withDrawTx.CurrentCommit.ToInt().String())
fmt.Println("withDraw commitNum", "resp", withDrawTx.CommitNumber.ToInt().String())
}
}
