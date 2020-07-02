package vm

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"github.com/tjfoc/gmsm/sm2"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	"pdx-chain/log"
	"pdx-chain/params"
	"pdx-chain/pdxcc/conf"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/pdxcc/util"
	"pdx-chain/rlp"
)

// PDXSafe implement PrecompiledContract interface
type PDXSafe struct{}

var JWT = "PDX_SAFE_AUTHZ"
var ACTION = "PDX_SAFE_ACTION"

// RequiredGas return different gas according to different action
// get action DONT require gas
// set action gas is calaculated by input size
func (c *PDXSafe) RequiredGas(input []byte) uint64 {
	return uint64(len(input)/192) * params.Bn256PairingPerPointGas
}

type domainItem struct {
	Name    string
	Type    string
	Desc    string
	Creator EC_PUBKEY
	Version uint64

	CAuth []EC_PUBKEY
	RAuth []EC_PUBKEY
	UAuth []EC_PUBKEY
	DAuth []EC_PUBKEY
}

type keyItem struct {
	Key     string
	Value   []byte
	Desc    string
	Creator EC_PUBKEY
	Version uint64

	RAuth []EC_PUBKEY
	UAuth []EC_PUBKEY
	DAuth []EC_PUBKEY
}

const keyCountPrefix = "PDXS_NKEY."
const PUBK_HEX_LEN = 66
const PUBK_BYTE_LEN = PUBK_HEX_LEN / 2

type EC_PUBKEY [PUBK_BYTE_LEN]byte

type jwtData struct {
	sender_key EC_PUBKEY     `json:"sk,omitempty"` // 被授权方的public key
	domain     string        `json:"d,omitempty"`  // domain`json:"aud,omitempty"`
	key        string        `json:"k,omitempty"`  // key
	action     []protos.PDXS `json:"a,omitempty`   // “C”, “R”, "U", or "D"
	nonce      string        `json:"n,omitempty"`  // nounce
	author_key EC_PUBKEY     `json:"ak",omitempty` // 授权方的public key
	sequence   uint64        `json:"s,omitempty"`  // 授权方为本次授权赋予的seq number
}

func parseJWT(stoken string, ctx *PrecompiledContractContext) (*jwtData, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(stoken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:

		if ctx.Evm.chainConfig.Sm2Crypto {
			if _, ok := token.Method.(*jwt.SigningMethodSM2); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			if token.Header["alg"] != "SM2" {
				return nil, fmt.Errorf("invalid signing alg:%v, only SM2 is prefered", token.Header["alg"])
			}
		} else {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			if token.Header["alg"] != "ES256" {
				return nil, fmt.Errorf("invalid signing alg:%v, only ES256 is prefered", token.Header["alg"])
			}
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_sec ret_key")
		ak, ok := token.Claims.(jwt.MapClaims)["ak"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: no \"ak\" in jwt payload")
		}
		hexKey, ok := ak.(string)
		if !ok || len(hexKey) != PUBK_HEX_LEN {
			return nil, fmt.Errorf("PDXSafe: invalid \"ak\" in jwt payload")
		}
		if ctx.Evm.chainConfig.Sm2Crypto {
			return sm2.Decompress(common.Hex2Bytes(hexKey)), nil
		}
		return crypto.DecompressPubkey(common.Hex2Bytes(hexKey))
	})

	if err != nil {
		return nil, err
	}

	pld := &jwtData{}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// parse sender public key(required)
		tmp, ok := claims["sk"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"sk\" in jwt payload")
		}
		s, ok := tmp.(string)
		if !ok || len(s) != PUBK_HEX_LEN {
			return nil, fmt.Errorf("PDXSafe: invalid \"sk:%s\" in jwt payload", s)
		}
		copy(pld.sender_key[:], common.Hex2Bytes(s))

		// parse domain(required)
		tmp, ok = claims["d"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"d\" in jwt payload")
		}
		s, ok = tmp.(string)
		if !ok || len(s) == 0 {
			return nil, fmt.Errorf("PDXSafe: invalid \"d:%s\" in jwt payload", s)
		}
		pld.domain = s

		// parse key(optional)
		tmp, ok = claims["k"]
		if ok {
			s, ok = tmp.(string)
			if ok && len(s) != 0 {
				pld.key = s
			}
		}

		// parse action(required)
		tmp, ok = claims["a"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"a\" in jwt payload")
		}
		s, ok = tmp.(string)
		if !ok || len(s) == 0 {
			return nil, fmt.Errorf("PDXSafe: invalid \"a:%v\" in jwt payload", s)
		}
		for _, a := range []byte(s) {
			// a should be 'C'/'R'/'U'/'D'
			switch a {
			case 'C':
				pld.action = append(pld.action, protos.PDXS_CREATE_KEY)
			case 'R':
				if pld.key == "" {
					pld.action = append(pld.action, protos.PDXS_GET_DOMAIN)
				} else {
					pld.action = append(pld.action, protos.PDXS_GET_KEY)
				}
			case 'U':
				if pld.key == "" {
					pld.action = append(pld.action, protos.PDXS_UPDATE_DOMAIN)
				} else {
					pld.action = append(pld.action, protos.PDXS_UPDATE_KEY)
				}
			case 'D':
				if pld.key == "" {
					pld.action = append(pld.action, protos.PDXS_DELETE_DOMAIN)
				} else {
					pld.action = append(pld.action, protos.PDXS_DELETE_KEY)
				}

			default:
				return nil, fmt.Errorf("PDXSafe: invalid \"a:%s\" in jwt payload", s)
			}
		}

		// parse nonce(required)
		tmp, ok = claims["n"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"n\" in jwt payload")
		}
		f, ok := tmp.(string)
		if !ok || len(f) == 0 {
			return nil, fmt.Errorf("PDXSafe: invalid \"n:%v\" in jwt payload", tmp)
		}
		pld.nonce = f

		// parse author pub key(required)
		tmp, ok = claims["ak"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"ak\" in jwt payload")
		}
		s, ok = tmp.(string)
		if !ok || len(s) != PUBK_HEX_LEN {
			return nil, fmt.Errorf("PDXSafe: invalid \"ak:%v\" in jwt payload", tmp)
		}
		copy(pld.author_key[:], common.Hex2Bytes(s))

		// parse sequence(required)
		tmp, ok = claims["s"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: nil \"s\" in jwt payload")
		}
		u, ok := tmp.(float64)
		if !ok || u == 0 {
			return nil, fmt.Errorf("PDXSafe: invalid \"s:%v\" in jwt payload", tmp)
		}
		pld.sequence = uint64(u)

	} else {
		return nil, fmt.Errorf("PDXSafe: invalid claims in jwt payload")
	}

	return pld, nil
}

