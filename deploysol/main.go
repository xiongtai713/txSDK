package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	crypto3 "github.com/ethereum/go-ethereum/crypto"
	"go-eth/eth"
	"log"
	"math/big"
	"time"
)

var flagNo0x = "26cad4db1a82fab9e41bf5c0fbb4937a27b93786a4f27d3b9704805f698d3e65"

var (
	path  = "6060604052341561000f57600080fd5b6040516104c73803806104c783398101604052808051820191905050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508060019080519060200190610081929190610088565b505061012d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100c957805160ff19168380011785556100f7565b828001600101855582156100f7579182015b828111156100f65782518255916020019190600101906100db565b5b5090506101049190610108565b5090565b61012a91905b8082111561012657600081600090555060010161010e565b5090565b90565b61038b8061013c6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806341c0e1b514610053578063a413686214610068578063cfae3217146100c557600080fd5b341561005e57600080fd5b610066610153565b005b341561007357600080fd5b6100c3600480803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919050506101e4565b005b34156100d057600080fd5b6100d86101fe565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156101185780820151818401526020810190506100fd565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156101e2576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b565b80600190805190602001906101fa9291906102a6565b5050565b610206610326565b60018054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561029c5780601f106102715761010080835404028352916020019161029c565b820191906000526020600020905b81548152906001019060200180831161027f57829003601f168201915b5050505050905090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106102e757805160ff1916838001178555610315565b82800160010185558215610315579182015b828111156103145782518255916020019190600101906102f9565b5b509050610322919061033a565b5090565b602060405190810160405280600081525090565b61035c91905b80821115610358576000816000905550600101610340565b5090565b905600a165627a7a723058201ef87c90fca101864bb70ef1ac60d5510f2af93c1374d2c7c4f6ee8d40d9e4ba002900000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000"
	host     = "http://127.0.0.1:8547"
	privKeys = "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"
)


func DeployWasm(privKey string, client *eth.Client, path string,nonce uint64) {

	//sol
	code, err := hex.DecodeString(path)

	fmt.Println(code[:10])
	if err != nil {
		log.Fatal("read fail", err)
		return
	}

	//type Meta map[string][]byte

	//meta := Meta{"name": []byte("tonysu"), "version": []byte("1.2"), "desc": []byte("tony's new contract"), "jwt": []byte("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhayI6IjAzYTY0M2NmOGIwMmI1MDA1YjRlNDhhY2RjOTRhOGY2OTMzNTVhNGViNTEyOTJkNTUwOGI5N2YzZTYxNTllM2RlNiIsImwiOjYwMDAwMDAwMDAwLCJuIjoiZWVmZmZlZnJlcmVkZmZkc3V1ZjJycmZkc21mbGpsanJyIiwiciI6ImQiLCJzIjoxNzM0NzgsInNrIjoiMDI1OTVkNTUzNjk3MzA1Yzc2NzBkZmQ5MjYyOGU1ZmY2ODA4MDMzNTI2NWVkZjgwNGFlYTRlNmU4ZGY1MTEyNDY0In0.NZAOqpZ65LqDP5x0OS6ECGdSUPN2zlMGcsGPf4Ibij7_aQEOECC5HmoX-P7M0rPhjI_ssAntMdZJJhKs1jKC7w")}
	//metaByte, _ := json.Marshal(meta)
	//metaStr := hex.EncodeToString(metaByte)
	//strcode:=flagNo0x+"_"+hex.EncodeToString(code)+"_"+metaStr
	pri, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.Fatal("err", err)
		return
	}
	from := crypto3.PubkeyToAddress(pri.PublicKey)
	fmt.Println("from", from.String())
	nonce, _ = client.EthClient.PendingNonceAt(context.Background(), from)
	if err != nil {
		log.Fatal("err2222", err)
		return
	}

	fmt.Println(time.Now(),"eeeeeeeeeeeeeeee",nonce)



	//msg := ethereum.CallMsg{
	//	From: from,
	//	Data: code, //code =wasm code1 =sol
	//}

	//gas, err := client.EthClient.EstimateGas(context.Background(), msg)
	//if err != nil {
	//	fmt.Println("预估的gas err", err)
	//	return
	//}

	//fmt.Println("预估的gas", gas)
	//for i:=0;i<=5;i++{
	fmt.Println("nonce", nonce)
	tx, err := types.SignTx(
		types.NewContractCreation(
			nonce,
			new(big.Int),
			338689,
			new(big.Int).Mul(big.NewInt(1e9), big.NewInt(18)),
			code),
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

func main() {
	client, err := eth.Connect("http://127.0.0.1:8547")
	//client, err := eth.Connect("http://10.0.0.116:33333")

	if err != nil {
		log.Fatal("ethclient Dial fail", err)
		return
	}
var nonce uint64
	//for{
		DeployWasm(privKeys, client, path,nonce)
		time.Sleep(10*time.Millisecond)
		//nonce++
	//}





}
