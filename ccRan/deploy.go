package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/golang/protobuf/proto"
	"go-eth/eth"
	"go-eth/eth/protos"
	"io/ioutil"
	"math/big"
)

const (
	owner   = "8000d109DAef5C81799bC01D4d82B0589dEEDb33"
	name    = "testcc02"
	version = "1.0.8"
)

//0xce9d6b7ce0bcef24fca92ff330a759300435c12b801c753317db44760378af7b
func main() {
	//if client, err := eth.Connect("http://10.0.0.155:8545"); err != nil {
	//if client, err := eth.Connect("http://10.0.0.110:8545"); err != nil {
	if client, err := eth.Connect("http://39.100.92.41:8545"); err != nil {
		//if client, err := eth.Connect("http://39.100.84.247:8545"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		// generate private key
		// privKey, err := crypto.GenerateKey()
		// sha3 helloeth
		//if privKey, err := crypto.HexToECDSA("ebc88c101457909fa3496bc02fb707a5a366e4a2e0a3e88de9434d2b45ab292f"); err != nil {
		//if privKey, err := crypto.HexToECDSA("b7c41d9f7b87df2a438881ce59120edaf90f77d57254134c9fac724f5e293735"); err != nil {
		if privKey, err := crypto.HexToECDSA("5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15"); err != nil {
			//if privKey, err := crypto.HexToECDSA("55dce718175480ba5e49e1cead7c85fd9e34bd25f2af78c2433ab8cb96293624"); err != nil {
			fmt.Printf(err.Error())
			return
		} else {
			from := crypto.PubkeyToAddress(privKey.PublicKey)
			fmt.Println("from:", from.String())
			//to := iKeccak256ToAddress(":baap-deploy:v1.0")
			to := iKeccak256ToAddress(":baap-deploy:v1.0")
			//to := iKeccak256ToAddress("tcUpdater")
			fmt.Printf("to:%s\n", to.String())
			fmt.Println(from.String())
			if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
				fmt.Printf(err.Error())
				return
			} else {
				amount := big.NewInt(0)
				gasLimit := uint64(4712388)
				//gasPrice := big.NewInt(2000000000)
				gasPrice := big.NewInt(43100000000000)

				deployInfo := struct {
					FileName    string `json:"fileName"`
					ChaincodeId string `json:"chaincodeId"`
					Pbk         string `json:"pbk"`
				}{
					"MyCc.java",
					owner + ":" + name + ":" + version,
					string(crypto.CompressPubkey(&privKey.PublicKey)),
				}
				deployInfoBuf, err := json.Marshal(deployInfo)
				if err != nil {
					fmt.Printf("marshal deployInfo err: %v", err)
					return
				}

				myccBuf, err := ioutil.ReadFile("/Users/liu/go/src/go-eth/ccRan/MyCc.java")
				if err != nil {
					fmt.Printf("read java file err:%v", err)
					return
				}
				invocation := &protos.Invocation{
					Fcn:  "deploy",
					Args: [][]byte{deployInfoBuf},
					Meta: map[string][]byte{
						"baap-tx-type": []byte("exec"),
						"jwt":          []byte("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhayI6IjAzYTY0M2NmOGIwMmI1MDA1YjRlNDhhY2RjOTRhOGY2OTMzNTVhNGViNTEyOTJkNTUwOGI5N2YzZTYxNTllM2RlNiIsImwiOjYwMDAwMDAwMDAwLCJuIjoiZWVmZmZlZnJlcmVkZmZkc3V1ZjJycmZkc21mbGpsanJyIiwiciI6ImQiLCJzIjoxNzM0NzgsInNrIjoiMDI1OTVkNTUzNjk3MzA1Yzc2NzBkZmQ5MjYyOGU1ZmY2ODA4MDMzNTI2NWVkZjgwNGFlYTRlNmU4ZGY1MTEyNDY0In0.NZAOqpZ65LqDP5x0OS6ECGdSUPN2zlMGcsGPf4Ibij7_aQEOECC5HmoX-P7M0rPhjI_ssAntMdZJJhKs1jKC7w"),
						"baap-cc-code": myccBuf,
					},
				}
				dep := &protos.Deployment{
					Owner:   owner,
					Name:    name,
					Version: version,
					Payload: invocation,
				}
				payload, err := proto.Marshal(dep)
				if err != nil {
					fmt.Printf("proto marshal invocation error:%v", err)
				}

				ptx := &protos.Transaction{
					Type:    2, //1invoke 2deploy
					Payload: payload,
				}
				data, err := proto.Marshal(ptx)
				if err != nil {
					fmt.Printf("!!!!!!!!proto marshal error:%v", err)
					return
				}

				fmt.Println("nounce:", nonce)
				tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
				// EIP155 signer
				// signer := types.NewEIP155Signer(big.NewInt(4))
				signer := types.HomesteadSigner{}
				signedTx, _ := types.SignTx(tx, signer, privKey)
				// client.EthClient.SendTransaction(context.TODO(), signedTx)
				if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
					fmt.Printf("send raw tx:%s", err.Error())
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

func iKeccak256ToAddress(ccName string) common.Address {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write([]byte(ccName))
	buf = hash.Sum(buf)

	fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

	return common.BytesToAddress(buf)
}
