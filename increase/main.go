package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	client2 "pdx-chain/utopia/utils/client"
)

func main() {

	//client, err := eth1.Connect("http://127.0.0.1:8547")
	client, err := client2.Connect("http://127.0.0.1:8547")


	gasLimit := uint64(4712388)
	gasPrice := big.NewInt(240000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

	//privKey, err := crypto.HexToECDSA("a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381")
	//privKey, err := crypto.HetopiaConsortiumFlag xToECDSA("141ebc1d272e88789c2b1eedee3ecb243c95ff6ae0fc8c41658270636cf930c8")

	privKey, err := crypto.HexToECDSA("d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	from := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("from:%s\n", from.String())
	ctx := context.TODO()
	nonce, err := client.EthClient.NonceAt(ctx, from, nil)
	if err != nil {
		fmt.Println("nonce err", err)
		return
	}
	var val int64 = 8888888
	payload1, _ := json.Marshal(val)
	to := common.BytesToAddress(crypto.Keccak256([]byte("inCrease"))[12:])
	tx := types.NewTransaction(nonce, to, big.NewInt(0), gasLimit, gasPrice, payload1)
	//EIP155 signer
	signer := types.NewEIP155Signer(big.NewInt(111))
	//signer := types.HomesteadSigner{}
	signedTx, _ := types.SignTx(tx, signer, privKey)
	if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("增发成功", "txhash", txHash.String(),"to",to.String())
	}
}
