package main

import (
	"fmt"
	"log"
	"pdx-chain/common"
	"pdx-chain/crypto"
	"pdx-chain/crypto/jwt-go"
	"time"
)

func NewJwt(domain ,key string) string {
	now := time.Now().Unix()
	tonken := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"f": float64(now),
		"t": float64(now + 1000000000),
		//7DE
		"sk":  "03d173890122ba3066b1d20c185c65555861b7de30066c6dc854cf6cb903d6b642", //被授权方的public key
		"d":   domain,                                                               //domain名称
		"key": key,                                                                //key
		"a":   "C",                                                                  //“C”, “R”, "U", or "D" 创建，获取，更新，删除
		"n":   "1",                                                                  //nonce 没发现作用
		//860
		"ak": "0272d910dcda9d55dc0504459852c32bfbc86e3f9baddf512f90a6811452a550ba", //授权方的public key,使用这个公钥来验证jwt,并且是权限控制的签发人,必须是这个ak签发的jwt,才能修改或者创建
		"s":  float64(123),                                                         //授权方为本次授权赋予的seq number 没作用
	},
	)
	priv2 := "d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84" //860

	privateKey, err := crypto.HexToECDSA(priv2)
	if err != nil {
		log.Fatal(err)
	}

	signedString, err := tonken.SignedString(privateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(signedString)
	return signedString
}

func parseJWT(stoken string) error {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	_, err := jwt.Parse(stoken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:

		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if token.Header["alg"] != "ES256" {
			return nil, fmt.Errorf("invalid signing alg:%v, only ES256 is prefered", token.Header["alg"])
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

		return crypto.DecompressPubkey(common.Hex2Bytes(hexKey))
	})

	if err != nil {
		fmt.Println("验证失败", err)
		return err
	}
	fmt.Println("验证成功")
	return nil
}
