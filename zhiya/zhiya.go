package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/rlp"
	"go-eth/eth"
	"math/big"
	"time"
)

type HypothecationAddress struct {
	Address common.Address `json:"address"` //抵押地址
}

var add1 = []string{
	//主网地址
	//"0xaC3D28DF6479721A7E89E88D51eA4A6A12876B60",
	//"0x645121680e00c90472E64007fcb98eB77f2f3748",
	//"0xf915aC430A797240C70eA0f87dd0758DFa1F345d",
	//"0xC5308bCD8fd4074B2F7EA2BE7d3D08eA03eF3333",
	//"0xFe0ebCb946Acc0417c9D10373f09C070aB20f3a7",
	//"0x3A556e875389956068Dd7662b35331F1AD5fC10f",
	//"0xF6031E7c6CB039a2ED4f4DF8c644CA2782a332B5",
	//"0x78959fafE08a25403241f2A378795BAc8C91c7f9",
	//"0xF37717D7C824e803d0c02f9A49548809CD666001",
	//"0xd0A39991723482B48EdFd7FE5079adF1Cd7DC7E5",
	//"0xC67cF55a9CA4A9850074B64da151Cc3F8d34A203",  //通过钱包发送
	//"0x28394786D5Ef821F129726770f42Ab9Ce0A30307",
	//"0xBD56Cb2Bb73177A1DE8eFaD09ed363b6DA619fF5",
	//"0x36910f65888b05294fAB86eA57455d199CD6D9e0",
	//"0x0AaC10F043F83510d7A171DC6cE43Dc2C1937C9f",
	//"0xfe4F28DC9FCf3364190eE7458F5C24CCd5aDfb43",
	//上面已经质押
	//"0x63B31Efa4E50D3c051bb0c68E05Bc670fF4Ef5B7",
	//"0x42903E431867594C7f584f7bB633115B0050A56a",
	//"0xd80d31b8c55aA8E0DBb76B0fFE96a8Fe36855dF1",
	//"0x32e507d61F9c894Ab36CEba030149cfC278b9E25",
	//"0x970eA8B298DF190ceC4D5C49317871b13CCFD01F",
	//"0xFa438a905CD413fCed5079Fd0c7585AaC687Dd4B",
	//"0xC29e28e3F3F58D94F2ba67E045Ab0a433F15151f",

	"0xFa438a905CD413fCed5079Fd0c7585AaC687Dd4B",
	"0xC29e28e3F3F58D94F2ba67E045Ab0a433F15151f",
}


func main() {

	//proxy := "http://10.0.0.180:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1ODUzMDM5MTcsIkZSRUUiOiJUUlVFIn0.wqT3bmVf2Zo5gbhq-wg3-pANSloYIarEUijHfbmJKkBStDWoft_bFZrviwZNux4C-WApGuCfqAHEarVxfuebUw"
	//client, err := eth.Connect("http://utopia-chain-739:8545", proxy, token)
	//client, err := eth.Connect("http://47.92.156.106:30233")
	//client, err := eth.Connect("http://39.100.210.156:30193")
	//client, err := eth.Connect("http://127.0.0.1:8547")
	client, err := eth.Connect("http://47.94.209.251:30111")
	//client, err := eth.Connect("http://10.0.0.34:33333")

	//if err != nil {
	//	fmt.Printf("1", err.Error())
	//	return
	//}

	gasLimit := uint64(4712388)
	gasPrice := new(big.Int).Mul(big.NewInt(1e9), big.NewInt(4000)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

	privKey, err := crypto.HexToECDSA("a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381") //0x7de2a31d6ca36302ea7b7917c4fc5ef4c12913b6
	//privKey, err := crypto.HexToECDSA("141ebc1d272e88789c2b1eedee3ecb243c95ff6ae0fc8c41658270636cf930c8")

	//privKey, err := crypto.HexToECDSA("d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	from := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("from:%s\n", from.String())
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	nonce, err := client.EthClient.PendingNonceAt(ctx, from)
	if err != nil {
		fmt.Println("nonce err", err)
		return
	}
	for i := 0; i < len(add1); i++ {

		//addr := common.HexToAddress("0x7De2a31d6CA36302eA7b7917C4FC5eF4c12913b6")
		addr := common.HexToAddress(add1[i])
		recaption := &HypothecationAddress{
			addr,
		}
		payload1, _ := rlp.EncodeToBytes(recaption)
		fmt.Println("nonce", nonce)

		nodeID := discover.PubkeyID(&privKey.PublicKey)
		fmt.Printf("nodeID:%s \n", nodeID.String())
		amount := new(big.Int).Mul(big.NewInt(100000), big.NewInt(1e18))
		//质押
		to := common.BytesToAddress(crypto.Keccak256([]byte("hypothecation"))[12:])

		fmt.Printf("to:%s\n", to.String())
		fmt.Println(from.String())

		//退钱

		//amount := big.NewInt(0)
		//to := common.BytesToAddress(crypto.Keccak256([]byte("redamption"))[12:])
		//fmt.Printf("to:%s\n", to.String())
		//
		//////退钱需要填充的
		//type RecaptionInfo struct {
		//	HypothecationAddr common.Address `json:"hypothecation_addr"` //质押的地址
		//
		//	RecaptionAddress common.Address `json:"recaption_address"` //退款地址
		//}
		//
		//recaption := &RecaptionInfo{
		//	HypothecationAddr: common.HexToAddress("0x7De2a31d6CA36302eA7b7917C4FC5eF4c12913b6"),
		//
		//	RecaptionAddress: common.HexToAddress("0x7De2a31d6CA36302eA7b7917C4FC5eF4c12913b5"),
		//}
		//payload1, err := rlp.EncodeToBytes(recaption)
		//
		//if err != nil {
		//	fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
		//	return
		//}

		msg := ethereum.CallMsg{
			From:  from,
			Data:  payload1, //code =wasm code1 =sol
			To:    &to,
			Value: amount,
		}
		gas, err := client.EthClient.EstimateGas(context.TODO(), msg)
		if err != nil {
			fmt.Println("错误拉", err)
		}
		price, err := client.EthClient.SuggestGasPrice(context.TODO())
		if err != nil {
			fmt.Println("错误拉", err)
		}
		fmt.Println("预估gas是", gas, "price", price)

		gasLimit = gas
		tx := types.NewTransaction(uint64(nonce), to, amount, gasLimit, gasPrice, payload1)
		fmt.Println("input",common.ToHex(payload1))
		//EIP155 signer
		signer := types.NewEIP155Signer(big.NewInt(739))
		//signer := types.HomesteadSigner{}
		signedTx, _ := types.SignTx(tx, signer, privKey)
		// client.EthClient.SendTransaction(context.TODO(), signedTx)
		if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
			fmt.Println("yerror", err.Error())
			return
		} else {
			fmt.Println("Transaction hash:", txHash.String(), "nonce", nonce)
			nonce++

		}
	}

}
