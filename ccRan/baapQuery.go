package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"pdx-chain/ethclient"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/common"
	"pdx-chain/crypto/sha3"
	"pdx-chain/rpc"
	"time"
)

func main() {
	//if client, err := eth1.Connect("http://192.168.0.120:8546"); err != nil {
	if client, err := ToolConnect2("http://10.0.0.113:22222"); err != nil {
		//if client, err := eth1.Connect("http://39.100.84.247:8545"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		to := eKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.7").String()
		invocation := &protos.Invocation{
			Fcn:  "get", //getHis,get,getRange,del,
			Args: [][]byte{[]byte("a")},
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
		//client.rpcClient.CallContext(c, &result, "eth_baapQuery", common.ToHex(data))
		result, err = client.BaapQuery1(c, data)
		//if err != nil {
		//	fmt.Println("query tx error", "err", err)
		//	return
		//}
		fmt.Printf("ellips:%s\n", time.Now().Sub(start).String())

		fmt.Println("baap query", "resp", result)
		//a := hex.EncodeToString(result[2:])
		if err != nil {
			fmt.Printf("err:%v", err)
			return
		}

		//fmt.Printf("query result:\n%s\n", string(a))
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

func ToolConnect2(host string, proxy ... string) (*Client1, error) {
	rpcClient, err := rpc.Dial(host,proxy...)
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)

	return &Client1{rpcClient, ethClient}, nil
}