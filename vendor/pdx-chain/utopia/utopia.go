/*************************************************************************
 * Copyright (C) 2016-2019 PDX Technologies, Inc. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *************************************************************************/
package utopia

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/dgrijalva/jwt-go"
	"github.com/tjfoc/gmsm/sm2"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	mathRand "math/rand"
	"net/http"
	"os/exec"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/crypto/sha3"
	"pdx-chain/log"
	"pdx-chain/pdxcc/util"
	"pdx-chain/rlp"
	"pdx-chain/utopia/utils"
	"pdx-chain/utopia/utils/client"
	"pdx-chain/utopia/utils/tcUpdate"

	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Org struct {
	Name   string   `json:"name"`
	NodeCa []string `json:"node_ca"` //node ca file
	UserCa []string `json:"user_ca"`
}
type consortiumConf struct {
	Name string `json:"name"`
	Orgs []Org  `json:"orgs"`
	//Cert string `json:"cert"`//current cert chain
	UserAuth bool `json:"user_auth"`
	DappAuth bool `json:"dapp_auth"`
}

type BlockFreeToken struct {
	DestChainId string
	Token       string
	Expired     int64
}

// map[destChainId]BlockFreeToken
// destChainId 是唯一的
type SafeBlockFreeCloud struct {
	BFCMap map[string]BlockFreeToken
	sync.RWMutex
}

//map[destChainId]BlockFreeToken
var blockFreeCloud = &SafeBlockFreeCloud{BFCMap: make(map[string]BlockFreeToken)}
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var (
	Config               *cli.Context
	Consortium           bool
	ConsortiumDir        string
	DataDir              string
	ConsortiumConf       consortiumConf
	Cert                 string //local node cert
	CertRootId           string
	RootIdCaMap          map[string]string
	UserCertPublicKeyMap map[string]struct{}
	NetMux               bool //是否走mux
	Perf                 bool
	StartTime            time.Time
	MinerPrivateKey *ecdsa.PrivateKey
	Syncing         *int32
)

const (
	BlackListExpire         = 1  //10
	XChainTransferCommitNum = 1  //10
	TrustTxCommitBlockLimit = 10 //30
	TrustTxGasLimit         = 10000

	LocalCert = "localhost.crt"

	ProxyRPC = "http://utopia-chain-%s:8545"

	TrustTxTokenType  = 0
	XChainTokenType   = 1
	RabbitMqTokenType = 2

	rabbitMQTokenKey = "rabbitMQ"
)

var ExtensionFlag = util.EthHash([]byte("utopia")).Hex()[2:]
var ExtensionFlagByte = []byte(util.EthHash([]byte("utopia")).Hex()[2:])

//data hash
func Hash(v interface{}) [32]byte {
	//bytes, err := json.Marshal(v)
	bytes, err := rlp.EncodeToBytes(v)
	if err != nil {
		log.Info("rlp encode error :")
	}

	return sha256.Sum256(bytes)
}

//create privateKey and publickKey
func NewKey() (*ecdsa.PrivateKey, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return privateKeyECDSA, nil
}

func SigToAddress(hash, sig []byte) (common.Address,*ecdsa.PublicKey, error) {
	if len(sig) > 65 {
		R := big.NewInt(0).SetBytes(sig[0:32])
		S := big.NewInt(0).SetBytes(sig[32:64])
		pub := sm2.Decompress(sig[64:])
		if sm2.Verify(pub, hash, R, S) {
			return crypto.PubkeyToAddress((ecdsa.PublicKey)(*pub)), (*ecdsa.PublicKey)(pub), nil
		}
		return common.Address{}, nil, errors.New("invalid signature")
	}
	key, err := crypto.SigToPub(hash, sig)
	if err != nil {
		log.Error("err", "err", err)

		return common.Address{},nil, errors.New("publicKey error")
	}
	pubBytes := crypto.FromECDSAPub(key)
	return common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:]),key, nil
}

