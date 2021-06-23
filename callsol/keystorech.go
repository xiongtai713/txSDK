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
	key, err := keystore.DecryptKey([]byte(`{"address":"b18d058c3d342932dddb3931094497b91f74e801","crypto":{"cipher":"aes-128-ctr","ciphertext":"413ee0c397f2a8b2cb62259e7850ede74736d705af51aaf7276d52a4793d645f","cipherparams":{"iv":"307a9b74d18219cf5e3674956faba617"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"3c343d3454b9b46506211596f409d7b5cfad7d72353a27c57f9f934e210299bf"},"mac":"68e147b22efbf610d28886f2e4a9df2febb9f27f4b19d7c6ad295cac19ce99a2"},"id":"a6ca5457-b516-4e86-9a3a-9a7f02643820","version":3}`), "48e9f541-e2a9-4019-bc71-942aeed93228")

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
