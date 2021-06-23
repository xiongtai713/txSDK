package ewasm

import (
	"context"
	"fmt"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	crypto2 "pdx-chain/crypto"
	"pdx-chain/utopia/utils/client"

	"log"
	"math/big"
)

func PutWasm(privKey string, client *client.Client, to common.Address) {
	pri, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal("err", err)
		return
	}
	from := crypto2.PubkeyToAddress(pri.PublicKey)
	fmt.Println("from", from.String())
	nonce, _ := client.EthClient.PendingNonceAt(context.Background(), from)
	//msg := ethereum.CallMsg{
	//	From: from,
	//	To:   &to,
	//	Data: []byte("put:pdx,222"), //code =wasm code1 =sol
	//}
	//
	//gas, err := client.EthClient.EstimateGas(context.Background(), msg)
	//if err != nil {
	//	fmt.Println("预估的gas err", err)
	//	return
	//}

	//fmt.Println("预估的gas", gas)
	tx, err := types.SignTx(
		types.NewTransaction(
			nonce,
			to,
			big.NewInt(0),
			40000000,
			new(big.Int).Mul(big.NewInt(1e9), big.NewInt(18)),
			[]byte("put:pdx,222")),
		types.NewEIP155Signer(big.NewInt(111)),
		pri,
	)
	if err != nil {
		log.Fatal("tx err", err)
	}

	hashes, err := client.SendRawTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("tx err", err)
	}
	fmt.Println("txHash", hashes.String())
}