//通过签名生成公钥
func Ecrecover(hash []byte, sig []byte) ([]byte, error) {
	if len(sig) > 65 {
		R := big.NewInt(0).SetBytes(sig[0:32])
		S := big.NewInt(0).SetBytes(sig[32:64])
		pub := sm2.Decompress(sig[64:])
		if sm2.Verify(pub, hash, R, S) {
			return sig[64:], nil
		}
		return nil, errors.New("invalid signature")
	}
	pub, err := crypto.Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}
	return pub, err
}

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

func VerifySignature(pubkey, hash, sign []byte) bool {

	if len(sign) != 65 {
		return false
	}
	sig := &btcec.Signature{R: new(big.Int).SetBytes(sign[:32]), S: new(big.Int).SetBytes(sign[32:])}
	key, err := btcec.ParsePubKey(pubkey, btcec.S256())
	if err != nil {
		return false
	}
	// Reject malleable signatures. libsecp256k1 does this check but btcec doesn't.
	//if sig.S.Cmp(secp256k1halfN) > 0 {
	//	return false
	//}
	return sig.Verify(hash, key)
}

func CompressPubkey(pubkey *ecdsa.PublicKey) []byte {
	return (*btcec.PublicKey)(pubkey).SerializeCompressed()
}

func Keccak256ToAddress(ccName string) common.Address {
	hash := sha3.NewKeccak256()

	var buf []byte
	hash.Write([]byte(ccName))
	buf = hash.Sum(buf)

	return common.BytesToAddress(buf)
}

func Shuffle(hosts []string) {
	for l := len(hosts); l > 1; l-- {
		n := mathRand.Intn(l)
		hosts[l-1], hosts[n] = hosts[n], hosts[l-1]
	}
}

func GetMaxNumFromTrustChain(hosts []string, chainID, destChainId, token string) *big.Int {
	retryN := 0
	Shuffle(hosts)
retry:
	for i, dstHost := range hosts {
		log.Info("try get max num", "i", i, "dstHost", dstHost)
		if i > 4 {
			log.Error("try five times fail", "i", i)
			break
		}

		proxy := Config.String("blockFreeCloud")
		cli, err := client.Connect(dstHost, proxy, token)
		if err != nil {
			log.Error("!!!!connect error", "err", err)
			continue
		}

		maxNum, err := cli.GetMaxNumFromTrustChain(chainID)
		if err != nil {
			log.Error("get maxNum from trust error", "err", err)
			if retryN == 0 {
				log.Info("get new hosts")
				trustHosts := GetTrustNodeFromIaas(chainID, destChainId, token)
				if len(trustHosts) == 0 {
					log.Error("trust nodes empty from Iaas")
					return nil
				}
				hosts = trustHosts
				retryN = 1
				goto retry
			}
			continue
		}
		return maxNum
	}
	return nil
}

