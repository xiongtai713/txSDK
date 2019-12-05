package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
)

func main() {


	for i:=0;i<=10000;i++{
		prv ,err := ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
		if err != nil {
			panic(err)
		}

		fmt.Println(hex.EncodeToString(crypto.FromECDSA(prv.ExportECDSA())))


	}

	//ecdsa私钥转地址
	//privateKey, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	//pubBytes := crypto.FromECDSAPub(&privateKey.PublicKey) //公钥转byte
	//fmt.Println(common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:]).String())
}