func matchAction(a protos.PDXS, actions []protos.PDXS) bool {
	for _, action := range actions {
		if action == a {
			return true
		}
	}
	return false
}

func (c *PDXSafe) Run(ctx *PrecompiledContractContext, input []byte, extra map[string][]byte) ([]byte, error) {
	// used for estimating gas
	if extra[conf.TxEstimateGas] != nil {
		log.Error("PDXSafe estimate err", "err", extra[conf.TxEstimateGas])
		return nil, nil
	}

	tx := &protos.Transaction{}

	err := proto.Unmarshal(input, tx)
	if err != nil {
		log.Error("PDXSafe input format error", "error", err)
		return nil, err
	}
	if tx.Type != types.Transaction_invoke {
		log.Error("PDXSafe: invalid tx type", "type", tx.Type)
		return nil, err
	}
	invocation := &protos.Invocation{}
	err = proto.Unmarshal(tx.Payload, invocation)
	if err != nil {
		log.Error("proto unmarshal invocation error", "err", err)
		return nil, err
	}
	byteAction := invocation.Meta[ACTION]
	if len(byteAction) != 1 {
		return nil, errors.New("PDXSafe: invalid PDX_SAFE_ACTION type")
	}
	action := protos.PDXS(byteAction[0])

	byteParam := invocation.Args[0]

	byteToken := invocation.Meta[JWT]
	var pld *jwtData
	if action != protos.PDXS_CREATE_DOMAIN &&
		action != protos.PDXS_GET_DOMAIN &&
		action != protos.PDXS_GET_KEY {
		pld, err = parseJWT(string(byteToken), ctx)
		if err != nil {
			return nil, err
		}
		//jwt.New(jwt.SigningMethodES256)
		if !matchAction(action, pld.action) {
			return nil, fmt.Errorf("PDXSafe: token action:%v doesn't match PDXSafe action:%v", pld.action, action)
		}
	}

	switch action {
	case protos.PDXS_CREATE_DOMAIN:
		return createDomain(byteParam, ctx, extra)
	case protos.PDXS_UPDATE_DOMAIN:
		return updateDomain(pld, byteParam, ctx, extra)
	case protos.PDXS_DELETE_DOMAIN:
		return deleteDomain(pld, byteParam, ctx, extra)
	case protos.PDXS_GET_DOMAIN:
		return getDomain(byteParam, ctx)
	case protos.PDXS_CREATE_KEY:
		return createKey(pld, byteParam, ctx, extra)
	case protos.PDXS_UPDATE_KEY:
		return updateKey(pld, byteParam, ctx, extra)
	case protos.PDXS_DELETE_KEY:
		return deleteKey(pld, byteParam, ctx, extra)
	case protos.PDXS_GET_KEY:
		return getKey(byteParam, ctx)

	default:
		return nil, errors.New("PDXSafe: invalid operation")
	}
}