//get trust chain hosts list
func GetTrustNodeFromIaas(chainID string, destChainID, token string) (hosts []string) {
	iaasServer := Config.String("iaas")
	proxyServer := Config.String("blockFreeCloud")
	log.Info("get trust node from iaas", "iaas", iaasServer, "proxy", proxyServer)
	if iaasServer == "" {
		return nil
	}

	if proxyServer != "" && token != "" {
		log.Info("goto proxy, not from iaas")
		return []string{fmt.Sprintf(ProxyRPC, destChainID)}
	}

	//http://localhost:8080/rest/chain/tchosts?chainId=739
	client := &http.Client{Timeout: time.Millisecond * 800}
	resp, err := client.Get(iaasServer + "/rest/chain/tcinfo?chainId=" + chainID)
	if err != nil {
		log.Warn("resp error", "err", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read resp body error", "err", err)
		return nil
	}
	log.Info("get trust node resp", "status", resp.Status)

	type Data struct {
		ChainId   string   `json:"chainId"` //down layer trust chain ID
		Hosts     []string `json:"hosts"`
		UnityCert string   `json:"unityCert"`
		Salt      string   `json:"salt"`
		Timestamp string   `json:"timestamp"`
		Signature string   `json:"signature"`
	}

	type RspBody struct {
		Status int               `json:""`
		Meta   map[string]string `json:"meta"`
		Data   Data              `json:"data"`
	}
	var result RspBody
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("body unmarshal error", "err", err)
		return nil
	}

	//验证签名
	toPub := types.GenPubKeyFromIAASCert(result.Data.UnityCert)
	if toPub == nil {
		log.Error("gen pub key from cert fail")
		return nil
	}
	certPubkey := crypto.FromECDSAPub(toPub)

	//digest hash
	hostStr := ""
	for _, v := range result.Data.Hosts {
		hostStr += v
	}
	digest := result.Data.ChainId + hostStr + result.Data.UnityCert + result.Data.Salt + result.Data.Timestamp
	digestHash := crypto.Keccak256([]byte(digest))
	//sig
	sig := result.Data.Signature
	sigBuf, err := base64.URLEncoding.DecodeString(sig)
	if err != nil {
		log.Error("sig decode", "err", err)
		return nil
	}

	sigPubkey, err := crypto.Ecrecover(digestHash, sigBuf)
	if err != nil {
		log.Error("ec recover sig pubkey", "err", err)
		return nil
	}
	if bytes.Compare(certPubkey, sigPubkey) != 0 {
		log.Error("cert pubkey != sig pubkey")
		return nil
	}

	//验证证书
	if !types.CheckIAASCert(result.Data.UnityCert, "tcadmTrustHosts.crt") {
		log.Error("check IAAS cert fail")
		return nil
	}

	if result.Status == 300 {
		result.Data.Hosts = []string{"300"}
	}

	//set cache
	tcUpdate.TrustHosts.Lock.Lock()
	tcUpdate.TrustHosts.CurChainID = result.Data.ChainId
	tcUpdate.TrustHosts.HostList = result.Data.Hosts
	tcUpdate.TrustHosts.Lock.Unlock()
	//copy
	hosts = make([]string, 0, len(result.Data.Hosts))
	hosts = append(hosts, result.Data.Hosts...)
	return hosts
}

//parse payload and extension
//flag_payload_extension
func ParseData2(data []byte) (payload []byte, err error) {
	if bytes.HasPrefix(data, ExtensionFlagByte) {
		dataString := string(data)
		splits := strings.Split(dataString, "_")
		if len(splits) < 3 {
			return payload, nil
		}
		payloadString := splits[1]
		if payloadString == "" {
			payload = nil
		} else {
			payload, err = hex.DecodeString(payloadString)
			if err != nil {
				return nil, fmt.Errorf("hex decode string:%v", err)
			}
		}

		return payload, err
	}

	return data, nil //raw payload
}

func ParseData(data []byte) (payload []byte, meta utils.Meta, err error) {
	if bytes.HasPrefix(data, ExtensionFlagByte) {
		dataString := string(data)
		splits := strings.Split(dataString, "_")
		if len(splits) != 3 {
			return nil, nil, fmt.Errorf("splits != 3, dataString:%s", dataString)
		}

		payloadString := splits[1]
		if payloadString == "" {
			payload = nil
		} else {
			payload, err = hex.DecodeString(payloadString)
			if err != nil {
				return nil, nil, fmt.Errorf("hex decode string:%v", err)
			}
		}

		metaString := splits[2]
		metaByte, err := hex.DecodeString(metaString)
		if err != nil {
			return nil, nil, fmt.Errorf("hex decode meta:%v", err)
		}
		err = json.Unmarshal(metaByte, &meta)
		if err != nil {
			return nil, nil, fmt.Errorf("json unmarshal meta:%v", err)
		}

		return payload, meta, err
	}

	return data, nil, nil //raw payload
}

