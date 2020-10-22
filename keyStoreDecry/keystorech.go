package main

import (
	"fmt"
	"log"
	"pdx-chain/accounts/keystore"
	"pdx-chain/crypto"
)

func main() {
	key, err := keystore.DecryptKey([]byte(`{"address":"e8aec858dc98adb748a6c0f57e314174081d0fc0","crypto":{"cipher":"aes-128-ctr","ciphertext":"5c746e0e4acb60033d106783346abbf54425d285d27c6deaa2c655b25c29095d","cipherparams":{"iv":"f8323b02bb287b353df8c5b2cb95fa97"},"kdf":"scrypt","kdfparams":{"dklen":32,"c": null,"n":262144,"p":1,"r":8,"salt":"90623d80bf32205b562f16f13560ac063083b2f65301069c75684f0f69277490"},"mac":"ed57d81e3f2d1772fefb927fc2b879a6d4fe4eff573f3fa8e2a613706b4ffa60"},"id":"bd0d2579-4aaf-406b-831d-4328e07994dc","version":3}`), "123456789")
	if err != nil {
		log.Fatal("错误了", err)
	}

	privKey := fmt.Sprintf("%x", crypto.FromECDSAPub(&key.PrivateKey.PublicKey))

	//privKey := fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey))
	fmt.Println(privKey)
}
