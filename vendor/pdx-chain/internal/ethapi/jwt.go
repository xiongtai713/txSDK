package ethapi

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/proto"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/publicBC"
	"pdx-chain/core/types"
	"pdx-chain/core/vm"
	"pdx-chain/crypto"
	"pdx-chain/ethdb"
	"pdx-chain/log"
	"pdx-chain/pdxcc"
	pb "pdx-chain/pdxcc/protos"
	"pdx-chain/pdxcc/util"
	"pdx-chain/utopia"
	"pdx-chain/utopia/utils"
)

func ParseToken(tx *types.Transaction, stateDB vm.StateDB) (token string, err error) {
	//if cc tx
	var ok bool
	var jwtBuf []byte
	to := tx.To()
	if to != nil {
		if pdxcc.CanExec(*to) {
			//is cc
			token, err := parseCCToken(tx.Data())
			if err != nil {
				return "", fmt.Errorf("parse cc token:%v", err)
			}
			return token, nil
		}

		//keyHash := utils.CCKeyHash
		//ccBuf := stateDB.GetPDXState(*to, keyHash)
		ccBuf := stateDB.GetCode(*to)
		if len(ccBuf) != 0 {
			var deploy pb.Deployment
			err = proto.Unmarshal(ccBuf, &deploy)
			if err == nil {
				//is cc
				token, err := parseCCToken(tx.Data())
				if err != nil {
					return "", fmt.Errorf("parse cc token:%v", err)
				}
				return token, nil
			}
		}
	}

	_, meta, err := utopia.ParseData(tx.Data())
	if err != nil {
		return "", fmt.Errorf("parse data:%v", err)
	}
	if meta == nil {
		return "", fmt.Errorf("meta == nil")
	}

	jwtBuf, ok = meta["jwt"]
	if !ok {
		return "", fmt.Errorf("jwt not in meta, meta:%v", meta)
	}

	return string(jwtBuf), nil
}

func parseCCToken(payload []byte) (token string, err error) {
	//payload, _, err := utopia.ParseData(tx.Data()) //maybe parse
	txPb := &pb.Transaction{}
	err = proto.Unmarshal(payload, txPb)
	if err != nil {
		return "", fmt.Errorf("proto unmarshal tx:%v", err)
	}

	var jwtBuf []byte
	var ok bool
	txType := txPb.Type
	switch txType {
	case types.Transaction_deploy:
		deploy := pb.Deployment{}
		err = proto.Unmarshal(txPb.Payload, &deploy)
		if err != nil {
			return "", fmt.Errorf("proto unmarshal deploy:%v", err)
		}
		jwtBuf, ok = deploy.Payload.Meta["jwt"]
		if !ok {
			return "", fmt.Errorf("jwt not in deploy meta")
		}

	case types.Transaction_invoke: //start stop withdraw
		invocation := pb.Invocation{}
		err = proto.Unmarshal(txPb.Payload, &invocation)
		if err != nil {
			return "", fmt.Errorf("proto unmarshal invocation:%v", err)
		}

		jwtBuf, ok = invocation.Meta["jwt"]
		if !ok {
			return "", fmt.Errorf("jwt not in invocation meta")
		}

	}

	return string(jwtBuf), nil
}

func CheckFreeze(tx *types.Transaction, stateDB vm.StateDB, from common.Address) bool {
	//var signer types.Signer = types.HomesteadSigner{}
	//if tx.Protected() {
	//	signer = types.NewEIP155Signer(tx.ChainId())
	//}
	//from, err := types.Sender(signer, tx)
	//if err != nil {
	//	log.Error("check freeze get sender", "err", err)
	//	return false
	//}
	fromHash := util.EthHash(from.Bytes())
	r := stateDB.GetPDXState(utils.AccessCtlContract, fromHash)
	if len(r) > 0 && string(r) == "1" {
		log.Trace("string(r) == 1")
		return true
	}

	return false
}

