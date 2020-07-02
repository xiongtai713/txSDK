package vm

import (
	"context"
	"fmt"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/log"
	"pdx-chain/p2p/discover"
	"pdx-chain/rlp"
	"pdx-chain/utopia/utils/client"
	"testing"
)

func TestHypothecation_Run3(t *testing.T) {
	if client, err := client.Connect("http://192.168.0.121:8545"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {

		if privKey, err := crypto.HexToECDSA("1263354aca47fe21d54e0fd4d125e879142a1367e6962bcdc8dd251830f4b092"); err != nil {
			fmt.Printf(err.Error())
			return
		} else {
			nodeID := discover.PubkeyID(&privKey.PublicKey)
			fmt.Printf("nodeID:%s \n", nodeID.String())
			from := crypto.PubkeyToAddress(privKey.PublicKey)
			fmt.Printf("from:%s\n", from.String())
			//退钱
			to := common.BytesToAddress(crypto.Keccak256([]byte("quiteQuorum"))[12:])

			fmt.Printf("to:%s\n", to.String())
			fmt.Println(from.String())
			if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
				fmt.Printf("ddddd:%s", err.Error())
				return
			} else {
				amount := big.NewInt(3000000000000000000)
				gasLimit := uint64(4712388)
				gasPrice := big.NewInt(240000000000) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

				////退钱需要填充的
				type QuiteQuorumInfo struct {
					QuiteAddress common.Address `json:"hypothecation_addr"` //质押的地址
				}

				/*
					对于退款地址可以通过以下方法设置
					addr := common.HexToAddress("0x41b84e81ba73ab6910ac9b2038d6f18c15a05367")
				*/

				recaption := &QuiteQuorumInfo{
					from,
				}

				payload, err := rlp.EncodeToBytes(recaption)

				log.Info("payload", "payload", string(payload))

				if err != nil {
					fmt.Printf("marshal XChainTransferWithdraw err:%s", err)
					return
				}

				tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, payload)
				//EIP155 signer
				signer := types.NewEIP155Signer(big.NewInt(123))
				signedTx, _ := types.SignTx(tx, signer, privKey)
				if txHash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
					fmt.Println("yerror", err.Error())
				} else {
					fmt.Printf("Transaction hash: %s\n", txHash.String())
				}
			}
		}
	}
}
