package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go-eth/eth"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

const (
	//host         = "http://39.100.34.235:30074"
	//host         = "http://39.100.93.177:30036"
	//host = "http://127.0.0.1:8547"
	host = "http://39.98.63.34:30402"
	//host         = "http://10.0.0.110:8545"
	sendDuration  = time.Minute * 60000000
	nonceTicker   = time.Minute * 10 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
	sleepDuration = time.Minute * 1  //查完nonce后休眠时间（1处）
	txNum         = 10
)

var privKeys = []string{
	//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381",
    // "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
    "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
}

func main() {
	log.SetFlags(log.Lmicroseconds)

	var wg sync.WaitGroup

	//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
	for i, privKey := range privKeys {
		log.Printf("i is = %d", i)
		wg.Add(1)

		go func(j int, key string) {
			defer wg.Done()

			sendTestTx(
				key,
				//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
				strconv.Itoa(j)+":")
		}(i, privKey)
	}

	wg.Wait()

}

func sendTestTx(privKey, flag string) {
	//proxy := "http://39.100.92.188:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1NzQ2NzM3ODIsIkZSRUUiOiJUUlVFIn0.DARdyLtPRhT9eNPpoOi4KIno3ZC-UTQ2D48yiBdOXkYBjaKjdiggJUVzoVNvTEnRqzaeBP8WizIp_ZMo_Eh_JA"
	//if client, err := eth.Connect("http://utopia-chain-739:8545", proxy, token); err != nil {
	//if client, err := eth.Connect("http://47.92.137.120:30402"); err != nil {
		if client, err := eth.Connect("http://127.0.0.1:8547"); err != nil {

			fmt.Printf(err.Error())
		return
	} else {
		//addr: 0xa2b67b7e4489549b855e423068cbded55f7c5daa
		//r := rand.Intn(len(privKeys))
		if privKey, err := crypto.HexToECDSA(privKey); err != nil {
			fmt.Printf(err.Error())
			return
		} else {
			from := crypto.PubkeyToAddress(privKey.PublicKey)
			fmt.Printf("from:%s\n", from.String())
			//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e

			if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
				fmt.Printf("nonce err: %s", err.Error())
				//return
			} else {
				//amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18))

				amount := big.NewInt(7)
				gasLimit := uint64(21000)
				gasPrice := big.NewInt(100000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

				timer := time.NewTimer(sendDuration)
				ticker := time.NewTicker(nonceTicker)
				log.Println(flag+"start:", time.Now().String())
				i := 0

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
						//time.Sleep(5*time.Millisecond)
						fmt.Println(flag+"nonce is what", ":", nonce, "from:", from.String())
						rand.Seed(time.Now().Unix())
						//n1 := rand.Int31n(9)
						//n2 := rand.Int31n(9)
						//to := common.HexToAddress(fmt.Sprintf("0x08b299d855734914cd7b19eea60dc84b%d25680f%d", n1, n2))
						to := common.HexToAddress("0x2C393b0723177801Ba69425C087D10E8F4C558FF")
						fmt.Printf("to:%s\n", to.String())
						tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
						//signer := types.HomesteadSigner{}
						//signer:=types.NewEIP155Signer(big.NewInt(1))
						signer := types.NewEIP155Signer(big.NewInt(739))
						signedTx, _ := types.SignTx(tx, signer, privKey)
						if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
							fmt.Println(flag+"send raw transaction err:", err.Error())
							nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
							return
						} else {
							fmt.Printf(flag+"Transaction hash: %s, %d, %s\n", txHash.String(), nonce, from.String())
							nonce++
							i++

							if txNum != -1 && i >= txNum {
								return
							}
						}
					}
				}
			}
		}
	}
}
