package ewasm

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go-eth/eth"
	"log"
)

func GetWasm(client *eth.Client, to common.Address) {
	key := "pdx"
	callMsg := ethereum.CallMsg{
		To:   &to,
		Data: []byte("get:" + key),
	}
	fmt.Println(hexutil.Bytes(callMsg.Data))
	bytes, err := client.EthClient.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Println(err)
	}
	fmt.Printf("key=%v,value=%v", key, bytes)
}
