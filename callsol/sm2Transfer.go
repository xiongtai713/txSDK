package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/tjfoc/gmsm/sm2"
	"log"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	client2 "pdx-chain/utopia/utils/client"
	"strconv"
	"sync"
	"time"
)

const (
	//host         = "http://39.100.34.235:30074"
	//host         = "http://39.100.93.177:30036"
	//host = "http://10.0.0.203:33333"
	//host3 = "http://10.0.0.211:22222"
	host1sm2          = "http://127.0.0.1:8547"
	sendDuration1  = time.Minute * 1
	nonceTicker1   = time.Minute * 10 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
	sleepDuration1 = time.Minute * 1  //查完nonce后休眠时间（1处）
	txNum1         = -1
)

var privKeys1 = []string{
		//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381",  //7DE
		//"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
		//"009f1dfe52be1015970d9087de0ad2a98f4c68f610711d1533aa21a71ccc8f4a", //from:0x00CFc66BBD69fb964df1C9782062D4282FfF0cda
		//"69192206e447dbc8b6627d7beb540e6c606c5b94afa9ebc00734ff404a1e5617",
		//
		//"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84", //086
		"72660cbaef2ca607751c0514922ca995a566f5bd508ccfae4896265db856d115", //sm2
		//"65035d9621f7be3bb6dc1f5a646e6ee2ef6bddf3f1ce57782d409c23857401a6", //sm2 不能过
		//"9f8bc5114579e4db3765771a70869678caa13a9ec3571a3cf2a6140d375b7417", //sm2 能过
		//"71fa69bf38e20b32fbf980645eee0496dd13c85dceb4b3e2c66514ceed27f40e",
}

func main() {
	log.SetFlags(log.Lmicroseconds)

	var wg sync.WaitGroup

	//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
	for i, privKey := range privKeys1 {
		log.Printf("i is = %d", i)
		wg.Add(1)

		go func(j int, key string) {
			defer wg.Done()

			sendTestTx1(
				key,
				//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
				strconv.Itoa(j)+":")
		}(i, privKey)
	}

	wg.Wait()

}

func sendTestTx1(privKey, flag string) {
	//proxy := "http://10.0.0.241:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1ODk5NzQ2NzIsIkZSRUUiOiJUUlVFIn0.sWYZ6awd8yRNX9iG5o7Ls4Uop5nfZrUtuprx9hwKxw2fS5zQtxunY11bccJ_h29VfnFMqyvaVvI9Tu3R0USlwQ"
	//if client, err := eth.Connect("http://utopia-chain-1001:8545", proxy, token); err != nil {
	if client, err := client2.Connect(host1sm2); err != nil {
		//	if client, err := eth.Connect("http://10.0.0.219:33333"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		//addr: 0xa2b67b7e4489549b855e423068cbded55f7c5daa
		//r := rand.Intn(len(privKeys))

		//pKey, err := crypto.HexToECDSA(privKey)
		//if err != nil {
		//	return
		//}

		sbin, ok := new(big.Int).SetString(privKey, 16)
		if !ok {
			return
		}
		sKey := sm2.InitKey(sbin)

		//pKey1, _ := crypto.HexToECDSA(privKey)
		//pbin, _ := new(big.Int).SetString(privKey, 16)
		//pKey1 := sm2.InitKey(pbin)

		from := crypto.PubkeyToAddress(sKey.PublicKey)
		fmt.Printf("from:%s\n", from.String())
		//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e
		//compress := sm2.Compress((*sm2.PublicKey)(&sKey.PublicKey))

		pub := crypto.FromECDSAPub(&sKey.PublicKey)
		fmt.Println("公钥",fmt.Sprintf("%x",pub))
		if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
			fmt.Printf("nonce err: %s", err.Error())
			//return
		} else {
			fmt.Println("nonce", nonce)

			//amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18))

			amount := big.NewInt(0).Mul(big.NewInt(0), big.NewInt(1e18))
			gasLimit := uint64(22000)
			gasPrice := new(big.Int).Mul(big.NewInt(1e9), big.NewInt(4000)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

			timer := time.NewTimer(sendDuration1)
			ticker := time.NewTicker(nonceTicker1)
			log.Println(flag+"start:", time.Now().String())
			i := 0
			//nonce = 100
			for {
				select {
				case <-timer.C:
					log.Println(flag + "time is over")
					log.Println(flag+"end:", time.Now().String())
					return
				case <-ticker.C:
					fmt.Println("sleep.........")
					time.Sleep(sleepDuration1)
					fmt.Println("get nonce again")
					if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
						fmt.Printf("nonce again err: %s", err.Error())
						return
					}

					fmt.Println("nonce again is", nonce)

				default:
					//time.Sleep(5*time.Millisecond)
					//fmt.Println(flag+"nonce is what", ":", nonce, "from:", from.String())
					//rand.Seed(time.Now().Unix())
					//n1 := rand.Int31n(9)
					//n2 := rand.Int31n(9)
					to := common.HexToAddress(fmt.Sprintf("0x08b299d855734914cd7b19eea60dc84b%d25680f%d", i, i))
					//to := common.HexToAddress("0x3A5eDBF73d1A088Ab359D28Ab99B6253453A2314")
					//fmt.Printf("to:%s\n", to.String())
					//data:=[]byte(string(i))
					tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)
					//tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, nil)

					//signer := types.HomesteadSigner{}
					signer := types.NewSm2Signer(big.NewInt(777))
					//signer := types.NewEIP155Signer(big.NewInt(777))
					sm := (*ecdsa.PrivateKey)(sKey)
					signedTx, _ := types.SignTx(tx, signer, sm)
					//compress := sm2.Compress((*sm2.PublicKey)(&sKey.PublicKey))
					//fmt.Println("公钥",hex.EncodeToString(compress))
					//signedTx, err := SignTx(tx, signer, pKey)
					//data, err := rlp.EncodeToBytes(signedTx)
					//fmt.Println(common.ToHex(data))

					//sm2.Sm2Verify(sKey.PublicKey,signer.Hash(tx).Bytes(),signer.ChainId.Bytes(),signedTx.Data())
					if err != nil {
						fmt.Println("types.SignTx", err)
						return
					}
					if _, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
						fmt.Println(flag+"send raw transaction err:", err.Error())
						nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						return
					} else {
						//fmt.Printf(flag+"Transaction hash: %s, %d, %s\n", txhash.String(), nonce, from.String())
						nonce++
						i++

						if txNum1 != -1 && i >= txNum1 {
							return
						}
					}
				}
			}
		}
	}
}