// assemble flag, payload and extension
// eg: flag_payload_extension
func AssemblePayload(payload []byte, meta utils.Meta) (data []byte, err error) {
	//ExtensionFlag := "26cad4db1a82fab9e41bf5c0fbb4937a27b93786a4f27d3b9704805f698d3e65"
	metaByts, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	metaStr := hex.EncodeToString(metaByts)

	payloadStr := hex.EncodeToString(payload)

	dataStr := ExtensionFlag + "_" +
		payloadStr +
		"_" + metaStr
	data = []byte(dataStr)

	return data, nil
}

//consortium conf set
func ConsortiumSet(localhostCert string) (ConsortiumConf consortiumConf, RootIdCaMap map[string]string, UserCertPublicKeyMap map[string]struct{}, Cert string, CertRootId string, err error) {
	RootIdCaMap = make(map[string]string)
	UserCertPublicKeyMap = make(map[string]struct{})

	cf := ConsortiumDir + "consortium.conf"
	cfBuf, err := ioutil.ReadFile(cf)
	if err != nil {
		err = fmt.Errorf("read consortium conf: %s", err.Error())
		return
	}

	err = json.Unmarshal(cfBuf, &ConsortiumConf)
	if err != nil {
		err = fmt.Errorf("unmarshal consortium conf: %s", err.Error())
		return
	}

	subKeyidReg := regexp.MustCompile(`X509v3 Subject Key Identifier: \n(.*)`)
	authKeyidReg := regexp.MustCompile(`keyid:(.*)`)
	certReg := regexp.MustCompile(`-{5}BEGIN CERTIFICATE-{5}\s+([^-]+)-{5}END CERTIFICATE-{5}`)
	for _, org := range ConsortiumConf.Orgs {
		for _, ca := range org.NodeCa {
			var caFile []byte
			caFile, err = ioutil.ReadFile(ConsortiumDir + ca)
			if err != nil {
				err = fmt.Errorf("read ca file:%s", err.Error())
				return
			}

			caStr := string(caFile)
			//node ca keyid
			keyids := subKeyidReg.FindAllString(caStr, -1)
			if len(keyids) == 0 {
				err = errors.New("node keyid not find")
				return
			}
			keyid := keyids[len(keyids)-1]
			if keyid == "" {
				err = errors.New("node keyid is empty")
				return
			}
			if len(keyid) < 59 {
				err = errors.New("node keyid len error")
				return
			}

			subKeyid := keyid[len(keyid)-59:]

			//root ca certificate block
			certBlocks := certReg.FindAllString(caStr, -1)
			if len(certBlocks) == 0 {
				err = errors.New("root cert block not find")
				return
			}

			certBlock := strings.Join(certBlocks, "\n")

			RootIdCaMap[subKeyid] = certBlock
		}

		for _, uCa := range org.UserCa {
			//exec command
			var out []byte
			shell := fmt.Sprintf("openssl x509 -pubkey -noout -in %s", ConsortiumDir+uCa)
			out, err = exec.Command("/bin/bash", "-c", shell).CombinedOutput()
			if err != nil {
				err = fmt.Errorf("openssl x509 -pubkey command exec:%s", err.Error())
				return
			}

			ss := strings.Split(string(out), "-----")
			noReturn := strings.Replace(ss[2], "\n", "", -1)

			var byts []byte
			byts, err = base64.StdEncoding.DecodeString(noReturn)
			if err != nil {
				err = fmt.Errorf("base64 decodeString:%s", err.Error())
				return
			}

			hexStr := hex.EncodeToString(byts)

			pubHexStr := hexStr[len(hexStr)-130:]
			var pubHexBytes []byte
			pubHexBytes, err = hex.DecodeString(pubHexStr)
			if err != nil {
				err = fmt.Errorf("hex decode string:%s", err.Error())
				return
			}
			var pubKeyObj *ecdsa.PublicKey
			pubKeyObj, err = crypto.UnmarshalPubkey(pubHexBytes)
			if err != nil {
				err = fmt.Errorf("unmarshal pubkey:%s", err.Error())
				return
			}
			pubkey := crypto.CompressPubkey(pubKeyObj)

			pubkeyString := hex.EncodeToString(pubkey)
			log.Info("public key string", "string", pubkeyString)

			UserCertPublicKeyMap[pubkeyString] = struct{}{}
		}
	}

	if localhostCert == "" {
		var cert []byte
		cert, err = ioutil.ReadFile(ConsortiumDir + LocalCert)
		if err != nil {
			err = fmt.Errorf("read cert file:%s", err.Error())
			return
		}

		if len(cert) == 0 {
			err = errors.New("cert empty")
			return
		}
		localhostCert = string(cert)
	}

	//get certificate blocks from chained cert
	certBlocks := certReg.FindAllString(localhostCert, -1)
	if len(certBlocks) == 0 {
		err = errors.New("cert certificate block not find")
		return
	}

	Cert = certBlocks[0]

	certkeyids := authKeyidReg.FindAllString(localhostCert, -1)
	if len(certkeyids) == 0 {
		//log.Error("keyid not find")
		err = errors.New("cert auth keyid not find")
		return
	}
	rootIdString := certkeyids[0]
	if rootIdString == "" {
		err = errors.New("cert auth keyid is empty")
		return
	}
	if len(rootIdString) < 59 {
		err = errors.New("cert auth keyid len error")
		return
	}
	rootId := rootIdString[len(rootIdString)-59:]
	if _, find := RootIdCaMap[rootId]; !find {
		err = errors.New("localhost cert no authorized by node ca")
		return
		//return consortiumConf{}, nil, nil, nil, nil, errors.New("localhost cert no authorized by node ca")
	}
	//var certKeyidRlp []byte
	//certKeyidRlp, err = rlp.EncodeToBytes(rootId)
	//if err != nil {
	//	err = fmt.Errorf("cert encode bytes:%s", err.Error())
	//	return
	//	//return consortiumConf{}, nil, nil, nil, nil, fmt.Errorf("cert encode bytes:%s", err.Error())
	//}

	CertRootId = rootId

	//log.Trace("!!!!!!!!!!!!consortium conf", "conf", ConsortiumConf)
	//log.Trace("!!!!!!!!!!!!consortium conf", "RootIdCaMap", RootIdCaMap)
	//log.Trace("!!!!!!!!!!!!consortium conf", "cert", certBlocks)
	//log.Trace("!!!!!!!!!!!!consortium conf", "certRootId", rootId)

	//log.Trace("!!!!!!!!!!!!!!!! baaphome", "baaphome", conf.BaapHome)
	return
}

