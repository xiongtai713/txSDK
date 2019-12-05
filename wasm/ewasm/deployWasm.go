package ewasm

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	crypto3 "github.com/ethereum/go-ethereum/crypto"
	"go-eth/eth"
	"io/ioutil"
	"log"
	"math/big"
)

var flagNo0x = "26cad4db1a82fab9e41bf5c0fbb4937a27b93786a4f27d3b9704805f698d3e65"

func IsWASM(code []byte) bool {
	fmt.Println([]byte("\000asm"))
	if len(code) < 4 || string(code[:4]) != "\000asm" {
		return false
	}
	return true
}

func DeployWasm(privKey string, client *eth.Client, path string) {
	code, err := ioutil.ReadFile(path)
	fmt.Println(IsWASM(code))
	if err != nil {
		log.Fatal("read fail", err)
		return
	}
	fmt.Println("3333333", code[:9])

	type Meta map[string][]byte

	//meta := Meta{"name": []byte("tonysu"), "version": []byte("1.2"), "desc": []byte("tony's new contract"), "jwt": []byte("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhayI6IjAzYTY0M2NmOGIwMmI1MDA1YjRlNDhhY2RjOTRhOGY2OTMzNTVhNGViNTEyOTJkNTUwOGI5N2YzZTYxNTllM2RlNiIsImwiOjYwMDAwMDAwMDAwLCJuIjoiZWVmZmZlZnJlcmVkZmZkc3V1ZjJycmZkc21mbGpsanJyIiwiciI6ImQiLCJzIjoxNzM0NzgsInNrIjoiMDI1OTVkNTUzNjk3MzA1Yzc2NzBkZmQ5MjYyOGU1ZmY2ODA4MDMzNTI2NWVkZjgwNGFlYTRlNmU4ZGY1MTEyNDY0In0.NZAOqpZ65LqDP5x0OS6ECGdSUPN2zlMGcsGPf4Ibij7_aQEOECC5HmoX-P7M0rPhjI_ssAntMdZJJhKs1jKC7w")}
	//metaByte, _ := json.Marshal(meta)
	//metaStr := hex.EncodeToString(metaByte)
	//strcode:=flagNo0x+"_"+hex.EncodeToString(code)+"_"+metaStr
	strcode := code
	pri, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal("err", err)
		return
	}
	from := crypto3.PubkeyToAddress(pri.PublicKey)
	fmt.Println("from", from.String())
	nonce, _ := client.EthClient.PendingNonceAt(context.Background(), from)

	//for i:=0;i<=5;i++{
	fmt.Println("nonce", nonce)
	tx, err := types.SignTx(
		types.NewContractCreation(
			nonce,
			new(big.Int),
			99000000,
			new(big.Int).Mul(big.NewInt(1e9), big.NewInt(18)),
			[]byte(strcode)),
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
	fmt.Println("txHash", hashes.Hex())

	fmt.Println("合约地址", crypto.CreateAddress(from, tx.Nonce()).Hex())
	//nonce++
	//}

}