func _setState(ctx *PrecompiledContractContext, key string, value []byte) {
	keyHash := util.EthHash([]byte(key))
	ctx.Evm.StateDB.SetState(ctx.Contract.Address(), keyHash, util.EthHash(value))
}

func _setPDXState(ctx *PrecompiledContractContext, value []byte) {
	keyHash := util.EthHash(value)
	ctx.Evm.StateDB.SetPDXState(ctx.Contract.Address(), keyHash, value)
}

func _getState(ctx *PrecompiledContractContext, key string) common.Hash {
	keyHash := util.EthHash([]byte(key))
	hash := ctx.Evm.StateDB.GetState(ctx.Contract.Address(), keyHash)
	return hash
}

func _getPDXState(ctx *PrecompiledContractContext, key common.Hash) []byte {
	return ctx.Evm.StateDB.GetPDXState(ctx.Contract.Address(), key)
}

func _getDomainState(ctx *PrecompiledContractContext, _domain string) (*domainItem, error) {
	domainkey := _getState(ctx, _domain)
	if (domainkey == common.Hash{}) {
		return nil, errors.New("PDXSafe: domain doesn't exist")
	}
	_dm := _getPDXState(ctx, domainkey)
	var di domainItem
	if err := rlp.DecodeBytes(_dm, &di); err != nil {
		return nil, err
	}
	return &di, nil
}

func _setDomainState(ctx *PrecompiledContractContext, di *domainItem) error {
	_domain, err := rlp.EncodeToBytes(di)
	if err != nil {
		return err
	}
	_setState(ctx, di.Name, _domain)
	_setPDXState(ctx, _domain)
	return nil
}

func _delDomainState(ctx *PrecompiledContractContext, di *domainItem) {
	keyHash := util.EthHash([]byte(di.Name))
	ctx.Evm.StateDB.SetState(ctx.Contract.Address(), keyHash, common.Hash{})
}

