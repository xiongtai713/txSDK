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
	"0xa74d8370dca6c1eef7695ffcba1cda49dff643e1",
	"0x3ff719eb4afd1dca24f14332881bb409a4ae892e",
	"0x17c4232c51efb424c545d4bf2a237e4ad23a95dd",
	"0x421404e3772361ebcc7fa0cca505d3acd5bf02a6",
	"0xa10d384a1cab19378aad1702f249006726620cb4",
	"0x5b3a6bb45b8913be639731020ab9bdef793d4fac",
	"0xbd792bedfceb34668c2ed43286fa83f735965047",
	"0x9f95da70de68a38846f85c3bd000cb5741ed1331",
	"0x4b6c14f9eb1928e2ca07950ebe33e8efa328d535",
	"0x63c4765de997924536195ea8330419e623a76e08",
	"0xa0d194b3c02723d0a64dd010a65a4d1bec33674f",
	"0x4b5419218cdb347e01ccab014cd489d57fbadf0c",
	"0xdee52e2500e64b452ddb99dc1a630ebc5c92e9a6",
	"0x1f3ecb95883ef87b44ca2e2f781f529f1f21135d",
	"0x38dadd6719d458a6f6983667aba1a5b769357bae",
	"0x0eeb438886a6cb19698decaf478bd30973ed2b5d",
	"0x2b30c2db92d862d1bab13b181ae0364c319b2cab",
	"0x16c51a2f2024bb9594d06205d5d40902d9aeb9a3",
	"0x3787a11c3af6ee4da02c182b7f4f54ecf7be5faf",
	"0xf9f7741e6605066d1eb3c610ffe7e3621e0ae888",
	"0x59fb171820984c828fb7198f6cb2755bd7e46ace",
	"0x6c37a7e64eee34f2404e0ad73148a832dd06631a",
	"0xb3addf37a66275cd0ff4d910cdf6727c7f22c897",
	"0xa40989859e684e847f15760c7c227e974e2dffe7",
	"0x2f4c7dcb930b6c23a57c71a518c9238dd0d20b84",
	"0xbd697dd3716a9e3949b21d5bd6933cc99ab87832",
	"0x8f1f4bfa1920dfb7fc3204bf0fd61b59963c8c97",


//"0x7De2a31d6CA36302eA7b7917C4FC5eF4c12913b6",
	//"08b299d855734914cd7b19eea60dc84b825680f9",
}

func main() {

	//proxy := "http://test.blockfree.pdx.ltd"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1NzQ4NDQwODcsIkZSRUUiOiJUUlVFIn0.lVPyP9xJ2iurl1_-UvdGXFWBGP65qS-NSqN5giUOpkZBvaQ8LO-X5MIP3-A1aUoV-0E3x9ucnMU6YCUUH89PCQ"
	//client, err := eth.Connect("http://utopia-chain-739:8545", proxy, token)
	//client, err := eth.Connect("http://127.0.0.1:8547")
	//client, err := eth.Connect("http://47.92.137.120:30402")
	client, err := eth.Connect("http://10.0.0.158:30049")

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
