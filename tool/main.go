package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"strings"
	"sync/atomic"
	"time"
)

var (
	//path     = "6060604052341561000f57600080fd5b6040516104c73803806104c783398101604052808051820191905050336000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508060019080519060200190610081929190610088565b505061012d565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100c957805160ff19168380011785556100f7565b828001600101855582156100f7579182015b828111156100f65782518255916020019190600101906100db565b5b5090506101049190610108565b5090565b61012a91905b8082111561012657600081600090555060010161010e565b5090565b90565b61038b8061013c6000396000f30060606040526000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806341c0e1b514610053578063a413686214610068578063cfae3217146100c557600080fd5b341561005e57600080fd5b610066610153565b005b341561007357600080fd5b6100c3600480803590602001908201803590602001908080601f016020809104026020016040519081016040528093929190818152602001838380828437820191505050505050919050506101e4565b005b34156100d057600080fd5b6100d86101fe565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156101185780820151818401526020810190506100fd565b50505050905090810190601f1680156101455780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614156101e2576000809054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16ff5b565b80600190805190602001906101fa9291906102a6565b5050565b610206610326565b60018054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561029c5780601f106102715761010080835404028352916020019161029c565b820191906000526020600020905b81548152906001019060200180831161027f57829003601f168201915b5050505050905090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106102e757805160ff1916838001178555610315565b82800160010185558215610315579182015b828111156103145782518255916020019190600101906102f9565b5b509050610322919061033a565b5090565b602060405190810160405280600081525090565b61035c91905b80821115610358576000816000905550600101610340565b5090565b905600a165627a7a723058201ef87c90fca101864bb70ef1ac60d5510f2af93c1374d2c7c4f6ee8d40d9e4ba002900000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000"
	host     = "http://127.0.0.1:8547"
	privKeys = "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"
)
var contract string

func main() {

	flag.Usage = func() {

		fmt.Println("-ip   默认值 127.0.0.1")
		fmt.Println("-port 默认值 33333")
		fmt.Println("-sleep 每笔交易发送完成后的休眠时间,单位=Millisecond 默认值 0")
		fmt.Println("-model 0=转账 1=sol调用(自动部署) 2=sol部署 默认=0")
	}

	var ip string
	var port string
	var timeSleep int64
	var model int64
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip地址")
	flag.StringVar(&port, "port", "33333", "port地址")
	flag.Int64Var(&timeSleep, "sleep", 0, "每秒交易发出后的休眠时间")
	flag.Int64Var(&model, "model", 0, "0=转账 1=sol调用(自动部署) 2=sol部署 默认=0")

	flag.Parse()

	client, err := Connect("http://" + ip + ":" + port)
	fmt.Println("要发送的地址" + "http://" + ip + ":" + port)
	//client, err := eth1.Connect("http://10.0.0.116:33333")

	if err != nil {
		log.Fatal("ethclient Dial fail", err)
		return
	}
	//pri, err := crypto.HexToECDSA(privKeys)
	//from := crypto.PubkeyToAddress(pri.PublicKey)
	//nonce, _ := client.EthClient.PendingNonceAt(context.TODO(), from)
	//fmt.Println(nonce)

	switch model {
	case 0:
		//transfer
		txsend(client, sendTestTx, timeSleep)

	case 1:
		contract = DeploySol(client)
		fmt.Println("部署的合约地址是", contract)
		time.Sleep(10 * time.Second)
		txsend(client, sendSolidityTx, timeSleep)
	case 2:
		//sol部署
		txsend(client, DeployWasm, timeSleep)
	default:
		fmt.Println("交易类型选错了")
		return

	}

}

func txsend(client *Client, tx func(privKey string, client *Client, nonce uint64) error, timeSleep int64) {

	var nonce uint64
	pri, err := crypto.HexToECDSA(privKeys)
	from := crypto.PubkeyToAddress(pri.PublicKey)
	nonce, err = client.EthClient.PendingNonceAt(context.TODO(), from)
	if err != nil {
		log.Fatal("err2222", err)
		return
	}
	fmt.Println("第一次查出的nonce", "nonce", nonce)
	n := &nonce
	stop := make(chan uint64, 10)

	go nonceNum(n) //查询每秒发送多少交易

	go checkNonce(stop, client) //每分钟检查nonce

	for {
		select {
		case <-stop:
			fmt.Println("进入stop开始休眠")
			time.Sleep(20 * time.Second)
			fmt.Println("休眠完毕")

			nonce1, err := client.EthClient.PendingNonceAt(context.TODO(), from)
			if err != nil {
				log.Fatal("stop client.EthClient.PendingNonceAt err", err)
				return
			}
			fmt.Println("出错后查询的nonce是", nonce1)
			atomic.StoreUint64(&nonce, nonce1+1)
			if len(stop) != 0 {
				fmt.Println("这里面有值")
				<-stop
			}


		default:

			err := tx(privKeys, client, nonce)
			//出错后等待20秒后查询nonce
			if err != nil && !strings.HasPrefix(err.Error(),"known transaction"){
				fmt.Println("发送交易出错拉", err)
				stop <- 1
			} else {
				//交易没问题nonce++
				time.Sleep(time.Duration(timeSleep) * time.Millisecond)
				atomic.AddUint64(&nonce, 1)

			}

		}

	}
}

func nonceNum(n *uint64) {
	var n1 uint64
	for {
		time.Sleep(2 * time.Second)

		fmt.Println("每秒发送多少交易", atomic.LoadUint64(n)-n1, "当前的nonce", *n)

		n1 = *n
	}
}

func checkNonce(stop chan uint64, client *Client) {
	var pre uint64
	pri, _ := crypto.HexToECDSA(privKeys)
	from := crypto.PubkeyToAddress(pri.PublicKey)
	for {
		time.Sleep(37 * time.Second)
		nonce, _ := client.EthClient.NonceAt(context.TODO(), from,nil)
		fmt.Println("查询", "nonce", nonce, "pre", pre)

		if nonce == pre {
			fmt.Println("nonce相同停止交易", "nonce", nonce, "pre", pre)
			//如果nonce没有变化,停止交易
			stop <- 1

		} else {
			pre = nonce
		}

	}

}
