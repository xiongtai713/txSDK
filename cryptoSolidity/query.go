package main

import (
"context"
"crypto/aes"
"crypto/cipher"
read "crypto/rand"
"encoding/json"
"fmt"
"github.com/tjfoc/gmsm/sm2"
"github.com/tjfoc/gmsm/sm4"

"go-eth/cryptoSolidity/sm"
"go-eth/cryptoSolidity/sol"
"io"
"math/big"
"pdx-chain/accounts/abi/bind"
"pdx-chain/common"
"pdx-chain/crypto"
"pdx-chain/ethclient"
"pdx-chain/rpc"
)

const (
contract1 = "0x137eB3207E55B4e96f092310696F1e854a93D5f9"
)

var ps = []byte("1234567812345678") //16位密码
var Sm4Key = []byte("1234567890abcdef") //sm4密码
var iv = make([]byte, sm4.BlockSize)
//sm2私钥
var SmPriKey = `{"Curve":{"RInverse":57896044551258231062740198220913455226441901632205615997740090104278065086466,"P":115792089210356248756420345214020892766250353991924191454421193933289684991999,"N":115792089210356248756420345214020892766061623724957744567843809356293439045923,"B":18505919022281880113072981827955639221458448578012075254857346196103069175443,"Gx":22963146547237050559479531362550074578802567295341616970375194840604139615431,"Gy":85132369209828568825618990617112496413088388631904505083283536607588877201568,"BitSize":256,"Name":"SM2-P-256"},"X":95358780753915548818615761914427850715792293415090713401623255125748501500131,"Y":69600721504380446880434664233264795080344811469653406074816717488674359372342,"D":112391593134245205376816760191347421678778244893082299210634629770515230151582}`

func main() {


//endpoint := "http://39.100.92.41:8545"
endpoint := "http://127.0.0.1:8546"

addr := common.HexToAddress(contract1)
client, _ := rpc.Dial(endpoint)
ethClient := ethclient.NewClient(client)
caller, _ := sol.NewApiname(addr, ethClient)
pri := new(sm2.PrivateKey)
json.Unmarshal([]byte(SmPriKey), pri)
pri.Curve = sm2.P256Sm2()
set(caller,ethClient,&pri.PublicKey)
//get(caller,pri)

}

func set(caller *sol.Apiname, ethClient *ethclient.Client, pub *sm2.PublicKey) {
key, err := crypto.HexToECDSA("5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15")
if err != nil {
fmt.Println("error", err)
}
from := crypto.PubkeyToAddress(key.PublicKey)
opts1 := bind.NewKeyedTransactor(key)
opts1.ChainId = big.NewInt(738)
nonce, _ := ethClient.NonceAt(context.TODO(), from, nil)
opts1.Nonce = big.NewInt(int64(nonce))
opts1.GasPrice = big.NewInt(39000000000)
opts1.GasLimit = uint64(4376600)
opts1.Value = big.NewInt(0)
opts1.Context = context.TODO()
fmt.Printf("发送的节点%+v nonce %v\n", opts1.From.String(),nonce)
//encrypt:= AESEncrypt1([]byte("xx"), ps)   //AES
//encrypt := sm.Sm2ENcrypt([]byte("pdx"), pub) //国密椭圆曲线sm2
encrypt, _ := sm.Sm4Encrypt(Sm4Key, iv, []byte("liu")) //国密椭圆曲线sm4
fmt.Println("存入的密文", string(encrypt),"存储的明文 liu",)
tx, err := caller.SetGreeting(opts1, string(encrypt)) //set
if err != nil {
fmt.Println("Set Error", err)
} else {
fmt.Printf("tx信息%+v\n", tx.Hash().Hex())
}

}

func get(caller *sol.Apiname, priv *sm2.PrivateKey) {
ctx := context.Background()
opts := &bind.CallOpts{
Context: ctx,
}
s, err := caller.Greet(opts) //get
if err != nil {
fmt.Println("Get Error", err)
} else {
fmt.Println("查询出的密文", string(s))
//dencrytp := AESDencrytp1([]byte(s), ps) //AES
//dencrytp := sm.Sm2Decrypt([]byte(s),priv) //国密椭圆曲线sm2
dencrytp, _ := sm.Sm4Decrypt(Sm4Key, iv, []byte(s)) //国密椭圆曲线sm4
fmt.Println("查询出的数据", string(dencrytp))
}
}

func AESEncrypt1(plaintex, key []byte) []byte {
block, _ := aes.NewCipher(key) //密钥分组
ciphertext := make([]byte, aes.BlockSize+len(plaintex))
//设置[:aes.BlockSize]这段的内存空间可读
iv := ciphertext[:aes.BlockSize] //数组切片
io.ReadFull(read.Reader, iv)     //设置成可读
//设置加密模式
stream := cipher.NewCFBEncrypter(block, iv) //返回一个加密流
//加密[利用ciphertext[:aes.BlockSize]]与明文做异或
stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintex) //拿密文,跟整个明文做异或判断
return ciphertext

}

func AESDencrytp1(ciphertext, key []byte) []byte {
block, _ := aes.NewCipher(key)
iv := ciphertext[:aes.BlockSize]        //拿出前几位
ciphertext = ciphertext[aes.BlockSize:] //把明文拆成两段
//设置解密方式
stream := cipher.NewCFBDecrypter(block, iv)
//解密
stream.XORKeyStream(ciphertext, ciphertext)
return ciphertext

}
