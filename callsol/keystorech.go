package main

import (
	"fmt"
	"log"
	"pdx-chain/accounts/keystore"
	"pdx-chain/crypto"
)

func main() {

	//
	//var keyStoreFlag = flag.String("key", "", "keystore")
	//var passwordFlag = flag.String("pw", "", "PassWord")
	//flag.Parse()
	//key, err := keystore.DecryptKey([]byte(*keyStoreFlag), *passwordFlag)
	//fmt.Println(*keyStoreFlag)
	//fmt.Println(*passwordFlag)
	//key, err := keystore.DecryptKey([]byte(`{"address":"e59dc7aec830883aea7a410f922644b7bfe3722b","crypto":{"cipher":"aes-128-ctr","ciphertext":"32af322c0d4ead7cc2046ab85839d1148a3ba169ca41fd316b6109dd2ce806f7","cipherparams":{"iv":"0bbedb246033270b8bb353f7071abb63"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"7f1bfd8bd117ff6bc3f52cd68b0df88cbde28ee6f4edb6f40308edf694e4e007"},"mac":"3c04dfb6c93265eb26d3eaf3ae36820c7001c440d74ad123bbbd477b5f416243"},"id":"f26b705a-0717-4f06-b5e2-237263ca4778","version":3}`), "466fdd66-e208-4b59-a83e-3f19c03c47f7")
	key, err := keystore.DecryptKey([]byte(`{"address":"27bfa304ede1be1a107639021da04e11e649d8a7","crypto":{"cipher":"aes-128-ctr","ciphertext":"b0890aa1c96c5cd5bf02423f9927d277d6297853ee5f8b6971c5f26c29b0eb9e","cipherparams":{"iv":"854ece61d38c8c3a9c2507e77c71725b"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"2eaebb9b0f21852113c53afb43369232de1fc17c925ea09070b494bcb9c482de"},"mac":"5285f041eceea9d742fb85f6aed2bfa9b97364092f09b04acd8ab6ce35a83fd1"},"id":"dd2bb448-1a2d-43b6-a73a-532304983205","version":3}`), "123456")

	if err != nil {
		log.Fatal("错误了", err)
	}

	//私钥
	fmt.Println(fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey)))

	//压缩公钥
	//compressedPubkey := crypto.CompressPubkey(&key.PrivateKey.PublicKey)
	//fmt.Println("compressed pukey:", hex.EncodeToString(compressedPubkey))
	//
	//
	//不带04
	//fmt.Println(hex.EncodeToString(discover.PubkeyID(&key.PrivateKey.PublicKey).Bytes()))
	//
	//带04  未压缩
	privKey := fmt.Sprintf("%x", crypto.FromECDSAPub(&key.PrivateKey.PublicKey))
	fmt.Println("1",len(privKey))


}