func createDomain(param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}
	author_key := EC_PUBKEY{}
	copy(author_key[:], pubkey)
	di := &protos.DomainItem{}
	err := proto.Unmarshal(param, di)
	if err != nil {
		log.Error("PDXSafe Unmarshal DomainItem error", "error", err)
		return nil, err
	}
	//domain doesn't exist
	if di.Name == "" {
		return nil, errors.New("PDXSafe: invalid domain name")
	}
	domainKey := _getState(ctx, di.Name)
	if (domainKey != common.Hash{}) {
		return nil, errors.New("PDXSafe: domain already exists")
	}

	d := &domainItem{}
	d.Name = di.Name
	d.Type = di.Type
	d.Desc = di.Desc
	d.Creator = author_key
	d.Version = 1
	var tmp EC_PUBKEY
	for _, pubKey := range di.Cauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Cauth pubkey")
		}
		copy(tmp[:], pubKey)
		d.CAuth = append(d.CAuth, tmp)
	}

	d.UAuth = append(d.UAuth, author_key)
	for _, pubKey := range di.Uauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Uauth pubkey")
		}
		copy(tmp[:], pubKey)
		d.UAuth = append(d.UAuth, tmp)
	}
	d.DAuth = append(d.DAuth, author_key)
	for _, pubKey := range di.Dauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Dauth pubkey")
		}
		copy(tmp[:], pubKey)
		d.DAuth = append(d.DAuth, tmp)
	}

	err = _setDomainState(ctx, d)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func updateDomain(pld *jwtData, param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	// check pld.sender is the same as tx sender
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}

	if !bytes.Equal(pld.sender_key[:], pubkey) {
		return nil, errors.New("PDXSafe: jwt sender doesn't match tx sender")
	}
	// check this jwt is for domain or key operation
	if pld.key != "" {
		return nil, errors.New("PDXSafe: this jwt is for key operation")
	}

	d := &protos.DomainItem{}
	err := proto.Unmarshal(param, d)
	if err != nil {
		log.Error("PDXSafe Unmarshal DomainItem error", "error", err)
		return nil, err
	}
	// check domain name in jwt equals domain name in param
	if d.Name != pld.domain {
		return nil, errors.New("PDXSafe: domain name in jwt dones't match domain name in param")
	}

	di, err := _getDomainState(ctx, d.Name)
	if err != nil {
		return nil, err
	}

	if d.Version != di.Version {
		return nil, errors.New("PDXSafe: domain version does not match")
	}

	var canUpdate bool
	// check author is in UAuth list
	for _, pubKey := range di.UAuth {
		if pld.author_key == pubKey {
			canUpdate = true
			break
		}
	}
	if !canUpdate {
		return nil, errors.New("PDXSafe: ak(author public key) is not in domain Uauth list")
	}

	if len(d.Type) != 0 {
		di.Type = d.Type
	}
	if len(d.Desc) != 0 {
		di.Desc = d.Desc
	}
	var tmp EC_PUBKEY
	if len(d.Cauth) != 0 {
		di.CAuth = nil
	}

	for _, pubKey := range d.Cauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Cauth pubkey")
		}
		copy(tmp[:], pubKey)
		di.CAuth = append(di.CAuth, tmp)
	}

	if len(d.Uauth) != 0 {
		// (trick) always keep owner of the domain
		di.UAuth = di.UAuth[0:1]
	}
	for _, pubKey := range d.Uauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Uauth pubkey")
		}
		copy(tmp[:], pubKey)
		if tmp == di.UAuth[0] {
			continue
		}
		di.UAuth = append(di.UAuth, tmp)
	}

	if len(d.Dauth) != 0 {
		// (trick) always keep owner of the domain
		di.DAuth = di.DAuth[0:1]
	}
	for _, pubKey := range d.Dauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Dauth pubkey")
		}
		copy(tmp[:], pubKey)
		if tmp == di.DAuth[0] {
			continue
		}
		di.DAuth = append(di.DAuth, tmp)
	}

	di.Version++

	err = _setDomainState(ctx, di)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func deleteDomain(pld *jwtData, param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	// check pld.sender is the same as tx sender
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}

	if !bytes.Equal(pld.sender_key[:], pubkey) {
		return nil, errors.New("PDXSafe: jwt sender doesn't match tx sender")
	}

	// check this jwt is for domain or key operation
	if pld.key != "" {
		return nil, errors.New("PDXSafe: this jwt is for key operation")
	}

	d := &protos.DomainItem{}
	err := proto.Unmarshal(param, d)
	if err != nil {
		log.Error("PDXSafe Unmarshal DomainItem error", "error", err)
		return nil, err
	}
	// check domain name in jwt equals domain name in param
	if d.Name != pld.domain {
		return nil, errors.New("PDXSafe: domain name in jwt dones't match domain name in param")
	}

	di, err := _getDomainState(ctx, d.Name)
	if err != nil {
		return nil, err
	}
	var canDelete bool
	// check author is in UAuth list
	for _, pubKey := range di.DAuth {
		if pld.author_key == pubKey {
			canDelete = true
		}
	}
	if !canDelete {
		return nil, errors.New("PDXSafe: ak(author public key) is not in domain Dauth list")
	}

	hv := _getState(ctx, keyCountPrefix+d.Name)
	if (hv == common.Hash{0}) {
		_delDomainState(ctx, di)
	}
	return nil, nil
}

