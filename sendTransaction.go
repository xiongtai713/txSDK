package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/crypto/sha3"
"go-eth/eth"
"math/big"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

func main() {
if client, err := eth.Connect("http://10.0.0.148:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
// generate private key
// privKey, err := crypto.GenerateKey()
// sha3 helloeth
if privKey, err := crypto.HexToECDSA("a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa"); err != nil {
//if privKey, err := crypto.HexToECDSA("55dce718175480ba5e49e1cead7c85fd9e34bd25f2af78c2433ab8cb96293624"); err != nil {
//if privKey, err := crypto.HexToECDSA("55ea77dc7293e0d8a231654e2757f881a916d7e21592ee30bb0a8840b634ce48"); err != nil {
fmt.Printf(err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf("from:%s\n", from.String())
to := xKeccak256ToAddress("trust-tx-white-list")
fmt.Printf("to:%s\n", to.String())
fmt.Println(from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("err:%s", err.Error())
return
} else {
amount := big.NewInt(9000000000000000000)
//amount := big.NewInt(0)
gasLimit := big.NewInt(4712388)
gasPrice := big.NewInt(20000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

type WhiteList struct {
Nodes []string `json:"nodes"`
Type  int      `json:"type"` //0:add 1:del
}

/*payload, err := json.Marshal(WhiteList{
Nodes: []string{"a"},
Type:  1,
})
if err != nil {
fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
return
}*/

/*payload, err := hexutil.Decode("0xabc123")
if err != nil {
fmt.Printf("hex decode err:%v\n", err)
return
}*/

fmt.Println("nonce is what", ":", nonce)
msg := eth.NewMessage(&from, &to, amount, gasLimit, gasPrice, nil)

if txHash, err := client.SendTransaction(context.TODO(), &msg); err != nil {
fmt.Println("send transaction error:", err.Error())
} else {
fmt.Printf("Transaction hash: %s\n", txHash.String())

receiptChan := make(chan *types.Receipt)
fmt.Printf("Transaction hash: %s\n", txHash.String())
_, isPending, _ := client.EthClient.TransactionByHash(context.TODO(), txHash)
fmt.Printf("Transaction pending: %v\n", isPending)
// check transaction receipt
client.CheckTransaction(context.TODO(), receiptChan, txHash, 1)
receipt := <-receiptChan
fmt.Printf("Transaction status: %v\n", receipt.Status)
}
}
}
}
}

func xKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}
