package ewasm

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	"go-eth/eth"
	"log"
	"math/big"
)

func PutWasm(privKey string, client *eth.Client, to common.Address) {
	pri, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal("err", err)
		return
	}
	from := crypto2.PubkeyToAddress(pri.PublicKey)
	fmt.Println("from", from.String())
	nonce, _ := client.EthClient.PendingNonceAt(context.Background(), from)
	tx, err := types.SignTx(
		types.NewTransaction(
			nonce,
			to,
			big.NewInt(0),
			90000000,
			new(big.Int).Mul(big.NewInt(1e9), big.NewInt(18)),
			[]byte("put:pdx,222")),
		types.NewEIP155Signer(big.NewInt(739)),
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
