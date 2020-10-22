module transfer

go 1.15

replace github.com/ethereum/go-ethereum => /Users/liu/go/src/go-eth/vendor/github.com/ethereum/go-ethereum

replace go-eth/eth => /Users/liu/go/src/go-eth

require (
	github.com/ethereum/go-ethereum v1.9.13
	github.com/syndtr/goleveldb v1.0.0 // indirect
	go-eth/eth v0.0.0-00010101000000-000000000000
)
