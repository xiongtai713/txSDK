package  main

import  (
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

const  (
solidityContract          =  "0xdd1588125971b4ca7a876a4ea7ae177e69bf436c"
//solidityContract          ="0x6741447c33a6ba5e3530880d554be3c2b6f0e7bc"
//contract          =  "0xdd1588125971b4ca7a876a4ea7ae177e69bf436c"
//node                  =  "http://10.1.0.104:8545"
//node                  =  "http://10.0.0.110:8545"
mynode                  =  "http://127.0.0.1:8546"
//mynode                  =  "http://127.0.0.1:8545"
mysendDuration  =  time.Millisecond  *  10
)

var  myinputs  =  []string{
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013100000000000000000000000000000000000000000000000000000000000000",  //0:  1
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013200000000000000000000000000000000000000000000000000000000000000",  //1:  2
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013300000000000000000000000000000000000000000000000000000000000000",  //2:  3
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013400000000000000000000000000000000000000000000000000000000000000",  //3:  4
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013500000000000000000000000000000000000000000000000000000000000000",  //4:  5
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013600000000000000000000000000000000000000000000000000000000000000",  //5:  6
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013700000000000000000000000000000000000000000000000000000000000000",  //6:  7
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013800000000000000000000000000000000000000000000000000000000000000",  //7:  8
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000013900000000000000000000000000000000000000000000000000000000000000",  //8:  9
"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000231300000000000000000000000 00000000000000000000000000000000000000", //9: 10
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000033130300000000000000000000000000000000000000000000000000000000000", //10: 100
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016100000000000000000000000000000000000000000000000000000000000000", //11: a
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016200000000000000000000000000000000000000000000000000000000000000", //12: b
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016300000000000000000000000000000000000000000000000000000000000000", //13: c
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016400000000000000000000000000000000000000000000000000000000000000", //14: d
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016500000000000000000000000000000000000000000000000000000000000000", //15: e
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016600000000000000000000000000000000000000000000000000000000000000", //16: f
"a4136862000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000016700000000000000000000000000000000000000000000000000000000000000", //17: g
}

var myresult = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "100", "a", "b", "c", "d", "e", "f", "g"}


var CCContract = ccKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()

func main() {
log.SetFlags(log.Lmicroseconds)

var wg sync.WaitGroup

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
"a135f278036250fea709196921425b86f3755e70d1cc61d7a7d908ffcffb53cb",

}

for i, privKey := range privKeys[:5] {
log.Printf("i is = %d", i)
wg.Add(1)

go func(j int, key string) {
defer wg.Done()

sendChainCodeTx(
CCContract,
key,
//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
strconv.Itoa(j)+":")
}(i, privKey)

}

for i, privKey := range privKeys[5:] {
log.Printf("i is = %d", i)
wg.Add(1)

go func(j int, key string) {
defer wg.Done()

sendSolidityTx(
solidityContract,
key,
//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
strconv.Itoa(j)+":")
}(i, privKey)
}

wg.Wait()
}

func sendChainCodeTx(contract, privKey, flag string) {
//if client, err := eth1.ToolConnect("http://10.0.0.124:8546"); err != nil {
//if client, err := eth1.ToolConnect("http://10.0.0.124:8546"); err != nil {
if client, err := eth1.ToolConnect(mynode); err != nil {
//if client, err := eth1.ToolConnect("http://10.1.0.104:8545"); err != nil {
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

//1s内连续发交易，记录开始和结束时间戳
timer := time.NewTimer(mysendDuration)
log.Println(flag+"start:", time.Now().String())
i := 0
for {
select {
case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
return
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
return
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

func sendSolidityTx(contract, privKey string, flag string) {
//if client, err := eth1.ToolConnect("http://10.1.0.101:8546"); err != nil {
if client, err := eth1.ToolConnect(mynode); err != nil {
//if client, err := eth1.ToolConnect("http://10.0.0.124:8546"); err != nil {
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

//1s内连续发交易，记录开始和结束时间戳
timer := time.NewTimer(mysendDuration)
log.Println(flag+"start:", time.Now().String())
for {
select {
case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
log.Println(flag+"nonce", nonce)
return
default:
//time.Sleep(time.Second * 3)
log.Println(flag+"nonce is:", nonce)

r := rand.Intn(len(myinputs))
input := myinputs[r]
log.Println(flag+"rand inputs:::::::::::::::::::::::", r, "from:", from.String(), "result:", myresult[r])
data := common.Hex2Bytes(input) //todo 如何解析???

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

func ccKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

//fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}