package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
	"go-eth/callsol"
"time"
)

func main() {
if client, err := eth1.ToolConnect("http://localhost:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
c, _ := context.WithTimeout(context.Background(), 2000 * time.Millisecond)
mytx := client.QueryUncompletedTxByAddr(c, common.HexToAddress("0x8000d109dAef5c81799bC01D4d82B0589dEEDb33"))
//mytx := client.QueryUncompletedTxByAddr(c, common.HexToAddress("0xde587CFf6C1137C4eb992cF8149ECFf3e163Ee07"))

for k, v := range mytx {
fmt.Printf("k[%d]:v[%s]", k, v.Hex())
}
fmt.Print("\n end.......................")

}
}
