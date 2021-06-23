package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
	"go-eth/callsol"
"log"
"math/big"
"strconv"
"sync"
"time"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

func main() {
log.SetFlags(log.Lmicroseconds)

var wg sync.WaitGroup

var privKeys = []string{
"a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa",
"5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15",
"65035d9621f7be3bb6dc1f5a646e6ee2ef6bddf3f1ce57782d409c23857401a6",
"6043505f0f0e3748d7b83bfef873ed1d3224998adae81fc85bd059c5e0a75baf",
//"e1b918a4006cb3651bfaf75cea0f885a948ea5ae5ade33fa7d22a3a080f6878e",
//"32869ab5cedc7d88d7c802744d0050380abfa566146b1b0836a21c3d5bb9df2a",
//"2a05f5e97817f97080ae9304c2c45fddf762096eb87da86eb651640a418c94c8",
//"9feafa9ab9903f7ddf771864fc7c5844cc9e49fcc1b34bfcfbee3bcace6b0c29",
//"51151f67c0a5e046caa0803b3decafec67706439ef1ec8fdbc6de400588f79f8",
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

}

contract := "0xdd1588125971b4ca7a876a4ea7ae177e69bf436c"
for i, privKey := range privKeys {
log.Printf("i is = %d", i)
wg.Add(1)

go func(j int, key string) {
defer wg.Done()

sendCCTx(
contract,
key,
"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
strconv.Itoa(j) + ":")
}(i, privKey)
}

wg.Wait()
}

func DD(j int) {
log.Printf("yyyy:%d", j)
}

func sendCCTx(contract, privKey, input, flag string) {
if client, err := eth1.ToolConnect("http://127.0.0.1:8546"); err != nil {
//if client, err := eth1.ToolConnect("http://10.1.0.101:8546"); err != nil {

//if client, err := eth1.ToolConnect("http://192.168.0.136:8545"); err != nil {
fmt.Printf(flag + err.Error())
return
} else {
if privKey, err := crypto.HexToECDSA(privKey); err != nil {
fmt.Printf(flag+"privKey error:%s \n", err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf(flag+"from:%s\n", from.String())
to := common.HexToAddress(contract)
fmt.Printf(flag+"to:%s from:%s\n", to.String(), from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf(flag+"ddddd:%s", err.Error())
return
} else {
amount := big.NewInt(0)
gasLimit := uint64(43766)
//gasLimit := uint64(4712388)
gasPrice := big.NewInt(240000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

data := common.Hex2Bytes(input) //todo 如何解析???

//1s内连续发交易，记录开始和结束时间戳
timer := time.NewTimer(time.Second * 30)
log.Println(flag+"start:", time.Now().String())
for {
select {
case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
return
default:
log.Println(flag+"nonce is:", nonce)
tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
log.Println(flag+"send raw transaction error", err.Error())
return
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
