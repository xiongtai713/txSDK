package main

import (
	"fmt"
	"log"
	"pdx-chain/accounts/keystore"
	"pdx-chain/crypto"
)

func main() {
	key, err := keystore.DecryptKey([]byte(`{"address":"fe7ad648f106aef54a2f623d60dcddfba296f5eb","crypto":{"cipher":"aes-128-ctr","ciphertext":"c8682f8f55ecac057ff69fa2463ec2bb89b8f571f43356ac59ad5c5658161f15","cipherparams":{"iv":"cc84f5c032502ce1d07b92f7aead5092"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"2900cb2ab9f6f9e3c4ad1a2ba9e2f51a824b574dac87061e716c0fe8bf067a55"},"mac":"80b51d28168030ee56dd0c08e37a2b07596b08429985b78ef5bfaec6eb384fc8"},"id":"fee82865-9cbb-4674-a16c-d75bc67a62ed","version":4}`), "5919f4a8-dd16-46ec-8530-04a76a3714e5")
	if err != nil {
		log.Fatal("错误了", err)
	}

	//privKey := fmt.Sprintf("%x", crypto.FromECDSAPub(&key.PrivateKey.PublicKey))
	from := crypto.PubkeyToAddress(key.PrivateKey.PublicKey)

	//privKey := fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey))
	fmt.Println(from.String())
}
