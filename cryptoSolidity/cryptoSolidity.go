package main

import (
"context"
"crypto/aes"
"crypto/cipher"
read "crypto/rand"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"go-eth/eth"
"io"
"log"
"math/big"
"math/rand"
"strconv"
"sync"
"time"

)


//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

const (
contract          =  "0x637e60dbe2cc091c14edbe72b6153a0642bf2cfd"
//contract = "0xd964de4e500a0d6abaca5545a4ea31feb6c9a652"
//contract            =  "0xdd1588125971b4ca7a876a4ea7ae177e69bf436c"
//node                  =  "http://127.0.0.1:8546"
node = "http://39.100.92.41:8545"

sendDuration  = time.Second * 5000000000
sleepDuration = time.Minute * 10 //运行多久之后
sleep         = time.Minute * 1  //休眠多久
)

var privKeys = []string{
"a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa",
"5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15",
"65035d9621f7be3bb6dc1f5a646e6ee2ef6bddf3f1ce57782d409c23857401a6",
"6043505f0f0e3748d7b83bfef873ed1d3224998adae81fc85bd059c5e0a75baf",
"e1b918a4006cb3651bfaf75cea0f885a948ea5ae5ade33fa7d22a3a080f6878e",
"32869ab5cedc7d88d7c802744d0050380abfa566146b1b0836a21c3d5bb9df2a",
"2a05f5e97817f97080ae9304c2c45fddf762096eb87da86eb651640a418c94c8",
"9feafa9ab9903f7ddf771864fc7c5844cc9e49fcc1b34bfcfbee3bcace6b0c29",
"51151f67c0a5e046caa0803b3decafec67706439ef1ec8fdbc6de400588f79f8",
//"a135f278036250fea709196921425b86f3755e70d1cc61d7a7d908ffcffb53cb",
//"29e17094583338e7d1b5d1062892ed41573a5283c939d320c06428e1c9b74d3f",
//"94fbab09742cadcc06cb2e663654547e618cd99ab72ecebeab529507fbab8bb5",
//"423e199b057c6206d85081a2ea2b72563e57d36c625c4187f30dfc1763798b03",
//"73d045af8b9892c54aad8286306d124b29f3da4022ff556930cb83d431404592",
//"0ed3332b016e4bbf8d4834399a9853983049b63b246aeb1fd6fdb690791c38f4",
//"4e393c8f2d2005c43f25e2b812963008e1d6f9e73c0e2008744841d87f7fd2bc",
//"2629938c775ab2fc0da7a00c0c0e7eb2300a77fe5bac33cc0d242ad291f4b657",
//"0c961cc45471597714359dfa9cbe74b1296ed434e92e050d58c2ecb82245c394",
//"284f4389d08a9f2a5cfdc2542521e17b03ea9043977791e5abc528307e93a723",
//"15375c54577d5bf690ac67bde21d9ed65b5690ece54078051acee700581c0228",
//"385896a8f68d2ec40afb822bbedf3594d20aa8f29b5e281759a15b30eb1f244d",
//"55bfbf57835787ba3cb37e5396c22ec71343acd5393ade23b2b38040c8e21747",
//"244435c591ca08aad2cf08a94cc45ba9687cd39dc2d8214282b56c41efa34492",
//"3027b5c4091068562c3e84c9a4bb96e33cd79f047abcc237e47af25516758102",
//"7f1f328c55b8973c07c66558ba0dc60a8eac9c658713bd9cb3573d668bd12f24",
//"7d86677bb13ec09bbd8f8b7d9e4f94cb3f3b948784eeb2a5d2625cf130afad3d",
//"68cfded92d20d5cae0f43a0b147fbfe7f9a0fd4bd347b845a4587c24a1ab3928",
//"6755e28172afc0b043172684d7c3dfeda1f46fab17a0883d697be6f90e5e605d",
//"74b6769a0d92f8a4bb45f314339ee5cc6baec10a3d3425a18a189578ac79444b",
//"21daf42f33d5b282bc52ad0001d8ce6898da997e579f2c233f2ffd2f1d08829b",
//"2b44621ce97f0f3e0759219178b5e6cc890519b7dc0c80c40d161a827fccf245",
}

