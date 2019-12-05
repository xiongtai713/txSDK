package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"go-eth/eth"
"log"
"math/big"
"strconv"
"sync"
"time"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

const (
//host         = "http://39.100.34.235:30074"
//host         = "http://39.100.84.63:8545"
//host = "http://127.0.0.1:8547"
host         ="http://47.92.137.120:30100"
//host         = "http://10.0.0.110:8545"
sendDuration  = time.Minute * 60000000
nonceTicker   = time.Minute * 10 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
sleepDuration = time.Minute * 1  //查完nonce后休眠时间（1处）
txNum         = 41
)

var privKeys = []string{
//"55bfbf57835787ba3cb37e5396c22ec71343acd5393ade23b2b38040c8e21747",
//"a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa",
//"5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15",
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
//"423e199b057c6206d85081a2ea2b72563e57d36c625c4187f30dfc1763798b03",
//"73d045af8b9892c54aad8286306d124b29f3da4022ff556930cb83d431404592",
//"0ed3332b016e4bbf8d4834399a9853983049b63b246aeb1fd6fdb690791c38f4",
//"4e393c8f2d2005c43f25e2b812963008e1d6f9e73c0e2008744841d87f7fd2bc",
//"2629938c775ab2fc0da7a00c0c0e7eb2300a77fe5bac33cc0d242ad291f4b657",
//"0c961cc45471597714359dfa9cbe74b1296ed434e92e050d58c2ecb82245c394",
//"284f4389d08a9f2a5cfdc2542521e17b03ea9043977791e5abc528307e93a723",
//"15375c54577d5bf690ac67bde21d9ed65b5690ece54078051acee700581c0228",
//"385896a8f68d2ec40afb822bbedf3594d20aa8f29b5e281759a15b30eb1f244d",
//"244435c591ca08aad2cf08a94cc45ba9687cd39dc2d8214282b56c41efa34492",
//"3027b5c4091068562c3e84c9a4bb96e33cd79f047abcc237e47af25516758102",
//"7f1f328c55b8973c07c66558ba0dc60a8eac9c658713bd9cb3573d668bd12f24",
//"7d86677bb13ec09bbd8f8b7d9e4f94cb3f3b948784eeb2a5d2625cf130afad3d",
//"68cfded92d20d5cae0f43a0b147fbfe7f9a0fd4bd347b845a4587c24a1ab3928",
//"6755e28172afc0b043172684d7c3dfeda1f46fab17a0883d697be6f90e5e605d",
//"74b6769a0d92f8a4bb45f314339ee5cc6baec10a3d3425a18a189578ac79444b",
//"21daf42f33d5b282bc52ad0001d8ce6898da997e579f2c233f2ffd2f1d08829b",
//"2b44621ce97f0f3e0759219178b5e6cc890519b7dc0c80c40d161a827fccf245",
//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381",
"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
}

