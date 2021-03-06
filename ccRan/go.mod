module go-eth/callsol

go 1.15

replace pdx-chain => /Users/liu/code/pdx-chain

replace github.com/cc14514/go-alibp2p => /Users/liu/code/pdx-chain/build/_deps/go-alibp2p

replace github.com/cc14514/go-alibp2p-ca => /Users/liu/code/pdx-chain/build/_deps/go-alibp2p-ca

replace github.com/cc14514/go-certool => /Users/liu/code/pdx-chain/build/_deps/go-certool

replace github.com/tjfoc/gmsm => github.com/PDXbaap/gmsm v1.3.11

require (
	github.com/ethereum/go-ethereum v1.9.20 // indirect
	github.com/golang/protobuf v1.3.3
	pdx-chain v0.0.0-00010101000000-000000000000
)