var inputs = []string{
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000",  //0:  1
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013200000000000000000000000000000000000000000000000000000000000000",  //1:  2
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013300000000000000000000000000000000000000000000000000000000000000",  //2:  3
"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000134000000000000000000000000000000000000000000000000000000000000 00", //3: 4
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013500000000000000000000000000000000000000000000000000000000000000",  //4: 5
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013600000000000000000000000000000000000000000000000000000000000000",  //5: 6
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013700000000000000000000000000000000000000000000000000000000000000",  //6: 7
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013800000000000000000000000000000000000000000000000000000000000000",  //7: 8
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013900000000000000000000000000000000000000000000000000000000000000",  //8: 9
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000023130000000000000000000000000000000000000000000000000000000000000",  //9: 10
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000033130300000000000000000000000000000000000000000000000000000000000",  //10: 100
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016100000000000000000000000000000000000000000000000000000000000000",  //11: a
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000",  //12: b
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016300000000000000000000000000000000000000000000000000000000000000",  //13: c
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016400000000000000000000000000000000000000000000000000000000000000",  //14: d
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016500000000000000000000000000000000000000000000000000000000000000",  //15: e
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016600000000000000000000000000000000000000000000000000000000000000",  //16: f
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016700000000000000000000000000000000000000000000000000000000000000",  //17: g
}

var result = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "100", "a", "b", "c", "d", "e", "f", "g"}

func main() {
log.SetFlags(log.Lmicroseconds)

var wg sync.WaitGroup
//contract := "0xdd1588125971b4ca7a876a4ea7ae177e69bf436c"
for i, privKey := range privKeys {
log.Printf("i is = %d", i)

wg.Add(1)
go func(j int, key string) {
defer wg.Done()
sendCCTx2(
contract,
key,
inputs,
strconv.Itoa(j)+":")
}(i, privKey)
}

wg.Wait()
}

func sendCCTx2(contract, privKey string, inputs []string, flag string) {
//if client, err := eth.Connect("http://10.1.0.101:8546"); err != nil {
if client, err := eth.Connect(node); err != nil {
//if client, err := eth.Connect("http://10.0.0.124:8546"); err != nil {
//if client, err := eth.Connect("http://192.168.0.136:8545"); err != nil {
fmt.Errorf("Connect出错%s\n", flag+err.Error())
return
} else {
if privKey, err := crypto.HexToECDSA(privKey); err != nil {
log.Fatalln(flag+"privKey error:%s \n", err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf(flag+"from:%s\n", from.String())
to := common.HexToAddress(contract)
fmt.Printf(flag+"to:%s from:%s\n", to.String(), from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
log.Fatal(flag+"ddddd:%s", err.Error())
return
} else {
nonce = 0
amount := big.NewInt(0)
gasLimit := uint64(43766)
//gasLimit := uint64(4712388)
gasPrice := big.NewInt(26000000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

//1s内连续发交易，记录开始和结束时间戳
timer := time.NewTimer(sendDuration)
ticker := time.NewTicker(sleepDuration)
log.Println(flag+"start:", time.Now().String())

for {
select {
case <-ticker.C:
fmt.Println("sleep.........")
time.Sleep(sleep)
fmt.Println("get nonce again")
if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("nonce again err: %s", err.Error())
return
}

fmt.Println("nonce again is", nonce)

case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
log.Println(flag+"nonce", nonce)
return
default:
//time.Sleep(time.Second * 3)
log.Println(flag+"nonce is:", nonce)

r := rand.Intn(len(inputs))
input := inputs[r]
//进行加密
//key:=[]byte("1234567812345678")  //16位密码
//encrypt := AESEncrypt([]byte(input), key)

log.Println(flag+"rand inputs:::::::::::::::::::::::", r)
data := common.Hex2Bytes(string(input)) //todo 如何解析???

tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
log.Println(flag+"send raw transaction error", err.Error())
nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil)
for {
if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Errorf("nonce again err: %s", err.Error())
//出错5秒后重试
time.Sleep(5 * time.Second)
} else {
break
}
}
} else {
log.Printf(flag+"Transaction hash: %s\n", txHash.String())
nonce++
}
}
}

}
}
}

}

//加密
func AESEncrypt(plaintex ,key []byte )[]byte  {
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
func AESDencrytp(ciphertext ,key[]byte)[]byte  {
block,_:=aes.NewCipher(key)
iv:=ciphertext[:aes.BlockSize] //拿出前几位
ciphertext=ciphertext[aes.BlockSize:]//把明文拆成两段
//设置解密方式
stream:=cipher.NewCFBDecrypter(block,iv)
//解密
stream.XORKeyStream(ciphertext,ciphertext)
return ciphertext

}

