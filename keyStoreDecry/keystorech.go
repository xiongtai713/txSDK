package main

import (
	"fmt"
	"log"
	"pdx-chain/accounts/keystore"
	"pdx-chain/crypto"
)

func main() {
	key, err := keystore.DecryptKey([]byte(`{"address":"0d6361220d06798d18bb7fd252ec1059979b81cb","crypto":{"cipher":"aes-128-ctr","ciphertext":"0d6468d1bca37b75c227c30a19171301cd332c7c44968ef9c6fb7cceaf3fd067","cipherparams":{"iv":"ebb74ea239f8aaa86fd178c0a7d1ea6b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"dd24740f470bbe4b0a915d3526a46894241df3cc5313553722f048b49158f525"},"mac":"7b31bb5b84092b47eca492fbdad00dd9929a2a17d01d0deda0922f3ce8766398"},"id":"15e03bfb-fc8e-4585-8467-3ea647eec7bc","version":3}`), "123")
	if err != nil {
		log.Fatal("错误了", err)
	}
	privKey := fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey))
	fmt.Println(privKey)
}
