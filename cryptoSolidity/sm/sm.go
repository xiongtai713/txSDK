package sm

import (
"bytes"
"crypto/cipher"
"fmt"
"github.com/tjfoc/gmsm/sm2"
"github.com/tjfoc/gmsm/sm3"
"github.com/tjfoc/gmsm/sm4"

"log"
)



func Sm2ENcrypt(msg []byte, pub *sm2.PublicKey) []byte {

ciphertxt, err := pub.Encrypt(msg)
if err != nil {
log.Fatal(err)
}
fmt.Printf("加密结果:%x\n", ciphertxt)
return ciphertxt

}

func Sm2Decrypt(ciphertxt []byte, priv *sm2.PrivateKey) []byte {
plaintxt, err := priv.Decrypt(ciphertxt)
if err != nil {
log.Fatal(err)
}
return plaintxt
}

func Sm3Hash(data []byte) []byte {
h := sm3.New()
h.Write([]byte(data))
sum := h.Sum(nil)
return sum

}

//sm4密钥是128比特的密钥,16个字节
//iv是128比特
func Sm4Encrypt(key, iv, plainText []byte) ([]byte, error) {
block, err := sm4.NewCipher(key)
if err != nil {
return nil, err
}
blockSize := block.BlockSize()
origData := pkcs5Padding(plainText, blockSize)
blockMode := cipher.NewCBCEncrypter(block, iv)
cryted := make([]byte, len(origData))
blockMode.CryptBlocks(cryted, origData)
return cryted, nil
}

func Sm4Decrypt(key, iv, cipherText []byte) ([]byte, error) {
block, err := sm4.NewCipher(key)
if err != nil {
return nil, err
}
blockMode := cipher.NewCBCDecrypter(block, iv)
origData := make([]byte, len(cipherText))
blockMode.CryptBlocks(origData, cipherText)
origData = pkcs5UnPadding(origData)
return origData, nil
}

// pkcs5填充
func pkcs5Padding(src []byte, blockSize int) []byte {
padding := blockSize - len(src)%blockSize
padtext := bytes.Repeat([]byte{byte(padding)}, padding)
return append(src, padtext...)
}

func pkcs5UnPadding(src []byte) []byte {
length := len(src)
unpadding := int(src[length-1])
return src[:(length - unpadding)]
}
