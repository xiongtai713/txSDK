package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"pdx-chain/crypto"
	"pdx-chain/crypto/ecies"
)

func main() {

	for i:=0;i<=20000;i++{
		prv ,err := ecies.GenerateKey(rand.Reader, crypto.S256(), nil)
		if err != nil {
			panic(err)
		}

		fmt.Println(hex.EncodeToString(crypto.FromECDSA(prv.ExportECDSA())))

	}

	//ecdsa私钥转地址
	//privateKey, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)

	//priv1 := "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84" //860
	//priv2 := "a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381" //7DE
	//
	//privateKey, err := crypto.HexToECDSA(priv2)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//pubBytes := crypto.FromECDSAPub(&privateKey.PublicKey) //公钥转byte
	//compressPubk:=crypto.CompressPubkey(&privateKey.PublicKey)
	//
	//fmt.Println("未压缩公钥",hex.EncodeToString(pubBytes))
	//fmt.Println("压缩未",hex.EncodeToString(compressPubk))
	//fmt.Println("地址",common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:]).String())
}

