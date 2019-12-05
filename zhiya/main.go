package main

import (
"context"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/p2p/discover"
"github.com/ethereum/go-ethereum/rlp"
"go-eth/eth"
"math/big"
)

func main() {
if client, err := eth.Connect("http://10.0.0.124:8547"); err != nil {
fmt.Printf(err.Error())
return
} else {

/*
  39.98.63.34 a52a3b32c752a7665226ce88d31b1961302c574bf1e5583374ca93a90931d591
  47.92.164.38 e3aca2d4daf6321463513426e5ca22bb98831593f1c8b9811f4441515c7242e5
  39.98.89.43 14d6a7f2125af91d34b93ff3c4d664e77c009247a5f99a9b89d8f58c5c9740f4
  39.100.92.41 8ad7069a6438c819e637811be8e28476ff756dc7d214d307f116e9af40afb064
*/

if privKey, err := crypto.HexToECDSA("1a1e2f5d3b55bc1d96909482758478fa0bec6ca5696126f80c2208068c9bc498"); err != nil {
fmt.Printf(err.Error())
return
} else {
nodeID := discover.PubkeyID(&privKey.PublicKey)
fmt.Printf("nodeID:%s \n", nodeID.String())
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf("from:%s\n", from.String())

//质押
//to := common.BytesToAddress(crypto.Keccak256([]byte("hypothecation"))[12:])

//退钱
to := common.BytesToAddress(crypto.Keccak256([]byte("redamption"))[12:])

fmt.Printf("to:%s\n", to.String())
fmt.Println(from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("ddddd:%s", err.Error())
return
} else {
amount := big.NewInt(3000000000000000000)
gasLimit := uint64(4712388)
gasPrice := big.NewInt(240000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

//退钱需要填充的
type RecaptionInfo struct {
RecaptionAddress common.Address `json:"recaption_address"`
Withdrawal       bool           `json:"withdrawal"` //是否是主动退出
}

recaption := &RecaptionInfo{
from,
true,
}
payload, err := rlp.EncodeToBytes(recaption)

if err != nil {
fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
return
}

tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, payload)
//EIP155 signer
signer := types.NewEIP155Signer(big.NewInt(738))
//signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
// client.EthClient.SendTransaction(context.TODO(), signedTx)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
fmt.Println("yerror", err.Error())
} else {
fmt.Printf("Transaction hash: %s\n", txHash.String())

}
}
}
}
}
