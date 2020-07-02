package vm

import (
	"pdx-chain/params"
	"testing"
)

func TestInTokenChainConfig(t *testing.T) {
	//tokenChain := []params.TokenChain{
	//	{ChainId:"123", RpcHosts:[]string{"a", "b", "c"}},
	//	{ChainId:"456", RpcHosts:[]string{}},
	//	{ChainId:"789", RpcHosts:[]string{"a", "b"}},
	//	{ChainId:"abn", RpcHosts:[]string{"a"}},
	//	{ChainId:"ddd", RpcHosts:[]string{"a", "b", "d"}},
	//}

	tokenChain2 := []params.TokenChain{}
	chainID := "dsfgr"
	in := inTokenChainConfig(tokenChain2, chainID)
	t.Log("----------", in)
}
