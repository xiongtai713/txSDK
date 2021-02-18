package main

import (
	"fmt"
	"pdx-chain/common"
	"pdx-chain/crypto/sha3"
)

func main()  {


		hash := sha3.NewKeccak256()

		var buf []byte
		hash.Write([]byte("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02"))
		buf = hash.Sum(buf)

		fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())




}
