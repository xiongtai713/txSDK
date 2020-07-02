package engine

import (
	"fmt"
	"math/big"
	"testing"
)

func TestHalfReward(t *testing.T) {
	bns := []*big.Int{
		big.NewInt(1),
		big.NewInt(1500000),
		big.NewInt(3000000),
		big.NewInt(3000001),
		big.NewInt(4000003),
		big.NewInt(9000000),
		big.NewInt(9000001),
		big.NewInt(10000003),
		big.NewInt(27000000),
		big.NewInt(27000001),
		big.NewInt(37000001),
		big.NewInt(45000000),
		big.NewInt(45000001),
		big.NewInt(55000001),
		big.NewInt(63000000),
		big.NewInt(103000000),
	}

	for _, bn := range bns {
		halfN := HalfReward(bn)
		r := big.NewInt(0).Div(FrontierBlockReward, big.NewInt(halfN))
		fmt.Println(bn.String(), "--->", r.String(), "======", FrontierBlockReward,"--->", halfN)
	}
}
