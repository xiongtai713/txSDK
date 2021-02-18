package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/big"
	"pdx-chain/pdxcc"
	"pdx-chain/pdxcc/protos"
	client2 "pdx-chain/utopia/utils/client"

	"pdx-chain/common"
	"pdx-chain/crypto/sha3"

	"time"
)

func main() {
	//if client, err := eth.Connect("http://192.168.0.120:8546"); err != nil {
	if client, err := client2.Connect("http://39.100.84.247:8545"); err != nil {
		//if client, err := eth.Connect("http://39.100.84.247:8545"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		to := eKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.7").String()
		invocation := &protos.Invocation{
			Fcn:  "getHis", //getHis,get,getRange,del,
			Args: [][]byte{[]byte("3")},
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
		var result []byte
		client.rpcClient.CallContext(c, &result, "eth_baapQuery", common.ToHex(data))
		//result, err := client.BaapQuery(c, data)
		//if err != nil {
		//	fmt.Println("query tx error", "err", err)
		//	return
		//}
		fmt.Printf("ellips:%s\n", time.Now().Sub(start).String())

		fmt.Println("baap query", "resp", result)
		a, err := hex.DecodeString(result[2:])
		if err != nil {
			fmt.Printf("err:%v", err)
			return
		}

		fmt.Printf("query result:\n%s\n", string(a))
	}
}

func eKeccak256ToAddress(ccName string) common.Address {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write([]byte(ccName))
	buf = hash.Sum(buf)

	fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

	return common.BytesToAddress(buf)
}
