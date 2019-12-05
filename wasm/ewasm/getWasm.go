package ewasm

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum"
"github.com/ethereum/go-ethereum/common"
"go-eth/eth"
"log"
)

func GetWasm( client *eth.Client, to common.Address) {
key:="pdx"
callMsg := ethereum.CallMsg{
To:   &to,
Data: []byte("get:"+key),
}

bytes, err := client.EthClient.CallContract(context.Background(), callMsg, nil)
if err!=nil{
log.Println(err)
}
fmt.Printf("key=%v,value=%v",key,string(bytes))
}
