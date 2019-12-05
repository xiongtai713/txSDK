package sol

import (
read "crypto/rand"
"crypto/aes"
"crypto/cipher"
"encoding/hex"
"fmt"
"io"
"testing"
)

func TestAESEncrypt(t *testing.T) {
key := []byte("1234567812345678")
cipher := AESEncrypt1([]byte("ss"), key)
fmt.Println(hex.EncodeToString(cipher))
dencrytp := AESDencrytp1(cipher, key)
fmt.Println(string(dencrytp))
}

func TestAESDencrytp(t *testing.T) {
//endpoint := "http://localhost:8545"
//addr := common.HexToAddress(contract)
//
//client, _ := rpc.Dial(endpoint)
//ethClient := ethclient.NewClient(client)
//caller := sol.NewApiname(addr, ethClient)
}


func AESEncrypt1(plaintex ,key []byte )[]byte  {
block,_:=aes.NewCipher(key) //密钥分组
ciphertext:=make([]byte,aes.BlockSize+len(plaintex))
//设置[:aes.BlockSize]这段的内存空间可读
iv :=ciphertext[:aes.BlockSize] //数组切片
io.ReadFull(read.Reader,iv)//设置成可读
//设置加密模式
stream:=cipher.NewCFBEncrypter(block,iv)//返回一个加密流
//加密[利用ciphertext[:aes.BlockSize]]与明文做异或
stream.XORKeyStream(ciphertext[aes.BlockSize:],plaintex)//拿密文,跟整个明文做异或判断
return ciphertext

}


func AESDencrytp1(ciphertext ,key[]byte)[]byte  {
block,_:=aes.NewCipher(key)
iv:=ciphertext[:aes.BlockSize] //拿出前几位
ciphertext=ciphertext[aes.BlockSize:]//把明文拆成两段
//设置解密方式
stream:=cipher.NewCFBDecrypter(block,iv)
//解密
stream.XORKeyStream(ciphertext,ciphertext)
return ciphertext

}