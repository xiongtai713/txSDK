package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"pdx-chain/accounts/keystore"
	"pdx-chain/crypto"
	"pdx-chain/p2p/discover"
)

func main() {


	var keyStoreFlag = flag.String("key", "", "keystore")
	var passwordFlag = flag.String("pw", "", "PassWord")
	flag.Parse()
	fmt.Println(*keyStoreFlag)
	fmt.Println(*passwordFlag)
	//key, err := keystore.DecryptKey([]byte(`{"address":"e59dc7aec830883aea7a410f922644b7bfe3722b","crypto":{"cipher":"aes-128-ctr","ciphertext":"32af322c0d4ead7cc2046ab85839d1148a3ba169ca41fd316b6109dd2ce806f7","cipherparams":{"iv":"0bbedb246033270b8bb353f7071abb63"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"7f1bfd8bd117ff6bc3f52cd68b0df88cbde28ee6f4edb6f40308edf694e4e007"},"mac":"3c04dfb6c93265eb26d3eaf3ae36820c7001c440d74ad123bbbd477b5f416243"},"id":"f26b705a-0717-4f06-b5e2-237263ca4778","version":3}`), "466fdd66-e208-4b59-a83e-3f19c03c47f7")
	key, err := keystore.DecryptKey([]byte(*keyStoreFlag), *passwordFlag)

	if err != nil {
		log.Fatal("错误了", err)
	}

	//不带04
	fmt.Println(hex.EncodeToString(discover.PubkeyID(&key.PrivateKey.PublicKey).Bytes()))

	//带04
	privKey := fmt.Sprintf("%x", crypto.FromECDSAPub(&key.PrivateKey.PublicKey))
	fmt.Println(privKey)
}
