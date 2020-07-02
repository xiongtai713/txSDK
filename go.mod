module go-eth

go 1.13

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/deckarep/golang-set v0.0.0-20180603214616-504e848d77ea
	github.com/ethereum/go-ethereum v1.9.13
	github.com/golang/protobuf v1.3.3
	github.com/rs/cors v0.0.0-20160617231935-a62a804a8a00
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tjfoc/gmsm v1.3.0
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce
)

replace github.com/ethereum/go-ethereum v1.9.13 => /Users/liu/go/src/go-eth/vendor/github.com/ethereum/go-ethereum
