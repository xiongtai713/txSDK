package main


import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/crypto/sha3"
"github.com/golang/protobuf/proto"
"go-eth/eth"
"go-eth/eth/protos"
"math/big"
)

//0xce9d6b7ce0bcef24fca92ff330a759300435c12b801c753317db44760378af7b
func main() {
//if client, err := eth.Connect("http://10.0.0.155:8545"); err != nil {
if client, err := eth.Connect("http://39.100.92.41:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
// generate private key
// privKey, err := crypto.GenerateKey()
// sha3 helloeth
//if privKey, err := crypto.HexToECDSA("ebc88c101457909fa3496bc02fb707a5a366e4a2e0a3e88de9434d2b45ab292f"); err != nil {
//if privKey, err := crypto.HexToECDSA("b7c41d9f7b87df2a438881ce59120edaf90f77d57254134c9fac724f5e293735"); err != nil {
if privKey, err := crypto.HexToECDSA("a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa"); err != nil {
//if privKey, err := crypto.HexToECDSA("55dce718175480ba5e49e1cead7c85fd9e34bd25f2af78c2433ab8cb96293624"); err != nil {
fmt.Printf(err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Println("from:", from.String())
to := cKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.7")
//to := cKeccak256ToAddress("baap-deploy")
fmt.Printf("to:%s\n", to.String())
fmt.Println(from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf(err.Error())
return
} else {
amount := big.NewInt(0)
gasLimit := uint64(4712388)
gasPrice := big.NewInt(20000000000)

invocation := &protos.Invocation{
Fcn:"put",
Args:[][]byte{[]byte("3"),[]byte("2.29")},
Meta:map[string][]byte{"baap-tx-type": []byte("exec")},
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

fmt.Println("nounce:", nonce)
tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
fmt.Printf(err.Error())
} else {
fmt.Printf("Transaction hash: %s\n", txHash.String())
}
}
}
}
}

func cKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}
