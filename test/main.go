package main

import (

	"github.com/multiformats/go-multiaddr"
)

//16*8+20

func main() {
	newMultiaddr, _ := NewMultiaddr("/ip4/0.0.0.0/tcp/9000")

}

//func PubkeyID(pub *ecdsa.PublicKey) discover.NodeID {
//	var id discover.NodeID
//	pbytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
//	if len(pbytes)-1 != len(id) {
//		panic(fmt.Errorf("need %d bit pubkey, got %d bits", (len(id)+1)*8, len(pbytes)))
//	}
//	copy(id[:], pbytes[1:])
//	return id
//}
//
//func rlpHash(x interface{}) (h common.Hash) {
//	hw := sha3.NewKeccak256()
//	rlp.Encode(hw, x)
//	hw.Sum(h[:0])
//	return h
//}
