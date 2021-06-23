package main

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"log"
	"math/big"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/pdxcc/util"
	client2 "pdx-chain/utopia/utils/client"
	"strconv"
	"sync"
	"time"
)

func CreatKey(key1, value, jwt string) {
	log.SetFlags(log.Lmicroseconds)

	var wg sync.WaitGroup

	//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
	for i, privKey := range privKeys {
		log.Printf("i is = %d", i)
		wg.Add(1)

		go func(j int, key string) {
			defer wg.Done()

			sendTestTxKey(key, strconv.Itoa(j)+":", key1, value, jwt, j)
		}(i, privKey)
	}

	wg.Wait()

}

func sendTestTxKey(privKey, flag, key, value, jwt string, x int, ) {
	//proxy := "http://10.0.0.241:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1ODk5NzQ2NzIsIkZSRUUiOiJUUlVFIn0.sWYZ6awd8yRNX9iG5o7Ls4Uop5nfZrUtuprx9hwKxw2fS5zQtxunY11bccJ_h29VfnFMqyvaVvI9Tu3R0USlwQ"
	//if client, err := eth1.Connect("http://utopia-chain-1001:8545", proxy, token); err != nil {
	//if client, err := client2.Connect("http://127.0.0.1:8547"); err != nil {
	if client, err := client2.Connect(host1); err != nil {

		//	if client, err := eth1.Connect("http://10.0.0.219:33333"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {

		pKey, err := crypto.HexToECDSA(privKey)
		if err != nil {
			return
		}
		from := crypto.PubkeyToAddress(pKey.PublicKey)
		fmt.Printf("from:%s\n", from.String())
		//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e
		if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {

			//if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
			fmt.Printf("nonce err: %s", err.Error())
			//return
		} else {
			fmt.Println("nonce", nonce)

			amount := big.NewInt(0).Mul(big.NewInt(100), big.NewInt(1e18))

			//amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18))
			gasLimit := uint64(2100000)
			gasPrice := new(big.Int).Mul(big.NewInt(1e9), big.NewInt(1000)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

			timer := time.NewTimer(sendDuration)
			ticker := time.NewTicker(nonceTicker)
			log.Println(flag+"start:", time.Now().String())
			i := 0
			//nonce = 235474
			for {
				select {
				case <-timer.C:
					log.Println(flag + "time is over")
					log.Println(flag+"end:", time.Now().String())
					return
				case <-ticker.C:
					fmt.Println("sleep.........")
					time.Sleep(sleepDuration)
					fmt.Println("get nonce again")
					if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
						fmt.Printf("nonce again err: %s", err.Error())
						return
					}

					fmt.Println("nonce again is", nonce)

				default:

					to := util.EthAddress("PDXSafe")
					var data []byte
					ki := &protos.KeyItem{}

					ki.Key = key
					ki.Value = []byte(value)

					dibyte, err := proto.Marshal(ki)
					if err != nil {
						fmt.Println("proto.Marshal1", err)
						return
					}
					invocation := &protos.Invocation{}

					invocation.Args = make([][]byte, 1)
					invocation.Args[0] = dibyte
					invocation.Meta = make(map[string][]byte)
					invocation.Meta[ACTION] = PDXS_CREATE_KEY
					invocation.Meta[JWT] = []byte("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJhIjoiQyIsImFrIjoiMDI3MmQ5MTBkY2RhOWQ1NWRjMDUwNDQ1OTg1MmMzMmJmYmM4NmUzZjliYWRkZjUxMmY5MGE2ODExNDUyYTU1MGJhIiwiZCI6ImxpdSIsImYiOjE2MTg5ODg2ODcsImtleSI6IjEyMyIsIm4iOiIxIiwicyI6MTIzLCJzayI6IjAzZDE3Mzg5MDEyMmJhMzA2NmIxZDIwYzE4NWM2NTU1NTg2MWI3ZGUzMDA2NmM2ZGM4NTRjZjZjYjkwM2Q2YjY0MiIsInQiOjI2MTg5ODg2ODd9.UVaFx8cYy4BkcYtfPgVxI_-IraIm0raoEy3vHNrz3tHCsC_aoGAoLkAzsoXdVGO6xmg-EtXHyVsHp-xx6kTMqw")
					invocationbyte, err := proto.Marshal(invocation)
					if err != nil {
						fmt.Println("proto.Marshal1", err)
						return
					}
					txp := &protos.Transaction{}
					txp.Payload = invocationbyte
					txp.Type = 1
					txp1, err := proto.Marshal(txp)
					if err != nil {
						fmt.Println("proto.Marshal2", err)
						return
					}
					data = txp1
					tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
					//tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, nil)

					//signer := types.HomesteadSigner{}
					//signer := types.NewSm2Signer(big.NewInt(111))
					signer := types.NewEIP155Signer(big.NewInt(777))

					//sm := (*ecdsa.PrivateKey)(pKey1)
					//signedTx, _ := types.SignTx(tx, signer, sm)
					signedTx, err := types.SignTx(tx, signer, pKey)

					if err != nil {
						fmt.Println("types.SignTx", err)
						return
					}
					if txhash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
						fmt.Println(flag+"send raw transaction err:", err.Error())
						nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						return
					} else {
						fmt.Printf("Transaction hash: %s, %d, %s\n", txhash.String(), nonce, from.String())
						nonce++
						//nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						i++
						//time.Sleep(1*time.Second)
						if txNum != -1 && i >= txNum {
							return
						}
					}
				}
			}
		}
	}
}
