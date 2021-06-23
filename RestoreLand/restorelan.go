package main

import (
	"context"
	"encoding/json"
	"fmt"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	client2 "pdx-chain/utopia/utils/client"

	"math/big"
	"time"
)

func main() {

	//proxy := "http://test.blockfree.pdx.ltd"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1NzQ4NDQwODcsIkZSRUUiOiJUUlVFIn0.lVPyP9xJ2iurl1_-UvdGXFWBGP65qS-NSqN5giUOpkZBvaQ8LO-X5MIP3-A1aUoV-0E3x9ucnMU6YCUUH89PCQ"
	//client, err := eth1.ToolConnect("http://utopia-chain-739:8545", proxy, token)
	client, err := client2.Connect("http://39.100.210.205:30100")
	//client, err := eth1.ToolConnect("http://47.92.137.120:30402")
	//client, err := eth1.ToolConnect("http://39.98.67.5:30402")

	//if err != nil {
	//	fmt.Printf("1", err.Error())
	//	return
	//}
	gasLimit := uint64(4712388)
	gasPrice := new(big.Int)
	gasPrice.Mul(big.NewInt(4000), big.NewInt(1e9)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

	//只有这个账户可以发恢复大陆合约
	//de2f935c24ba2af4fa37c7794b944803d3c44992c04330bd4b80631ea24397c6
	privKey, err := crypto.HexToECDSA("88d443480e5f6954cfcb394da4b8f1575ddfa97d893e4decefabbe0364fdec27") //iass 的收款账户

	//privKey, err := crypto.HexToECDSA("d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84")  //860
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	from := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("from:%s\n", from.String())
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	nonce, err := client.EthClient.NonceAt(ctx, from, nil)
	if err != nil {
		fmt.Println("nonce err", err)
		return
	}
	amount := big.NewInt(0)
	to := common.BytesToAddress(crypto.Keccak256([]byte("QuorumReset"))[12:])
	fmt.Printf("to:%s\n", to.String())
	fmt.Println(from.String())
	vote, _ := client.EthClient.NonceAt(ctx, to, nil)
	if vote == 0 {
		vote = 1
	}
	fmt.Println("查询的合约账户的轮数是", "vote", vote)
	bytes, _ := json.Marshal(vote)
	tx := types.NewTransaction(uint64(nonce), to, amount, gasLimit, gasPrice, bytes)
	//EIP155 signer
	signer := types.NewEIP155Signer(big.NewInt(1001))
	//signer := types.HomesteadSigner{}
	signedTx, _ := types.SignTx(tx, signer, privKey)
	// client.EthClient.SendTransaction(context.TODO(), signedTx)
	if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
		fmt.Println("yerror", err.Error())
		return
	} else {
		fmt.Println("Transaction hash:", txHash.String(), "nonce", nonce)
		nonce++

	}
}