func GetToken(privk *ecdsa.PrivateKey, srcChainId, destChainId string, tokenType int) (token string) {
	iaasServer := Config.String("iaas")
	proxyServer := Config.String("blockFreeCloud")

	//todo debug-------
	//iaasServer := "http://10.0.0.62:8080"
	//proxyServer := "http://10.0.0.128:9999"
	//log.Debug("!!!!!!!!!!!!!!!!!!!!!! getToken", "iaas", iaasServer, "proxy", proxyServer)
	//todo-----------
	if iaasServer == "" || proxyServer == "" {
		log.Info("server is empty", "iaasServer", iaasServer, "proxyServer", proxyServer)
		return
	}

	if (srcChainId == "" || destChainId == "") && tokenType != RabbitMqTokenType {
		log.Warn("empty", "srcChainId", srcChainId, "destChainId", destChainId)
		return
	}

	switch tokenType {
	case TrustTxTokenType, XChainTokenType:
	case RabbitMqTokenType: // rabbit mq
	default:
		log.Error("token type err", "type", tokenType)
		return
	}

	//先去缓存中获取
	blockFreeCloud.RLock() //add lock====
	if tokenType == RabbitMqTokenType {
		if bfc, find := blockFreeCloud.BFCMap[rabbitMQTokenKey]; find && bfc.Token != "" && time.Now().Unix()+10 < bfc.Expired {
			//log.Debug("------------------------<<<<<<<<<", "expired", bfc.Expired, "token", bfc.Token)
			blockFreeCloud.RUnlock() // note: before return must unlock
			return bfc.Token
		}
	} else {
		if bfc, find := blockFreeCloud.BFCMap[destChainId]; find && bfc.Token != "" && time.Now().Unix()+10 < bfc.Expired {
			//log.Debug("------------------------<<<<<<<<<", "expired", bfc.Expired, "token", bfc.Token)
			blockFreeCloud.RUnlock() // note: before return must unlock
			return bfc.Token
		}
	}
	blockFreeCloud.RUnlock() //unlock====

	pubKeyObj := privk.PublicKey
	pubK := fmt.Sprintf("%x", crypto.FromECDSAPub(&pubKeyObj))
	now := time.Now().Unix()
	timestamp := strconv.FormatInt(now, 10)
	salt := randSeq(16)
	digest := strconv.Itoa(tokenType) + srcChainId + destChainId + pubK + salt + timestamp

	sig, err := crypto.Sign(crypto.Keccak256([]byte(digest)), privk)
	if err != nil {
		log.Error("sign error", "err", err)
		return
	}

	params := struct {
		Type        int    `json:"type"`
		FromChainId string `json:"fromChainId"`
		ToChainId   string `json:"toChainId"`
		Publickey   string `json:"publickey"`
		Salt        string `json:"salt"`
		Timestamp   string `json:"timestamp"`
		Signature   string `json:"signature"`
	}{
		tokenType,
		srcChainId,
		destChainId,
		pubK,
		salt,
		timestamp,
		common.Bytes2Hex(sig),
	}

	paramByts, err := json.Marshal(params)
	if err != nil {
		log.Error("marshal param", "err", err)
		return
	}

	buf := bytes.NewBuffer(paramByts)
	req, err := http.NewRequest(
		"POST",
		//Config.String("iaas")+"/rest/chain/bftoken",
		iaasServer+"/rest/chain/bftoken",
		buf,
	)
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	client.Timeout = time.Millisecond * 800
	resp, err := client.Do(req)
	if err != nil {
		log.Warn("client do", "err", err)
		return
	}
	defer resp.Body.Close()

	respBuffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read resp body", "err", err)
		return
	}

	type Data struct {
		Token     string `json:"token"`
		UnityCert string `json:"unityCert"`
		Salt      string `json:"salt"`
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
	}

	type RspBody struct {
		Status int               `json:""`
		Meta   map[string]string `json:"meta"`
		Data   Data              `json:"data"`
	}
	var result RspBody
	err = json.Unmarshal(respBuffer, &result)
	if err != nil {
		log.Error("unmarshal resp body", "err", err)
		return
	}

	log.Debug("resp result", "status", result.Status, "token", result.Data.Token)

	//parse token
	p := new(jwt.Parser)
	tokenObj, _, err := p.ParseUnverified(result.Data.Token, jwt.MapClaims{})
	if err != nil {
		log.Error("parse token unverified", "err", err)
		return
	}

	var e int64
	if claims, ok := tokenObj.Claims.(jwt.MapClaims); ok {
		exp, find := claims["exp"]
		if !find {
			log.Error("exp no exist in claims")
			return
		}

		ex, ok := exp.(float64)
		if !ok {
			log.Error("exp assertion fail", "type", reflect.TypeOf(exp).String())
			return
		}

		e = int64(ex)
	}

	//log.Debug("%%%%%%%%%%%%%%%%%%%%", "blockFreeCloud", blockFreeCloud.BFCMap)

	//存入缓存
	blockFreeCloud.Lock() //add lock=====
	if tokenType == RabbitMqTokenType {
		blockFreeCloud.BFCMap[rabbitMQTokenKey] = BlockFreeToken{destChainId, result.Data.Token, e}
	} else {
		blockFreeCloud.BFCMap[destChainId] = BlockFreeToken{destChainId, result.Data.Token, e}
	}
	blockFreeCloud.Unlock() //unlock=====

	return result.Data.Token

}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[mathRand.Intn(len(letters))]
	}
	return string(b)
}
