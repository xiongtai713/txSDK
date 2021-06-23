package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go-eth/callsol"
	"math/big"
)

var prikey = string(
	//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381",
	"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84", )

type ConsensusNodeUpdate struct {
	NodeType     int              `json:"nodeType"` //consensus=0 ,observer=1
	FromChainID  uint64           `json:"fromChainID"`
	ToChainID    uint64           `json:"toChainID"`
	Cert         string           `json:"cert"`
	Address      []common.Address `json:"address"`
	CommitHeight uint64           `json:"commitHeight"`
	NuAdmCrt     string           `json:"nuAdmCrt"`
}

func main() {

	client, _ := eth1.ToolConnect("http://127.0.0.1:8547")

	key, _ := crypto.HexToECDSA(prikey)

	from := crypto.PubkeyToAddress(key.PublicKey)

	nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil)
	if err != nil {
		fmt.Println("Nonce err", err)
		return
	}

	to := common.BytesToAddress(crypto.Keccak256([]byte("consensus-node-update"))[12:])

	node := ConsensusNodeUpdate{
		//0x86082FA9d3c14D00a8627aF13CfA893e80B39101 0x7de2a31d6ca36302ea7b7917c4fc5ef4c12913b6
		Address:      []common.Address{common.HexToAddress("0x7de2a31d6ca36302ea7b7917c4fc5ef4c12913b6")},
		CommitHeight: 505,
	}
	data, _ := json.Marshal(node)

	gasPrice := big.NewInt(9000000000)

	//msg := ethereum.CallMsg{
	//	From: from,
	//	To:   &to,
	//	Data: data,
	//}
	//
	//gaslimit, err := client.EthClient.EstimateGas(context.TODO(), msg)
	//if err != nil {
	//	fmt.Println("err", err)
	//	return
	//}
	gaslimit := uint64(21000)
	fmt.Println("查询出来的 gas", "limit", gaslimit)

	tx := types.NewTransaction(nonce, to, big.NewInt(0), gaslimit, gasPrice, data)

	signer := types.NewEIP155Signer(big.NewInt(739))

	tx1, _ := types.SignTx(tx, signer, key)

	hashes, err := client.SendRawTransaction(context.TODO(), tx1)
	if err != nil {
		fmt.Println("交易hash", err)
		return
	}

	fmt.Println("交易hash", "hash", hashes.String(), "nonce", nonce)

}
