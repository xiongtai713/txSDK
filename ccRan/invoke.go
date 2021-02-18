package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/crypto/sha3"
	"pdx-chain/pdxcc/protos"
	client2 "pdx-chain/utopia/utils/client"
)

//0xce9d6b7ce0bcef24fca92ff330a759300435c12b801c753317db44760378af7b
func main() {
	//if client, err := client2.Connect("http://10.0.0.211:8545"); err != nil {
	if client, err := client2.Connect("http://127.0.0.1:8547"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		// generate private key
		// privKey, err := crypto.GenerateKey()
		// sha3 helloeth
		//if privKey, err := crypto.HexToECDSA("ebc88c101457909fa3496bc02fb707a5a366e4a2e0a3e88de9434d2b45ab292f"); err != nil {
		//if privKey, err := crypto.HexToECDSA("b7c41d9f7b87df2a438881ce59120edaf90f77d57254134c9fac724f5e293735"); err != nil {
		if privKey, err := crypto.HexToECDSA("d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"); err != nil {
			//if privKey, err := crypto.HexToECDSA("55dce718175480ba5e49e1cead7c85fd9e34bd25f2af78c2433ab8cb96293624"); err != nil {
			fmt.Printf(err.Error())
			return
		} else {
			from := crypto.PubkeyToAddress(privKey.PublicKey)
			fmt.Println("from:", from.String())
			//to := cKeccak256ToAddress("88705146a11b741b960b9a6343487d6974d99e91:mycc")
			to:= common.HexToAddress("0x14632e85D8Cb91D943BD63e08E15DA8411DF2382")
			//to := cKeccak256ToAddress("baap-deploy")
			fmt.Printf("to:%s\n", to.String())
			fmt.Println(from.String())
			if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
				fmt.Printf(err.Error())
				return
			} else {
				//amount := big.NewInt(0)
				gasLimit := uint64(4712388)
				gasPrice := big.NewInt(20000000000)
				var a []byte
				for i:=0;i<10;i++{
					a=append(a,[]byte("a")...)
				}
				invocation := &protos.Invocation{
					Fcn:  "put",
					Args: [][]byte{[]byte("a"), a},
					Meta: map[string][]byte{"baap-tx-type": []byte("exec")},
				}


				payload, err := proto.Marshal(invocation)
				if err != nil {
					fmt.Printf("proto marshal invocation error:%v", err)
				}

				ptx := &protos.Transaction{
					Type:    1, //1invoke 2deploy
					Payload: payload,
				}
				data, err := proto.Marshal(ptx)
				if err != nil {
					fmt.Printf("!!!!!!!!proto marshal error:%v", err)
					return
				}
				//nonce=10
				fmt.Println("nounce:", nonce)

				for i:=0;i<1;i++{
					tx := types.NewTransaction(nonce, to, nil, gasLimit, gasPrice, data)
					signer := types.NewEIP155Signer(big.NewInt(111))

					signedTx, _ := types.SignTx(tx, signer, privKey)
					if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
						fmt.Printf(err.Error())
						return
					} else {
						fmt.Printf("Transaction hash: %s\n", txHash.String())
						nonce++
					}
					//time.Sleep(1*time.Second)
				}

			}
		}
	}
}

func cKeccak256ToAddress(ccName string) common.Address {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write([]byte(ccName))
	buf = hash.Sum(buf)

	fmt.Println("keccak256ToAddress:", common.BytesToAddress(buf).String())

	return common.BytesToAddress(buf)
}
