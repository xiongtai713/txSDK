package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"
)

func main() {


	a:=[]byte{3,4}

	bytes, _ := rlp.EncodeToBytes(a)
	fmt.Println(bytes)

}
