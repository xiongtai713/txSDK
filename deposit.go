package main

import (
"context"
"encoding/json"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/crypto/sha3"
  "go-eth/callsol"
"math/big"
)
//0xce9d6b7ce0bcef24fca92ff330a759300435c12b801c753317db44760378af7b
func main() {
//if client, err := eth1.ToolConnect("http://10.0.0.155:8545"); err != nil {
if client, err := eth1.ToolConnect("http://10.0.0.110:8545"); err != nil {
fmt.Printf(err.Error())
return
} else {
// generate private key
// privKey, err := crypto.GenerateKey()
// sha3 helloeth
//if privKey, err := crypto.HexToECDSA("ebc88c101457909fa3496bc02fb707a5a366e4a2e0a3e88de9434d2b45ab292f"); err != nil {
//if privKey, err := crypto.HexToECDSA("b7c41d9f7b87df2a438881ce59120edaf90f77d57254134c9fac724f5e293735"); err != nil {
//if privKey, err := crypto.HexToECDSA("a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa"); err != nil {
if privKey, err := crypto.HexToECDSA("a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa"); err != nil {
fmt.Printf(err.Error())
return
} else {
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Println("from:", from.String())
to := Keccak256ToAddress("x-chain-transfer-deposit")
fmt.Printf("to:%s\n", to.String())
fmt.Println(from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf(err.Error())
return
} else {
amount := big.NewInt(0)
gasLimit := uint64(4712388)
gasPrice := big.NewInt(20000000000)

type XChainTransferDeposit struct {
SrcChainID string `json:"src_chain_id"`
SrcChainOwner common.Address `json:"src_chain_owner"`
//TxHash common.Hash `json:"tx_hash"`
TxMsg string `json:"tx_msg"`
}

//param, err := json.Marshal(XChainTransferWithdraw{"738", common.HexToAddress("123"), "321"})
paload, err := json.Marshal(
XChainTransferDeposit{
"739",
common.HexToAddress("123"),
  //common.HexToHash("0x6d3063fb6f4b76829e2f69a9f736b9270c368bc2d44a0f7c3eb8577e2cdd2b03"),
  "0xf9011a018537e11d60008347e7c49457509fce31f67c5cf5307d1b265481c8803052d6887ce66c50e2840000b8ac7b226473745f636861696e5f6964223a22373338222c226473745f636861696e5f6f776e6572223a22307830303030303030303030303030303030303030303030303030303030303030303030303030313233222c226473745f757365725f61646472223a22307832353162333734306130326131633563663566666364663630643432656432613833393864646338222c226473745f636f6e74726163745f61646472223a22333231227d1ba06abd76d9248ea6d4bb442d484bde43f55efa57f38578dc9511b68f1827f127d2a01182ddfc958b33b6c40b0b750dbfec5003fcdba3f4abfb5950cbc8b8b2a05876",
})
if err != nil {
fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
return
}

fmt.Println("nounce:", nonce)
tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, paload)
// EIP155 signer
// signer := types.NewEIP155Signer(big.NewInt(4))
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
// client.EthClient.SendTransaction(context.TODO(), signedTx)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
fmt.Printf(err.Error())
} else {
fmt.Printf("Transaction hash: %s\n", txHash.String())
/*receiptChan := make(chan *types.Receipt)
fmt.Printf("Transaction hash: %s\n", txHash.String())
_, isPending, _ := client.EthClient.TransactionByHash(context.TODO(), txHash)
fmt.Printf("Transaction pending: %v\n", isPending)
// check transaction receipt
client.CheckTransaction(context.TODO(), receiptChan, txHash, 1)
receipt := <-receiptChan
fmt.Printf("Transaction status: %v\n", receipt.Status)*/
}
}
}
}
}

func Keccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}
