package main

import (
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"pdx-chain/pdxcc/protos"
)

func main()  {
	input,err:=hex.DecodeString("080112230a0370757412037064781201301a140a0c626161702d74782d74797065120465786563")
	if err!=nil{
		fmt.Println("err",err)
	}

	ptx := &protos.Transaction{}
	proto.Unmarshal(input,ptx)
	invocation := &protos.Invocation{}
	proto.Unmarshal(ptx.Payload,invocation)
	fmt.Println(invocation)

}