var add = []string{
"0xde587cff6c1137c4eb992cf8149ecff3e163ee07",
"0x1ceb7edecea8d481aa315b9a51b65c4def9b3dc6",
"0x89293fd490355c515aa5df19215b7f3b48e697ec",
"0x5de0589274363cb4f80029db33f8e5c70e762e95",
"0xe8d2ee92a42d494b9574d3d856647d2995623aea",
"0x6997e5b29d3d5466b6c9ad3d2b79c83cecda6375",
"0x806951812847172c834c5d5c6e22862ebeee18fa",
"0xbddf6dae1363d3950c6a1b4e74f6f8d358c831d6",
"0xd94cb8bd4bd0ffb7c1ec4ce1b76ccd15d3eee136",
"0xfa962629c97f6e4da9ac7de75b91ffb566432dff",
"0x37dedd34e52e19681937fe5ee854e5ab1ae2b490",
"0xe3b14eed02502156a53bf3985fec15ca44b1a859",
"0xb05b62e185b3508e315b7fde5ae68c704d6d5ab3",
"0x149502da15637bf0652459243caa6a7dfaad88bc",
"0x022c246d04809730dfacd6b86c4f3fce39aae56b",
"0x4f16914675888dbeed14d81a330698b20b126fdb",
"0xf75179cad6230dfdcc89a909fc1635397c91784e",
"0xe2e5eab9cf073790d924ab82b34dcf938a287e10",
"0x491c69fda8319f8e7cc77edca2525a7322f7e492",
"0xd255ff0136cb2255a8996bdb8208cdc50486ca60",
"0x3ae403f36ac8a3b336f2e931be03077b0ec4f935",
"0x5fcc67d0d737900db68dd62f029b5c62759bb35c",
"0x1170dab93a4612b7e56b0b541d7d203f7d239fa8",
"0xa61191d95cd8c457c8dec274065a4be052ed1df0",
"0xc4b8c2737d8e6b522a9d6f36c2cccb26c42ce00c",
"0x95454465f639dc9640c9a39a45ec3b031cfa890d",
"0x2deee1ac815d3f81a969cbfc4a087bc929e43805",
"0xf5ea3b72e56822b1d638b5f6d6e8666da0b0a88f",
"0xebbcb59c19f4a2e6c12d8055883f256493f2a62d",
"0x824a1eb49226f853943b83c7875d1323f45b18b6",
"0xa3d61c23d456df12b8007ce6fdcbafe0e12044b2",
"0x69633dbb9ebc560ca28f94e18d743f774b8ef01c",
"0xc7496966fcd7dd5375c288d524c88a40b944de66",
"0xaa1a3cbec641330c4d832a584668ac2e20008f95",
"0x28b60a979b79ec34711b6cee23b131781cb232d7",
"0x8447995a4478065329701eeb00d8a5615327110e",
"0x202aeda58992f9ecd897340011f38ae752f69cab",
"0x2709eb3dc15f2b46e4012bc0bff5ab2606261202",
"0x14e1c6f458a67b4d1eafd5fb5cd9704b8b3ad283",
"0x5b5c21690b903d3335d388368772ed7baef54725",
"0x60c045c1e49e364a40405e0818243e0fcbc2c6ba",
}

func main() {
log.SetFlags(log.Lmicroseconds)

var wg sync.WaitGroup

//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
for i, privKey := range privKeys {
log.Printf("i is = %d", i)
wg.Add(1)

go func(j int, key string) {
defer wg.Done()

sendTestTx(
key,
//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
strconv.Itoa(j)+":")
}(i, privKey)
}

wg.Wait()

}

func sendTestTx(privKey, flag string) {
if client, err := eth.Connect(host); err != nil {
fmt.Printf(err.Error())
return
} else {
//addr: 0xa2b67b7e4489549b855e423068cbded55f7c5daa
//r := rand.Intn(len(privKeys))
if privKey, err := crypto.HexToECDSA(privKey); err != nil {
fmt.Printf(err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf("from:%s\n", from.String())
//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e

if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("nonce err: %s", err.Error())
//return
} else {
amount := big.NewInt(0).Mul(big.NewInt(20), big.NewInt(1e18))
//amount := big.NewInt(7)
gasLimit := uint64(21000)
gasPrice := big.NewInt(24000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

timer := time.NewTimer(sendDuration)
ticker := time.NewTicker(nonceTicker)
log.Println(flag+"start:", time.Now().String())
i := 0

for  j := 0; j < len(add); j++{
var to common.Address
to = common.HexToAddress(add[j])
fmt.Printf("to:%s\n", to.String())

select {
case <-timer.C:
log.Println(flag + "time is over")
log.Println(flag+"end:", time.Now().String())
return
case <-ticker.C:
fmt.Println("sleep.........")
time.Sleep(sleepDuration)
fmt.Println("get nonce again")
if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("nonce again err: %s", err.Error())
return
}

fmt.Println("nonce again is", nonce)

default:
//to := common.HexToAddress("0x08b299d855734914cd7b19eea60dc84b825680f9")

fmt.Println(flag+"nonce is what", ":", nonce, "from:", from.String())
tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
//signer := types.HomesteadSigner{}
//signer:=types.NewEIP155Signer(big.NewInt(1))
signer := types.NewEIP155Signer(big.NewInt(739))
signedTx, _ := types.SignTx(tx, signer, privKey)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
fmt.Println(flag+"send raw transaction err:", err.Error())
return
} else {
fmt.Printf(flag+"Transaction hash: %s, %d, %s\n", txHash.String(), nonce, from.String())
nonce++
i++

if txNum != -1 && i >= txNum {
return
}
}
}
}
}
}
}
}
