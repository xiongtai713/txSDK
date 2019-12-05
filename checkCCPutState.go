package main

import (
"context"
"encoding/hex"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/crypto/sha3"
"github.com/golang/protobuf/proto"
"go-eth/eth"
"go-eth/eth/protos"
"time"
)

func main() {
if client, err := eth.Connect("http://192.168.0.120:8546"); err != nil {
fmt.Printf(err.Error())
return
} else {
to := xxKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
invocation := &protos.Invocation{
Fcn:  "get",
Args: [][]byte{[]byte("name")},
Meta: map[string][]byte{"to": []byte(to)},
}

payload, err := proto.Marshal(invocation)
if err != nil {
fmt.Printf("proto marshal invocation error:%v", err)
}

ptx := &protos.Transaction{
Type:    1, //1invoke 2deploy
Payload: payload,
}
data, err := proto.Marshal(ptx)
if err != nil {
fmt.Printf("!!!!!!!!proto marshal error:%v", err)
return
}

c, _ := context.WithTimeout(context.Background(), 800*time.Millisecond)
start := time.Now()
result, err := client.BaapQuery(c, data)
if err != nil {
fmt.Println("query tx error", "err", err)
return
}
fmt.Printf("ellips:%s\n", time.Now().Sub(start).String())

fmt.Println("baap query", "resp", result)
a, err := hex.DecodeString(result[2:])
if err != nil {
fmt.Printf("err:%v", err)
return
}

fmt.Printf("query result:%s\n", string(a))
}
}

func xxKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}