func getDomain(param []byte, ctx *PrecompiledContractContext) ([]byte, error) {
	d := &protos.DomainItem{}
	err := proto.Unmarshal(param, d)
	if err != nil {
		log.Error("PDXSafe Unmarshal DomainItem error", "error", err)
		return nil, err
	}

	di, err := _getDomainState(ctx, d.Name)
	if err != nil {
		return nil, err
	}

	ret := &protos.DomainItem{}
	ret.Name = di.Name
	ret.Type = di.Type
	ret.Desc = di.Desc
	ret.Version = di.Version

	for _idx, _ := range di.CAuth {
		ret.Cauth = append(ret.Cauth, di.CAuth[_idx][:])
	}
	for _idx, _ := range di.UAuth {
		ret.Uauth = append(ret.Uauth, di.UAuth[_idx][:])
	}
	for _idx, _ := range di.DAuth {
		ret.Dauth = append(ret.Dauth, di.DAuth[_idx][:])
	}

	hv := _getState(ctx, keyCountPrefix+di.Name)
	ret.Keycount = big.NewInt(1).SetBytes(hv.Bytes()).Uint64()

	buf, err := proto.Marshal(ret)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func _getKeyState(ctx *PrecompiledContractContext, _domain, _key string) (*keyItem, error) {
	valueHash := _getState(ctx, _domain+_key)
	if (valueHash == common.Hash{}) {
		return nil, errors.New("PDXSafe: key doesn't exist")
	}
	//key buf
	_kb := _getPDXState(ctx, valueHash)
	var ki keyItem
	if err := rlp.DecodeBytes(_kb, &ki); err != nil {
		return nil, err
	}
	return &ki, nil
}

func _setKeyState(ctx *PrecompiledContractContext, _domain string, ki *keyItem) error {
	_kb, err := rlp.EncodeToBytes(*ki)
	if err != nil {
		return err
	}
	_setState(ctx, _domain+ki.Key, _kb)
	_setPDXState(ctx, _kb)
	return nil
}

func _delKeyState(ctx *PrecompiledContractContext, _domain string, ki *keyItem) {
	keyHash := util.EthHash([]byte(_domain + ki.Key))
	ctx.Evm.StateDB.SetState(ctx.Contract.Address(), keyHash, common.Hash{})
}

// createKey required params are (Domain, Key, Value), optional AuthKey, don't need Version
// everyone can create a key under a public domain
// Auth owner can create a key under a protected domain
// someone has the signature of Auth's can create under a protected domain
func createKey(pld *jwtData, param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	// check pld.sender is the same as tx sender
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}

	if !bytes.Equal(pld.sender_key[:], pubkey) {
		return nil, errors.New("PDXSafe: jwt sender doesn't match tx sender")
	}

	di, err := _getDomainState(ctx, pld.domain)
	if err != nil {
		return nil, err
	}
	var canCreate bool
	if len(di.CAuth) == 0 {
		// everyone can create key
		canCreate = true
	} else {
		// check author is in UAuth list
		for _, pubKey := range di.CAuth {
			if pld.author_key == pubKey {
				canCreate = true
				break
			}
		}
	}

	if !canCreate {
		return nil, errors.New("PDXSafe: ak(author public key) is not in domain Uauth list")
	}

	ki := &protos.KeyItem{}
	err = proto.Unmarshal(param, ki)
	if err != nil {
		log.Error("PDXSafe Unmarshal KeyItem error", "error", err)
		return nil, err
	}
	// jwt is only for pld.key if it is not nil
	if pld.key != "" && pld.key != ki.Key {
		return nil, errors.New("PDXSafe: key in jwt doesn't match it from keyitem")
	}

	valueHash := _getState(ctx, di.Name+ki.Key)
	if (valueHash != common.Hash{}) {
		return nil, errors.New("PDXSafe: key already exists")
	}

	k := &keyItem{}
	k.Key = ki.Key
	k.Value = ki.Value
	k.Desc = ki.Desc
	k.Version = 1
	var tmp EC_PUBKEY
	k.UAuth = append(k.UAuth, pld.sender_key)
	for _, pubKey := range ki.Uauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid key Uauth pubkey")
		}
		copy(tmp[:], pubKey)
		k.UAuth = append(k.UAuth, tmp)
	}
	k.DAuth = append(k.DAuth, pld.sender_key)
	for _, pubKey := range ki.Dauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid domain Dauth pubkey")
		}
		copy(tmp[:], pubKey)
		k.DAuth = append(k.DAuth, tmp)
	}

	err = _setKeyState(ctx, di.Name, k)
	if err != nil {
		return nil, err
	}

	hv := _getState(ctx, keyCountPrefix+di.Name)
	v := big.NewInt(0).SetBytes(hv.Bytes())
	v.Add(v, big.NewInt(1))
	keyHash := util.EthHash([]byte(keyCountPrefix + di.Name))
	ctx.Evm.StateDB.SetState(ctx.Contract.Address(), keyHash, common.BigToHash(v))

	return nil, nil
}

