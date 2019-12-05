package main

import (
"fmt"
"log"
"pdx-chain/accounts/keystore"
"pdx-chain/crypto"
)

func main() {
key, err := keystore.DecryptKey([]byte(`{"address":"08b299d855734914cd7b19eea60dc84b825680f9","crypto":{"cipher":"aes-128-ctr","ciphertext":"bb838029b54dedea95108dcd97e47ba33ebf511d4839cd17d58d1cd8a03022f3","cipherparams":{"iv":"d1dc3c5a32582ef65075b5142fdaf04c"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"2a0cf023a8efef8e987b357d5ed3f72946b5e8441f5853f44443bb551ccb7dea"},"mac":"efe114073d5680a064c438cf5e7e56f6f6358e45d630e1ba498e726d0ad2304d"},"id":"e8082480-79d3-49c3-9d4c-a4a4c88fdaea","version":3}`), "123")
if err!=nil{
log.Fatal("错误了",err)
}
privKey := fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey))
fmt.Println(privKey)
}
