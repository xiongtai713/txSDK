package main

import (
	"context"
	"fmt"
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

//阿里云测试
//var add1 = []string{
//	"0xBbd299Ef1Cf31Be0dC4f1eD2fabCC1f4D885a578",
//	"0xECBAE96618e1B5e267665D4bd2C71E0856a427dD",
//	"0x0353aE126A02BA31F52652B120B845879DBD7Ff7",
//	"0x1c10F4AE153451aaE542A37204D3e2860f841C16",
//	"0x86082FA9d3c14D00a8627aF13CfA893e80B39101",
//	"0xdE11F74a62adb0AdDb5226F1ca806ebE61fB3372",
//	"0x24aFB10c3b11bE3472D0d353bAc70727f33Fc030",
//	"0x9d511c023bc5e6ddDf84A7719d5c576526405d27",
//	"0xFbF1dDf207776586c1cb5109AF131832D9a12FDf",
//	"0xD15646F79F40eA7764e06FC60b93a5CC857596f5",
//	"0xA688ba6B79bE0E74923219dba5F38Afd9c5cF70B",
//	"0xa15867ed27c07289F4605489AAcc8D4FC53CDDf1",
//	"0xf934dbece7A9467e01185ebeEE3132FE9EFaB407",
//	"0x18B9F96D416dEb262d2210E8a290A1048Af90703",
//	"0x3ea1E2452d20371E3c6fBC21944AD57bff4389fc",
//	"0x7a1B4bCfB687f2E3c24c6E9b07d07fcF3f3DFD77",
//	"0x06b0796938f0B80bc60Ef5fa47EadD5687DeCb5E",
//	"0x0AC357A27de8D0446DEbB2Ee2be8A2C142D698D6",
//	"0xa996e580dC8CA570F1A0107077Ca009337a3A60D",
//	"0x26d3B027085f09590d8cCE3d0518Beb83a50EeD7",
//	"0xAFDA1382F1fc0caB3b9B6Cc14F2b73407F80d6CE",
//	"0xD2c4fbf53e122F1a395d93160e52824696cE428f",
//	"0x5C0C470CBD233d9e15BCD09cbDaf221952fd4846",
//	"0xDb57866e3bb2C5e639bA55273d8d0fd351888Ccb",
//	"0x9df62A69c22dc2e480EE480F1389253a9Df124ea",
//	"0xb908c89486894088Ae2A1e566EBB4DfcB1B6cd9C",
//	"0xbC1BD303ebFcaEDe23CF2e71daBe4Fb5b63CeF13",
//	"0xcC206495041D856C9412Ec70e1cD86Ebcf9dee03",
//	"0x2C5887CeCd1cbA43C9B4Dc3041150c4727f6dec8",
//	"0x4dA44d91f4F361e29CC479D6F8C9307266C355E9",
//	"0x761aE9255a2877177C979130c7eA243A75777c22",
//}
//
var add1 = []string{
	"86082fa9d3c14d00a8627af13cfa893e80b39101",
	"2bc2cc3ce7ec05ce452185543f823697144f92b7",
	"c2f1b734992ecccb668680985ba1fa05c1b3b4d3",
	"eeef8fa829d1a6498feac6bef8f2a2fda9432709",
	"36acc454e6d7dcae3dc5f636eb473d04ebb63651",
	"a2e40246fba6cb5ff0557b08e852b7d5938b55ed",
	"78c69ad07da7af48e9b3123523b80027fcb4218c",
	"574bcdac1a7af9c92811c18ada2a94c4ca27a188",
	"04dfa2f3144a01f3a030321e98377b98036c079d",
	"f615866e9d3dbdd681d3ebd5888db9a562cc23ca",

	//"0x123f1F261fc7f9B99D36ba55B8463b961BdCc752",
	//"0x86082FA9d3c14D00a8627aF13CfA893e80B39101",
	//"0x9c3595413A95C8F8514fc6246F627a0A319F06e3",
	//"0xE1B0e8EC932089901655159ff056E0Bd03285c1A",


	//"0x7De2a31d6CA36302eA7b7917C4FC5eF4c12913b6",
	//"0xeaf7d7d101cf089fb87523dd4e5187c49dbc78dd",
	//"08b299d855734914cd7b19eea60dc84b825680f9",
}

func main() {

	//proxy := "http://10.0.0.180:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1ODUzMDM5MTcsIkZSRUUiOiJUUlVFIn0.wqT3bmVf2Zo5gbhq-wg3-pANSloYIarEUijHfbmJKkBStDWoft_bFZrviwZNux4C-WApGuCfqAHEarVxfuebUw"
	//client, err := eth.Connect("http://utopia-chain-739:8545", proxy, token)
	//client, err := eth.Connect("http://127.0.0.1:8547")
	//client, err := eth.Connect("http://39.100.210.156:30193")
	//client, err := eth.Connect("http://10.0.0.116:30100")
	client, err := eth.Connect("http://10.0.0.107:33333")

	//if err != nil {
	//	fmt.Printf("1", err.Error())
	//	return
	//}
	gasLimit := uint64(4712388)
	gasPrice := big.NewInt(240000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

	//privKey, err := crypto.HexToECDSA("a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381")
	//privKey, err := crypto.HexToECDSA("141ebc1d272e88789c2b1eedee3ecb243c95ff6ae0fc8c41658270636cf930c8")

	privKey, err := crypto.HexToECDSA("d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84")
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	from := crypto.PubkeyToAddress(privKey.PublicKey)
	fmt.Printf("from:%s\n", from.String())
	ctx, _ := context.WithTimeout(context.TODO(), 2*time.Second)
	nonce, err := client.EthClient.NonceAt(ctx, from, nil)
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

		nodeID := discover.PubkeyID(&privKey.PublicKey)
		fmt.Printf("nodeID:%s \n", nodeID.String())
		amount := big.NewInt(3000000000000000000)
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

		tx := types.NewTransaction(uint64(nonce), to, amount, gasLimit, gasPrice, payload1)
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
