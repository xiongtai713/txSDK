package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/crypto/gmsm/sm2"
	"pdx-chain/utopia/utils/client"
	"time"
)

//private:5ca4829b9ad9ba68e74a747115e33ef3998f0f786f924e6f3b6ec2e56504ed15
//public:03da13b473060d0e7c5217923cc395eac4c500bdc7b33731680ec97c503c5b181d
//address:251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8

func main() {
	if client, err := client.Connect("http://192.168.3.3:8546"); err != nil {
		//if client, err := eth.Connect("http://10.0.0.124:8548"); err != nil {
		//if client, err := eth.Connect("http://47.92.137.120:30100"); err != nil {
		//if client, err := eth.Connect("http://10.0.0.17:8546"); err != nil {
		//if client, err := eth.Connect("http://10.0.0.8:8547"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		//addr: 0xa2b67b7e4489549b855e423068cbded55f7c5daa
		//if privKey, err := crypto.HexToECDSA("0425d98f3aa229626c144bb278d1e17943e4e154f82144da824d18cdc3951287"); err != nil {

		sbin, ok := new(big.Int).SetString("a2f1a32e5234f64a6624210b871c22909034f24a52166369c2619681390433aa", 16)
		if !ok {
			return
		}
		privKey := sm2.InitKey(sbin)

		//if privKey, err := crypto.HexToECDSA(""); err != nil {
		//	fmt.Printf(err.Error())
		//	return
		//} else {
		from := crypto.PubkeyToAddress(privKey.PublicKey)
		fmt.Printf("from:%s\n", from.String())
		//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e
		to := common.HexToAddress("251b3740a02a1c5cf5ffcdf60d42ed2a8398ddc8")
		//to := from //转给自己
		fmt.Printf("to:%s\n", to.String())

		ctx, _ := context.WithTimeout(context.TODO(), 30*time.Second)

		if nonce, err := client.EthClient.NonceAt(ctx, from, nil); err != nil {
			fmt.Printf("nonce err: %s", err.Error())
			return
		} else {
			//amount := big.NewInt(0).Mul(big.NewInt(1000), big.NewInt(1e18))
			amount := big.NewInt(0).Mul(big.NewInt(0), big.NewInt(1000000000000000000))
			//amount := big.NewInt(0).Mul(big.NewInt(42000), big.NewInt(10000000000))
			gasLimit := uint64(21000)
			gasPrice := big.NewInt(399000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了
			//18900000000000
			//2999999999000000000000000000
			fmt.Println("nonce is what", ":", nonce)
			//nonce = 7000

			nc, err := client.EthClient.NonceAt(ctx, to, nil)
			if err != nil {
				fmt.Println("to nonce err", err.Error())
				return
			}
			fmt.Println("to nonce is------->", nc)

			type Meta map[string][]byte
			meta := Meta{"jwt": []byte("eyJhbGciOiJTTTIiLCJ0eXAiOiJKV1QifQ.eyJhayI6IjAyYjk5NGY4NzMxODBhNWQ0ZmRmMWY3YWYwMjM4MDIyYjIyYzAzYjA5ZWU4NDA0NGI3ODAxZDRiNzg0MTc1MWNkOCIsImwiOjYwMDAwMDAwMDAwMDAwMDAsIm4iOiJlZWZmZmVmcmVyZWRmZmRzdXVmMnJyZmRzbWZsamxqcnJhIiwiciI6InUiLCJzIjoxNzM0OCwic2siOiIwMjU5NWQ1NTM2OTczMDVjNzY3MGRmZDkyNjI4ZTVmZjY4MDgwMzM1MjY1ZWRmODA0YWVhNGU2ZThkZjUxMTI0NjQifQ.bwi21lPcJsN9uzoV7gwkyXWy0Mfd9u6oKvc1leY57kBAY-xLqUzmccqwEMQGwHetU-CypcUlpb71QKCSigw4Rw")}
			metaByte, _ := json.Marshal(meta)
			metaStr := hex.EncodeToString(metaByte)
			flagNo0x := "26cad4db1a82fab9e41bf5c0fbb4937a27b93786a4f27d3b9704805f698d3e65"
			dataStr := flagNo0x + "_" + "" + "_" + metaStr
			data := []byte(dataStr)

			//data := make([]byte, 1000*300)
			tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
			//signer := types.HomesteadSigner{}
			//signer := types.NewEIP155Signer(big.NewInt(738))
			signer := types.NewSm2Signer(big.NewInt(738))
			signedTx, _ := types.SignTx(tx, signer, (*ecdsa.PrivateKey)(privKey))
			if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
				fmt.Println("send raw transaction err:", err.Error())
			} else {
				fmt.Printf("Transaction hash: %s\n", txHash.String())
			}
		}
	}
}
