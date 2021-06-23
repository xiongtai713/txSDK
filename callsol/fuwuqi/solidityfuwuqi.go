package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	ethereum "pdx-chain"
	"pdx-chain/accounts/abi"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	client2 "pdx-chain/utopia/utils/client"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//host         = "http://39.100.34.235:30074"
	//host         = "http://39.100.93.177:30036"
	//host = "http://10.0.0.203:33333"
	//host = "http://10.0.0.4:33333"
	//host = "http://127.0.0.1:8547"

	sendDuration2  = time.Second * 1000000
	nonceTicker2   = time.Minute * 1 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
	sleepDuration2 = time.Minute * 1  //查完nonce后休眠时间（1处）
	txNum2         = -1
)

var privKeys2 = []string{
	"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381", //7DE
	"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
	"009f1dfe52be1015970d9087de0ad2a98f4c68f610711d1533aa21a71ccc8f4a", //from:0x00CFc66BBD69fb964df1C9782062D4282FfF0cda
	"69192206e447dbc8b6627d7beb540e6c606c5b94afa9ebc00734ff404a1e5617",
	"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84", //086
	"72660cbaef2ca607751c0514922ca995a566f5bd508ccfae4896265db856d115", //sm2
	"e7bb5c0bc8456fe0c2af281fe5753095c5150fbdbd622776330cece60e9feaec",
	"5dc4c81695f4606394b535e44315f9a444b7f860f1951f4eb5eb8fdf59913b0a",
	//新超给的私钥
	"de2f935c24ba2af4fa37c7794b944803d3c44992c04330bd4b80631ea24397c6",
	"d8f9556d406e09b2a692e91b849705869df7c8b6cc02d650a0048cd4ad3884ee",
	"f5e166ecd28e2c0afdf7f764e9844588ec1b7a34c77c7be3b673f5a4d9b31f6a",
	"1598e1b4f6b2f911142fe5720554360bdd45c5487ff922e06930513c25c87577",
	"b5df2ce6d414388ca6d425282e116f6910b679d13d576a7d70616d3b7bc9180f",
	"9985bfb012da6de074b2377287723b86704a7cffc1ce3661ddf5fa091b789a72",
	"e3686e780a7ee0c1e02a3f2fae004c3d763199cb4247e9ecbb0117a7718d4bc6",

}
var host1 string

func main() {
	var start = flag.Int("start", 0, "从第几个私钥开始,最大值500")
	var end = flag.Int("end", 1, "到第几个私钥结束,最大值501")
	var ip = flag.String("ip", "127.0.0.1", "目标ip默认值127.0.0.1")
	flag.Parse()
	if *start >= 15 {
		*start = 15
	}
	if *end >= 15 {
		*end = 15
	}

	host1 = "http://" + *ip + ":8547"

	log.SetFlags(log.Lmicroseconds)
	actual := privKeys2[*start:*end]
	var wg sync.WaitGroup
	//最多543
	//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
	for i, privKey := range actual {
		log.Printf("i is = %d", i)
		wg.Add(1)

		go func(j int, key string) {
			defer wg.Done()

			sendTestTx2(
				key,
				//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
				strconv.Itoa(j)+":", j)
		}(i, privKey)
	}

	wg.Wait()

}

