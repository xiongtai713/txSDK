package sol

import (
	"context"
	"fmt"
	"go-eth/rpc"
	"pdx-chain/accounts/abi/bind"
	"pdx-chain/common"
	"pdx-chain/ethclient"
)

func main() {
	ctx := context.Background()


	endpoint := "http://39.100.92.41:8545"
	addr := common.HexToAddress(contract1)


	client, _ := rpc.Dial(endpoint)
	ethClient := ethclient.NewClient(client)  //ethClient实现了backend
	caller,_ := NewApiname(addr,ethClient)

	opts := &bind.CallOpts{
		Context: ctx,
	}
	s, _ := caller.Greet(opts)
	fmt.Println(s)
}
