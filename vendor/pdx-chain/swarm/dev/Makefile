.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    pdx-chain/cmd/swarm
	go build -o bin/geth     pdx-chain/cmd/geth
	go build -o bin/bootnode pdx-chain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v pdx-chain/swarm/...