func sendTestTx2(privKey, flag string, x int) {
	if client, err := client2.Connect(host1); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		pKey, err := crypto.HexToECDSA(privKey)
		if err != nil {
			return
		}
		from := crypto.PubkeyToAddress(pKey.PublicKey)
		if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {

			fmt.Printf("nonce err: %s", err.Error())
		} else {

			amount := big.NewInt(0).Mul(big.NewInt(0), big.NewInt(1e18))

			//amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18))
			deployGasLimit := uint64(245567)

			gasPrice := new(big.Int).Mul(big.NewInt(1e9), big.NewInt(4000)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

			code, err := hex.DecodeString("608060405234801561001057600080fd5b506102a7806100206000396000f30060806040526004361061004b5763ffffffff7c0100000000000000000000000000000000000000000000000000000000600035041663054c1a758114610050578063841321fd146100da575b600080fd5b34801561005c57600080fd5b50610065610135565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561009f578181015183820152602001610087565b50505050905090810190601f1680156100cc5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b3480156100e657600080fd5b506040805160206004803580820135601f81018490048402850184019095528484526101339436949293602493928401919081908401838280828437509497506101cc9650505050505050565b005b60008054604080516020601f60026000196101006001881615020190951694909404938401819004810282018101909252828152606093909290918301828280156101c15780601f10610196576101008083540402835291602001916101c1565b820191906000526020600020905b8154815290600101906020018083116101a457829003601f168201915b505050505090505b90565b80516101df9060009060208401906101e3565b5050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061022457805160ff1916838001178555610251565b82800160010185558215610251579182015b82811115610251578251825591602001919060010190610236565b5061025d929150610261565b5090565b6101c991905b8082111561025d57600081556001016102675600a165627a7a72305820da4127a6e95f68f97561ea72893a4b4cf035d8a59bf39ad7348a3df781e844ba0029")

			tx, err := types.SignTx(
				types.NewContractCreation(
					nonce,
					new(big.Int),
					deployGasLimit,
					new(big.Int).Mul(big.NewInt(1e9), big.NewInt(18)),
					code),
				types.NewEIP155Signer(big.NewInt(777)),
				//types.NewSm2Signer(big.NewInt(777)),
				pKey,
			)
			if err != nil {
				log.Fatal("tx err", err)
			}

			hashes, err := client.SendRawTransaction(context.Background(), tx)
			if err != nil {
				fmt.Println("tx err", err)
				return
			}
			fmt.Println("txHash", hashes.Hex())

			to := crypto.CreateAddress(from, tx.Nonce())

			time.Sleep(20 * time.Second)
			var data []byte
			data = creatAbi()
			msg := ethereum.CallMsg{
				From: from,
				Data: data, //code =wasm code1 =sol
				To:   &to,
			}

			gas, err := client.EthClient.EstimateGas(context.Background(), msg)
			if err != nil {
				fmt.Println("预估的gas err", err)
				return
			}
			fmt.Println("预估的gas", gas)

			gasLimit := uint64(gas)
			timer := time.NewTimer(sendDuration2)
			ticker := time.NewTicker(nonceTicker2)
			//log.Println(flag+"start:", time.Now().String())
			i := 0
			if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
				return
			}

			//nonce = 2
			for {
				select {
				case <-timer.C:
					//log.Println(flag + "time is over")
					log.Println("发送交易数量", i)
					//log.Println(flag+"end:", time.Now().String(),)
					return
				case <-ticker.C:
					fmt.Println("sleep.........")
					time.Sleep(sleepDuration2)
					fmt.Println("get nonce again")
					if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
						fmt.Printf("nonce again err: %s", err.Error())
						return
					}

					//fmt.Println("nonce again is", nonce)

				default:


					tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)

					signer := types.NewEIP155Signer(big.NewInt(777))


					signedTx, err := types.SignTx(tx, signer, pKey)

					if err != nil {
						fmt.Println("types.SignTx", err)
						return
					}
					if _, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
						fmt.Println(flag+"send raw transaction err:", err.Error())
						nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						return
					} else {
						//fmt.Printf("Transaction hash: %s, %d, %s\n", txhash.String(), nonce, from.String())
						nonce++
						//nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						i++
						//time.Sleep(1*time.Second)
						if txNum2 != -1 && i >= txNum2 {
							return
						}
					}
				}
			}
		}
	}
}

func creatAbi() []byte {

	myContractAbi := `[
	{
		"constant": false,   
		"inputs": [],
		"name": "get1",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{
				"name": "s",
				"type": "string"
			}
		],
		"name": "put1",
		"outputs": [],
		"payable": false,
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
	/*
	    constant 是否改变合约
	   	inputs  方法参数
	    name 方法名称
	    outputs 返回值
	    payable 是否可以转账
	    type 类型 function，constructor，fallback（缺省方法）

	*/

	abi, err := abi.JSON(strings.NewReader(myContractAbi))
	if err != nil {
		log.Fatalln("1", err)
	}

	abiBuf, err := abi.Pack("put1", "111")
	if err != nil {
		log.Fatalln("2", err)
	}
	//fmt.Println("abibuf", fmt.Sprintf("%x", abiBuf))
	return abiBuf
}
