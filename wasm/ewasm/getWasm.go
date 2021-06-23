package ewasm

import (
	"context"
	"fmt"
	"log"
	ethereum2 "pdx-chain"
	"pdx-chain/common"
	"pdx-chain/common/hexutil"
	"pdx-chain/utopia/utils/client"
)

func GetWasm(client *client.Client, to common.Address) {
	key := "pdx"
	callMsg := ethereum2.CallMsg{
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