func CheckToken(tx *types.Transaction, stateDB vm.StateDB) bool {
	if tx.To() != nil && *tx.To() == utils.AccessCtlContract {
		log.Trace("access ctl token verify")
		tokenString, err := ParseToken(tx, stateDB)
		if err != nil {
			log.Error("parse token", "err", err)
			return false
		}

		if !checkJWT(tokenString, "a") {
			return false
		}
		return true
	}

	//DappAuth:T UserAuth:T Deploy:d regular Tx:u/d
	if utopia.ConsortiumConf.DappAuth && utopia.ConsortiumConf.UserAuth {
		tokenString, err := ParseToken(tx, stateDB)
		if err != nil {
			log.Error("parse token", "err", err)
			return false
		}
		//if deploy contract tx must role d
		if tx.To() == nil || *tx.To() == utils.CCBaapDeploy {
			if !checkJWT(tokenString, "d") {
				return false
			}
		} else {
			//if not , regular tx must role u at least
			if !checkJWT(tokenString, "u/d") {
				return false
			}
		}
	}

	//DappAuth:F UserAuth:T Deploy:u/d regular Tx:u/d
	if !utopia.ConsortiumConf.DappAuth && utopia.ConsortiumConf.UserAuth {
		tokenString, err := ParseToken(tx, stateDB)
		if err != nil {
			log.Error("parse token", "err", err)
			return false
		}

		//if deploy contract tx must role d
		if tx.To() == nil || *tx.To() == utils.CCBaapDeploy {
			if !checkJWT(tokenString, "u/d") {
				return false
			}
		} else {
			//if not , regular tx must role u at least
			if !checkJWT(tokenString, "u/d") {
				return false
			}
		}
	}

	//DappAuth:T UserAuth:F Deploy:d regular Tx:-
	if utopia.ConsortiumConf.DappAuth && !utopia.ConsortiumConf.UserAuth {
		//if deploy contract tx must role d
		if tx.To() == nil || *tx.To() == utils.CCBaapDeploy {
			tokenString, err := ParseToken(tx, stateDB)
			if err != nil {
				log.Error("parse token", "err", err)
				return false
			}

			if !checkJWT(tokenString, "d") {
				return false
			}
		}
	}

	//DappAuth:F UserAuth:F Deploy:- regular Tx:-

	return true
}

func checkJWT(tokenString string, roleLimit string) (success bool) {
	if tokenString == "" {
		log.Error("tokenString empty")
		return false
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if token.Header["alg"] != "ES256" {
			return nil, fmt.Errorf("invalid signing alg:%v, only ES256 is prefered", token.Header["alg"])
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("token claims type error")
		}

		ak, ok := claims["ak"]
		if !ok {
			return nil, fmt.Errorf("ak no exist in claims")
		}
		hexPubKey, ok := ak.(string)
		if !ok || len(hexPubKey) != vm.PUBK_HEX_LEN {
			return nil, fmt.Errorf("ak format error")
		}

		//check public key
		_, ok = utopia.UserCertPublicKeyMap[hexPubKey]
		if !ok {
			return nil, fmt.Errorf("ak no exist in user cert public key")
		}

		return crypto.DecompressPubkey(common.Hex2Bytes(hexPubKey))
	})

	if err != nil {
		log.Error("jwt parse", "err", err)
		return false
	}

	if claims, success := token.Claims.(jwt.MapClaims); success && token.Valid {
		limit, success := claims["l"].(float64)
		if !success {
			log.Error("l not correct")
			return false
		}
		if !checkLimit(tokenString, int64(limit)) {
			log.Error("check limit fail")
			return false
		}

		role, success := claims["r"]
		if !success {
			log.Error("role no match", "role", role, "ok", success)

			return false
		}
		if roleLimit == "d" || roleLimit == "a" {
			if role != roleLimit {
				log.Error("role no auth", "role", role, "roleLimit", roleLimit)
				return false
			}
		} else {
			if role == "u" || role == "d" {
			} else {
				log.Error("role no exist", "role", role)
				return false
			}
		}

	} else {
		log.Error("token invalid")
		return false
	}

	return true
}

func checkLimit(tokenString string, limit int64) bool {
	db := *ethdb.ChainDb
	tokenHash := util.EthHash([]byte(tokenString))
	has, err := db.Has(tokenHash.Bytes())
	if err != nil {
		log.Error("db has", "err", err)
		return false
	}

	currentBlockNum := public.BC.CurrentBlock().Number()
	if !has {
		expiredBlockNum := big.NewInt(0).Add(big.NewInt(limit), currentBlockNum)
		err = db.Put(tokenHash.Bytes(), expiredBlockNum.Bytes())
		if err != nil {
			log.Error("db put tokenHash", "err", err)
			return false
		}
	} else {
		numByts, err := db.Get(tokenHash.Bytes())
		if err != nil {
			log.Error("db get tokenHash", "err", err)
			return false
		}

		expiredBlockNum := big.NewInt(0).SetBytes(numByts)
		if currentBlockNum.Cmp(expiredBlockNum) > 0 {
			log.Error("out of limit", "currentBlockNum", currentBlockNum.String(), "expiredBlockNum", expiredBlockNum)
			return false
		}
	}

	return true
}
