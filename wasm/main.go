package main

import (
	"github.com/ethereum/go-ethereum/common"
	"go-eth/eth"
	"go-eth/wasm/ewasm"
	"log"
)

var (
	path  = "/Users/liu/rust/hello-wasm/pkg/hello_wasm_bg.wasm"
	path2 = "/Users/liu/rust/ewasm-rust-demo/hello-wasm-abi/pkg/hello_wasm_abi_bg.wasm"
	//host     = "http://39.100.84.63:8545"
	host     = "http://127.0.0.1:8549"
	privKeys = "a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381"
)

func main() {
	//client, err := eth.Connect(host)

	proxy := "http://test.blockfree.pdx.ltd"
	token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1NzQ4NDQwODcsIkZSRUUiOiJUUlVFIn0.lVPyP9xJ2iurl1_-UvdGXFWBGP65qS-NSqN5giUOpkZBvaQ8LO-X5MIP3-A1aUoV-0E3x9ucnMU6YCUUH89PCQ"
	client, err := eth.Connect("http://utopia-chain-1000004:8545", proxy, token)

	if err != nil {
		log.Fatal("ethclient Dial fail", err)
		return
	}

	//deploy(client)

	to := common.HexToAddress("0x7d5dd48b41876e07201086602fda19d0b56ee1e1")
	//put(client, to)
	get(client, to)
}

func deploy(client *eth.Client) {
	ewasm.DeployWasm(privKeys, client, path)
}

func put(client *eth.Client, to common.Address) {

	ewasm.PutWasm(privKeys, client, to)
}

func get(client *eth.Client, to common.Address) {
	ewasm.GetWasm(client, to)
}
