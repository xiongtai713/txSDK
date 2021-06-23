package main

import (
	"time"
)

const (
	//host         = "http://39.100.34.235:30074"
	//host         = "http://39.100.93.177:30036"
	//host = "http://10.0.0.203:33333"
	//host11 = "http://39.100.210.205:30178"
	//host1 = "http://10.0.0.245:30100"
	host1         = "http://127.0.0.1:8547"
	sendDuration  = time.Minute * 100
	nonceTicker   = time.Minute * 100 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
	sleepDuration = time.Minute * 100 //查完nonce后休眠时间（1处）
	txNum         = 1

	ACTION              = "PDX_SAFE_ACTION"
	JWT                 = "PDX_SAFE_AUTHZ"
	PUBK_HEX_LEN        = 66

)


var (
	PDXS_CREATE_DOMAIN = []byte{0}
	PDXS_UPDATE_DOMAIN = []byte{1}
	PDXS_DELETE_DOMAIN = []byte{2}
	PDXS_GET_DOMAIN    = []byte{3}
	PDXS_CREATE_KEY    = []byte{4}
	PDXS_UPDATE_KEY    = []byte{5}
	PDXS_DELETE_KEY    = []byte{6}
	PDXS_GET_KEY       = []byte{7}
)

var privKeys = []string{
	//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381", //7DE
	//"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
	"8160598508637f8fef0ef7021afac26c8e8edb5fb21a4bfa9c679b54b74ea701",

}

func main() {
	var (
		domain = "刘蛟"
		//key    = "hh"
		//value  = "ooo"
	)
	//creatDomain(domain)
	getDomain(domain)
	//CreatKey(key, value, NewJwt(domain, key))
	//getKey(domain, key)
	//parseJWT(NewJwt())
	//fmt.Println(domain, key, value)
}
