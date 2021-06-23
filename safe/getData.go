package main

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"log"
	ethereum "pdx-chain"
	"pdx-chain/ethclient"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/pdxcc/util"
)

func getDomain(domain string) {

	di := &protos.DomainItem{}

	di.Name = domain


	dibyte, err := proto.Marshal(di)
	if err != nil {
		fmt.Println("proto.Marshal1", err)
		return
	}
	invocation := &protos.Invocation{}

	invocation.Args = make([][]byte, 1)
	invocation.Args[0] = dibyte
	invocation.Meta = make(map[string][]byte)
	invocation.Meta[ACTION] = PDXS_GET_DOMAIN

	invocationbyte, err := proto.Marshal(invocation)
	if err != nil {
		fmt.Println("proto.Marshal1", err)
		return
	}
	txp := &protos.Transaction{}
	txp.Payload = invocationbyte
	txp.Type = 1
	txp1, err := proto.Marshal(txp)
	if err != nil {
		fmt.Println("proto.Marshal2", err)
		return
	}
	data := txp1

	to := util.EthAddress("PDXSafe")
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	ctx := context.TODO()
	client, err := ethclient.Dial("http://127.0.0.1:8547")
	result, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		log.Fatalln("ss", err)
	}
	dir := &protos.DomainItem{}
	proto.Unmarshal(result, dir)

	fmt.Println(dir.Name)
	for _,n:=range dir.Name{


		fmt.Println(string(n))
	}

}

func getKey(domain ,key string) {

	ki := &protos.KeyItem{}
	ki.Key = key
	ki.Dname=domain
	dibyte, err := proto.Marshal(ki)
	if err != nil {
		fmt.Println("proto.Marshal1", err)
		return
	}
	invocation := &protos.Invocation{}

	invocation.Args = make([][]byte, 1)
	invocation.Args[0] = dibyte
	invocation.Meta = make(map[string][]byte)
	invocation.Meta[ACTION] = PDXS_GET_KEY

	invocationbyte, err := proto.Marshal(invocation)
	if err != nil {
		fmt.Println("proto.Marshal1", err)
		return
	}
	txp := &protos.Transaction{}
	txp.Payload = invocationbyte
	txp.Type = 1
	txp1, err := proto.Marshal(txp)
	if err != nil {
		fmt.Println("proto.Marshal2", err)
		return
	}
	data := txp1

	to := util.EthAddress("PDXSafe")
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
	ctx := context.TODO()
	client, err := ethclient.Dial("http://127.0.0.1:8547")
	result, err := client.CallContract(ctx, callMsg, nil)
	if err != nil {
		log.Fatalln("ss", err)
	}
	dir := &protos.DomainItem{}
	proto.Unmarshal(result, dir)
	for _,n:=range dir.Name{

		//w:=rune(n)
		fmt.Println(string(n))
	}
	fmt.Println(dir.Name)
}
