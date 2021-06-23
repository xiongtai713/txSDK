package main

import (
"context"
"encoding/json"
"fmt"
"github.com/ethereum/go-ethereum/common"
"github.com/ethereum/go-ethereum/core/types"
"github.com/ethereum/go-ethereum/crypto"
"github.com/ethereum/go-ethereum/crypto/sha3"
"github.com/ethereum/go-ethereum/p2p/discover"
	"go-eth/callsol"
"math/big"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

func main() {
if client, err := eth1.ToolConnect("http://127.0.0.1:8546"); err != nil {
fmt.Printf(err.Error())
return
} else {
// generate private key
// privKey, err := crypto.GenerateKey()
// sha3 helloeth
if privKey, err := crypto.HexToECDSA("a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381"); err != nil {
//if privKey, err := crypto.HexToECDSA("55dce718175480ba5e49e1cead7c85fd9e34bd25f2af78c2433ab8cb96293624"); err != nil {
//if privKey, err := crypto.HexToECDSA("55ea77dc7293e0d8a231654e2757f881a916d7e21592ee30bb0a8840b634ce48"); err != nil {
fmt.Printf(err.Error())
return
} else {
nodeID := discover.PubkeyID(&privKey.PublicKey)
fmt.Printf("nodeID:%s \n", nodeID.String())
from := crypto.PubkeyToAddress(privKey.PublicKey)
fmt.Printf("from:%s\n", from.String())
to := MyKeccak256ToAddress("x-chain-transfer-withdraw")
fmt.Printf("to:%s\n", to.String())
//to := myKeccak256ToAddress("baap-deploy")//todo test duplicated txid
fmt.Println(from.String())
if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
fmt.Printf("ddddd:%s", err.Error())
return
} else {
nonce = 0
//amount := big.NewInt(0).Mul(big.NewInt(1000), big.NewInt(1e18))
amount := big.NewInt(9000000000000000000)
gasLimit := uint64(4712388)
gasPrice := big.NewInt(240000000000)//todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

type XChainTransferWithdraw struct {
DstChainID string `json:"dst_chain_id"`
DstChainOwner common.Address `json:"dst_chain_owner"`
DstUserAddr common.Address `json:"dst_user_addr"`
DstContractAddr string `json:"dst_contract_addr"`
}

payload, err := json.Marshal(XChainTransferWithdraw{
"738",
common.HexToAddress("123"),
common.HexToAddress("251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8"),
"321",
})
if err != nil {
fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
return
}

fmt.Println("nonce is what", ":", nonce)
tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, payload)
//EIP155 signer
//signer := types.NewEIP155Signer(big.NewInt(738))
signer := types.HomesteadSigner{}
signedTx, _ := types.SignTx(tx, signer, privKey)
// client.EthClient.SendTransaction(context.TODO(), signedTx)
if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
fmt.Println("yerror", err.Error())
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

func MyKeccak256ToAddress(ccName string) common.Address {
hash := sha3.NewKeccak256()

var buf []byte
hash.Write([]byte(ccName))
buf = hash.Sum(buf)

fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

return common.BytesToAddress(buf)
}
