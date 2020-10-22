package main

import (
	"bytes"
	"compress/gzip"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"log"
	"pdx-chain/common"
	"pdx-chain/crypto/sha3"
	"pdx-chain/p2p/discover"
	"pdx-chain/rlp"
	"strconv"
)

//16*8+20

func main() {

	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	defer writer.Close()
	for i := 0; i < 100; i++ {
		itoa := strconv.Itoa(i)
		writer.Write([]byte(itoa))
	}
	writer.Flush()
	fmt.Println("压缩后查看大小", buffer.Len())

	reader, err := gzip.NewReader(&buffer)
	if err != nil {
		log.Fatal("33",err)
	}
	all:=make([]byte,buffer.Len())
	defer reader.Close()
	_, err = reader.Read(all)
	//all, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(all))

}

func PubkeyID(pub *ecdsa.PublicKey) discover.NodeID {
	var id discover.NodeID
	pbytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	if len(pbytes)-1 != len(id) {
		panic(fmt.Errorf("need %d bit pubkey, got %d bits", (len(id)+1)*8, len(pbytes)))
	}
	copy(id[:], pbytes[1:])
	return id
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}
