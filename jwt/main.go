package main


import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"pdx-chain/crypto/gmsm/sm2"
	"pdx-chain/crypto/jwt-go"
)

func main()  {
	GenTokenSM2()
}

func verifyTokenSM2(tokenString string) error {
	fmt.Println("tokenString::::::::::::::", tokenString)
	const (
		PUBK_HEX_LEN = 66
	)

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodSM2); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if token.Header["alg"] != "SM2" {
			return nil, fmt.Errorf("invalid signing alg:%v, only ES256 is prefered", token.Header["alg"])
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("token claims type error")
		}

		role, ok := claims["r"]
		if !ok {
			return nil, fmt.Errorf("r not in claims")
		}
		if role == "d" || role == "u" || role == "a" {
		} else {
			return nil, fmt.Errorf("role no auth")
		}

		ak, ok := claims["ak"]
		if !ok {
			return nil, fmt.Errorf("PDXSafe: no \"ak\" in jwt payload")
		}
		hexKey, ok := ak.(string)
		fmt.Println("hex key len:", len(hexKey), "ak:", hexKey)
		if !ok || len(hexKey) != PUBK_HEX_LEN {
			return nil, fmt.Errorf("PDXSafe: invalid \"ak\" in jwt payload")
		}
		a, err := hex.DecodeString(hexKey)
		if err != nil {
			return nil, err
		}
		pub := sm2.Decompress(a)
		fmt.Printf("auth key after decompress!!!!!!!: %x \n", sm2.Compress(pub))
		return pub, nil
		//return crypto.DecompressPubkey(common.Hex2Bytes(hexKey))
	})

	if err != nil {
		fmt.Println("jwt parse err:", err)
		return err
	}

	//jwt.SigningMethodECDSA.Verify(tokenString, privKey.Public())

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Println(claims["l"])
	} else {
		return err
	}

	return nil
}

func GenTokenSM2() (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SM2Signing, jwt.MapClaims{
		//"ak": "0390d5d104823304eb44276545ce4b3bbedba28171628a1262b0ff0b58b59e3d2f",//auth pubKey
		//"ak": "0297764c3303a0bbb2108cc09acb632b1cc27209497b2803f5e8021d1203030b2a",//auth pubKey AoPeng
		"ak": "0232c4ae2c1f1981195f9904466a39c9948fe30bbff2660be1715a4589334c74c7", //auth pubKey test
		//"ak": "03b57dbbfc029e0483faa07de15ad78261a4abf626d77cfd05d582300fdb193722",//auth pubKey gansu
		"sk": "02595d553697305c7670dfd92628e5ff68080335265edf804aea4e6e8df5112464", //sender pubKey
		"r":  "u",                                                                  //d:developer, u:end-user, a:admin
		"l":  6000000000000000,                                                     //limit
		"s":  17348,                                                                //sequence
		"n":  "eefffefreredffdsuuf2rrfdsmfljljrra",                                 //nonce
	})

	// Sign and get the complete encoded token as a string using the secret
	privKey := sm2.InitKey(big.NewInt(1))
	//privKey, err := sm2.GenerateKey()
	//if err != nil {
	//  fmt.Println("sm2 generateKey err:", err)
	//  return "", err
	//}
	tokenString, err := token.SignedString(privKey)
	if err != nil {
		fmt.Println("tokenString err:", err)
		return "", err
	}
	ecdsaPub, ok := privKey.Public().(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("assert public err")
		return "", err
	}
	sm2Pub := (sm2.PublicKey)(*ecdsaPub)
	fmt.Printf("auth key: %x \n", sm2.Compress(&sm2Pub))
	fmt.Println("tokenString;;;;;;;;;;;;;;;;;;;;;;;;", tokenString)

	return tokenString, err
}