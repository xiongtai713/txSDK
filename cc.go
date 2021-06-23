package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/crypto/sha3"
"github.com/golang/protobuf/proto"
	"go-eth/callsol"
	"go-eth/callsol/protos"
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
nonceTicker = time.Millisecond * 1
sleepCCDuration = time.Second * 40  //运行多久之后
runNewTime = time.Second*30000000
)




func main() {
log.SetFlags(log.Lmicroseconds)

var wg sync.WaitGroup

var privKeys = []string{
//"a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa",
"5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15",
//"65035d9621f7be3bb6dc1f5a646e6ee2ef6bddf3f1ce57782d409c23857401a6",
//"6043505f0f0e3748d7b83bfef873ed1d3224998adae81fc85bd059c5e0a75baf",
//"e1b918a4006cb3651bfaf75cea0f885a948ea5ae5ade33fa7d22a3a080f6878e",
//"32869ab5cedc7d88d7c802744d0050380abfa566146b1b0836a21c3d5bb9df2a",
//"2a05f5e97817f97080ae9304c2c45fddf762096eb87da86eb651640a418c94c8",
//"9feafa9ab9903f7ddf771864fc7c5844cc9e49fcc1b34bfcfbee3bcace6b0c29",
//"51151f67c0a5e046caa0803b3decafec67706439ef1ec8fdbc6de400588f79f8",
//"a135f278036250fea709196921425b86f3755e70d1cc61d7a7d908ffcffb53cb",
//"29e17094583338e7d1b5d1062892ed41573a5283c939d320c06428e1c9b74d3f",
//"94fbab09742cadcc06cb2e663654547e618cd99ab72ecebeab529507fbab8bb5",
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

//contract := ggKeccak256ToAddress("eb7533d56603cf2defb12e419f61026283436c62:lj1:1.0.0").String()
contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()

for i, privKey := range privKeys {
log.Printf("i is = %d", i)
wg.Add(1)

go func(j int, key string) {
defer wg.Done()

sendTestTx(
contract,
key,
"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
strconv.Itoa(j)+":")
}(i, privKey)
}

wg.Wait()
}

func sendTestTx(contract, privKey, input, flag string) {
//if client, err := eth1.ToolConnect("http://39.98.244.76:30165"); err != nil {
if client, err := eth1.ToolConnect("http://127.0.0.1:8546"); err != nil {

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
gasPrice := big.NewInt(46000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

//1s内连续发交易，记录开始和结束时间戳
timer := time.NewTimer(runNewTime)
log.Println(flag+"start:", time.Now().String())
rand.Seed(time.Now().UnixNano())
i := rand.Int()
ticker := time.NewTicker(nonceTicker)
for {
select {
case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
return
case <-ticker.C:
fmt.Println("sleep.........")
time.Sleep(sleepCCDuration)
fmt.Println("get nonce again")
for{
if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Errorf("nonce again err: %s", err.Error())
//出错5秒后重试
time.Sleep(5*time.Second)
}else {
break
}
}


fmt.Println("nonce again is", nonce)



default:
log.Println(flag+"nonce is:", nonce)

invocation := &protos.Invocation{
Fcn:  "put",
Args: [][]byte{[]byte("name"), []byte(strconv.Itoa(i))},
Meta: map[string][]byte{"baap-tx-type": []byte("exec")},
}

payload, err := proto.Marshal(invocation)
if err != nil {
fmt.Printf("proto marshal invocation error:%v", err)
}

ptx := &protos.Transaction{
Type:    1, //1invoke 2deploy
Payload: payload,
}
data, err := proto.Marshal(ptx)
if err != nil {
fmt.Printf("!!!!!!!!proto marshal error:%v", err)
return
}

tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
log.Println(flag+"send raw transaction error", err.Error())
nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil)
for{
if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Errorf("nonce again err: %s", err.Error())
//出错5秒后重试
time.Sleep(5*time.Second)
}else {
break
}
}

} else {
log.Printf(flag+"Transaction hash: %s, i:%d, nonce:%d\n", txHash.String(), i, nonce)
i++
nonce++
}
}
}
}
}
}

}

func ggKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()
var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}