func updateKey(pld *jwtData, param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	// check pld.sender is the same as tx sender
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}

	if !bytes.Equal(pld.sender_key[:], pubkey) {
		return nil, errors.New("PDXSafe: jwt sender doesn't match tx sender")
	}
	// check domain is valid
	di, err := _getDomainState(ctx, pld.domain)
	if err != nil {
		return nil, err
	}
	// check key is valid
	ki := &protos.KeyItem{}
	err = proto.Unmarshal(param, ki)
	if err != nil {
		log.Error("PDXSafe Unmarshal KeyItem error", "error", err)
		return nil, err
	}
	//
	if pld.key != ki.Key {
		return nil, errors.New("PDXSafe: key in jwt doesn't match it from keyitem")
	}

	k, err := _getKeyState(ctx, di.Name, ki.Key)
	if err != nil {
		return nil, err
	}

	if k.Version != ki.Version {
		return nil, errors.New("PDXSafe: key version does not match")
	}

	var canUpdate bool
	// check author is in UAuth list
	for _, pubKey := range k.UAuth {
		if pld.author_key == pubKey {
			canUpdate = true
			break
		}
	}
	if !canUpdate {
		return nil, errors.New("PDXSafe: ak(author public key) is not in key Uauth list")
	}

	k.Value = ki.Value
	k.Desc = ki.Desc
	var tmp EC_PUBKEY
	if len(ki.Uauth) != 0 {
		k.UAuth = k.UAuth[0:1]
	}
	for _, pubKey := range ki.Uauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid key Uauth pubkey")
		}
		copy(tmp[:], pubKey)
		if tmp == k.UAuth[0] {
			continue
		}
		k.UAuth = append(k.UAuth, tmp)
	}
	if len(ki.Dauth) != 0 {
		k.DAuth = k.DAuth[0:1]
	}
	for _, pubKey := range ki.Dauth {
		if len(pubKey) != PUBK_BYTE_LEN {
			return nil, errors.New("PDXSafe: invalid key Dauth pubkey")
		}
		copy(tmp[:], pubKey)
		if tmp == k.DAuth[0] {
			continue
		}
		k.DAuth = append(k.DAuth, tmp)
	}

	k.Version++

	err = _setKeyState(ctx, di.Name, k)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func deleteKey(pld *jwtData, param []byte, ctx *PrecompiledContractContext, extra map[string][]byte) ([]byte, error) {
	// check pld.sender is the same as tx sender
	pubkey := extra["baap-sender-pubk"]
	if len(pubkey) != PUBK_BYTE_LEN {
		return nil, errors.New("PDXSafe: invalid public key in extra[\"baap-sender-pubk\"")
	}

	if !bytes.Equal(pld.sender_key[:], pubkey) {
		return nil, errors.New("PDXSafe: jwt sender doesn't match tx sender")
	}
	// check domain is valid
	di, err := _getDomainState(ctx, pld.domain)
	if err != nil {
		return nil, err
	}
	// check key is valid
	ki := &protos.KeyItem{}
	err = proto.Unmarshal(param, ki)
	if err != nil {
		log.Error("PDXSafe Unmarshal KeyItem error", "error", err)
		return nil, err
	}
	//
	if pld.key != ki.Key {
		return nil, errors.New("PDXSafe: key in jwt doesn't match it from keyitem")
	}

	k, err := _getKeyState(ctx, di.Name, ki.Key)
	if err != nil {
		return nil, err
	}
	var canDelete bool
	// check author is in UAuth list
	for _, pubKey := range k.DAuth {
		if pld.author_key == pubKey {
			canDelete = true
			break
		}
	}
	if !canDelete {
		return nil, errors.New("PDXSafe: ak(author public key) is not in key Dauth list")
	}

	_delKeyState(ctx, di.Name, k)

	hv := _getState(ctx, keyCountPrefix+di.Name)
	v := big.NewInt(0).SetBytes(hv.Bytes())
	v.Sub(v, big.NewInt(1))
	keyHash := util.EthHash([]byte(keyCountPrefix + di.Name))
	ctx.Evm.StateDB.SetState(ctx.Contract.Address(), keyHash, common.BigToHash(v))

	return nil, nil
}

func getKey(param []byte, ctx *PrecompiledContractContext) ([]byte, error) {
	k := &protos.KeyItem{}
	err := proto.Unmarshal(param, k)
	if err != nil {
		log.Error("PDXSafe Unmarshal KeyItem error", "error", err)
		return nil, err
	}

	ki, err := _getKeyState(ctx, k.Dname, k.Key)
	if err != nil {
		return nil, err
	}

	k.Desc = ki.Desc
	k.Value = ki.Value
	k.Version = ki.Version

	for _idx, _ := range ki.UAuth {
		k.Uauth = append(k.Uauth, ki.UAuth[_idx][:])
	}
	for _idx, _ := range ki.DAuth {
		k.Dauth = append(k.Dauth, ki.DAuth[_idx][:])
	}

	buf, err := proto.Marshal(k)
	if err != nil {
		return nil, err
	}
	return buf, nil

}
