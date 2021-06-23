package main

import (
	"go-eth/callsol/ewasm"
	"pdx-chain/common"
	"pdx-chain/utopia/utils/client"

	"log"
)

var (
	path  = "/Users/liu/rust/hello-wasm/pkg/hello_wasm_bg.wasm"
	path2 = "/Users/liu/rust/ewasm-rust-demo/hello-wasm-abi/pkg/hello_wasm_abi_bg.wasm"
	path3 = "/Users/liu/Desktop/11/hello.sol"
	//host     = "http://39.100.84.63:8545"
	//host     = "http://127.0.0.1:8547"
	privKeys = "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84"
)

func main() {
	//client, err := eth1.Connect(host)

	//proxy := "http://test.blockfree.pdx.ltd"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1NzQ4NDQwODcsIkZSRUUiOiJUUlVFIn0.lVPyP9xJ2iurl1_-UvdGXFWBGP65qS-NSqN5giUOpkZBvaQ8LO-X5MIP3-A1aUoV-0E3x9ucnMU6YCUUH89PCQ"
	//client, err := eth1.Connect("http://utopia-chain-1000004:8545", proxy, token)
	client, err := client.Connect("http://39.100.210.205:30100")
	//client, err := client.Connect("http://127.0.0.1:8546")
	//client, err := client.Connect("http://39.100.210.205:30178")

	if err != nil {
		log.Fatal("ethclient Dial fail", err)
		return
	}
	 deploy(client)
	//to := common.HexToAddress("0xe20a7ec406bd368404f61f6d72eaa0977ee44c41")
	////
	////time.Sleep(10 * time.Second)
	//for i := 0; i < 100; i++ {
	//	put(client, to)
	//}

	//put(client, to)
	//get(client, to)
}

func deploy(client *client.Client) common.Address {
	return ewasm.DeployWasm(privKeys, client, path)
}

func put(client *client.Client, to common.Address) {

	ewasm.PutWasm(privKeys, client, to)
}

func get(client *client.Client, to common.Address) {
	ewasm.GetWasm(client, to)
}
