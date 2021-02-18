package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/core/types"
	"pdx-chain/crypto"
	client2 "pdx-chain/utopia/utils/client"
	"strconv"
	"sync"
	"time"
)

const (
	//host         = "http://39.100.34.235:30074"
	//host         = "http://39.100.93.177:30036"
	//host = "http://10.0.0.203:33333"
	host1 = "http://192.168.3.6:8546"
	//host1 = "http://127.0.0.1:8547"
	//host          = "http://10.0.0.112:22222"
	sendDuration  = time.Minute * 1
	nonceTicker   = time.Minute * 10 //多久重新查一次nonce （note:此处应该大于1处， 否则ticker会不断执行）
	sleepDuration = time.Minute * 1  //查完nonce后休眠时间（1处）
	txNum         = -1
)

var privKeys = []string{
	//"a9f1481564399443bb39188d3f8da55585c9238ab175010b81e7a28956559381", //7DE
	//"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84",
	//"009f1dfe52be1015970d9087de0ad2a98f4c68f610711d1533aa21a71ccc8f4a", //from:0x00CFc66BBD69fb964df1C9782062D4282FfF0cda
	//"69192206e447dbc8b6627d7beb540e6c606c5b94afa9ebc00734ff404a1e5617",

	//"d29ce71545474451d8292838d4a0680a8444e6e4c14da018b4a08345fb2bbb84", //086
	//"72660cbaef2ca607751c0514922ca995a566f5bd508ccfae4896265db856d115", //sm2
	//"e7bb5c0bc8456fe0c2af281fe5753095c5150fbdbd622776330cece60e9feaec",
	//"5dc4c81695f4606394b535e44315f9a444b7f860f1951f4eb5eb8fdf59913b0a",
	//新超给的私钥
	//"281ad87954531160ad1fc57800db50b6ff079b27d66e20218125d84446554592",
	//"3f11dd8b4f14498c35030947dc34f87a13259fda566a6fa60091953a24dc71ae",
	//"4b5965545ab5153b0c7a34df6d82e830113d94877329088f492d98a222b27c88",
	//"5efb903a2d86f08b5c1ff57608f8ad4c80cda1a5f44367d67100c46608810779",
	//"1888510b0afc8faac0b801e5c4213e596e28d887cb019d5babd8ecd8c0c50d84",
	//"7346e1f34c5c3f48ad48f0198d06cceffe1c87ab3b4062e892c186cd99646866",
	//"5dc4c81695f4606394b535e44315f9a444b7f860f1951f4eb5eb8fdf59913b0a",
	//"4df63e1c54355237f2893431a5a66af2a5097da5092d470ec480bf41e8914f56",
	//"03598508ee5aef3ac3d19b6e47f94377df44f447090918786df2464d31d0b751",
	//"f5e166ecd28e2c0afdf7f764e9844588ec1b7a34c77c7be3b673f5a4d9b31f6a",
	//"1598e1b4f6b2f911142fe5720554360bdd45c5487ff922e06930513c25c87577",



	//补nonce
	//"68d4fbe1493f5389a33980e5d2ac0090fe354b5d05e38b6b4034b0616d6f164",
	//"6e4843c7bc72932e2b44b2836aa5eb224476ef353133892a39ad3cf5348ef900",
	//"727ad32681d6d007e063e19a0ceb8aec38cd1d5c1a3b6889f569d34f094bf3af",
	//"514a6138c365ba53bb7e8c9f73f4b0065ba3440d359980f89e81dbb66807e37a",
	//"61c13380be886da2d304bb8d62990f1c0f58af1c3c37505e97d2477fcde00d75",
	//"2ee805397d7404c345bad19148091f68a19d0ee145e995211f1d023b092b8b0e",
	//"25d492ba08c07f4c8182366ad1fdc7ab6c728ace0e1d5aff656e166291d88c9b",
	//"1ff1d107f87a66a60fbf6991ad406ff9d1af93cba1ebfeaff69f5ebf6502b8f0",
	//"042689b8f459d6402b6dc37eb8ae18879c624f8da0dc3b0726b855eb9129ec35",
	//"5e9491f03ab78c44e8de73f3361f6a4ab2ca311819e19c20a968353a14fe28d7",
	//"4093b168f98fb02908dd8085b6c69899b92e813bf9258a917073256ccac5f5ed",
	//"37210ba7eaa1675f5573594a3765389dc57e0ce162916cb61a8b7f68a323c313",
	//"41fdd3443249039c064c36e61d91508237571b7ba52156fb3d2ba7137292e07f",
	//"02d20c590d4facfb0997edec834f02f91233e6ddb0becdd0c58c6ec8924b4f79",
	//"776adb8f2e411303e9d34b3a4d6974aaf65d23458a64b2d4a0a8aea384c7b10d",
	//"012051b9b3e58552c15a564d942283a2ff48aa2358b131a33256291757ed0c34",
	//"29ae7b35f6177521750af13d5ce05366d04ff8099f26bb2e74bba697c27c168a",
	//"479c1c4a37d50600286c4175e0a6db26313a6226eb66ec489bec9ed8701f7bae",
	//"783e52f91bb91a5a93bcf06c6052c46ca1f97f53f535ef5fdd0038cabc477226",
	//"054090203935377b53f773665dffd124a287152f376e495ce76d5a3854443303",
	//"65ad54761f6c4dadba32f1f4c7d81de0885cb32a2fb56695a5d7f7e9946553cd",
	//"04786bf12f2a22d56ac7e98d21a8d0c797e8c201546d265f84d61ca1a68a2025",
	//"2d76f3c0e3ebb8221117285caca06cd056a8adccb86ea4fe7501f95a04dde35a",
	//"4f4f4649ec7db9e2b596a011ddcd83349b7a815bb863ff325abc6d78163a57e3",
	//"26657bf9e21c3930d49bf0fc5d6a7834c3bf73d9f47c72a7a00129cd535317ea",
	//"1aecd4c41caed151dea2183874cf4c7088fd0b98fdb0f38b375c0f85572a1841",
	//"1dee92abe9d63f037714d03470f282f94d55d1df89cfd5b0ade679fccb4b3b04",
	//"10a718aa2fcb0d28c70dceb7b613f2ecce5d59c506424f900cc53f51a95e6bf1",
	//"4a4b6af5ae7fea13e12409f2be56dac9cd9cdbb2f541bb544a17cb4c6a80b13c",
	//"0af93ccc99eb586a67594dc3cfcadc94561750e08ab2fca9ae5a0fc97b34a259",
	//"134c9d45c9b4f089267617c4d172f54d7548d1aa5086975617b6f70ad4959718",
	//"323141003702f6cbdb0cd9dd481311d3e275d0f133d1fbd900c9b680cfc7da2f",
	//"35cd51e3d196e01fcff47afe26641d3f908e887b07d1b8e5f4c3bd4e82fa300e",
	//"06cc126598e2c55419c6d90e978be72338e889a04d611ce0c0ae22e861ed6333",
	//"762f7961b29db622ecbd5254214e224c8e3ce40b36940ea0b9dc4a7c71d28fea",
	//"65f2104a0bdb08ceef8b28356e7e896ff719e48af7df5db04e7ef94c34fd19e4",
	//"39b572a10c8e01a36585f9a8c19bd256de9183a11ea2ad92503c75c900c651f7",
	//"56415cc2b0f0c63d38363c18cbdbb2d79044d2374ec354c07162668f9eac300c",
	//"08c29a2d13e7290de9a8642ed08994e8f1c6c76912e7d6432ea3b2300a8ac643",
	//"7234d3a75b967096e8860ddadd80826af6a4b5f842267f5d496070f4e4e9cfc2",
	//"24a0309ecee990c471abd9f991ed53c90273999bae52a4d04d45197e506b08d3",
	//"7ddf59f03b2a585c840d26c10e4b5dfe77a5e5e9bf81f9e4794ed91f7d6059a9",
	//"0f51634758aa8dc2fb954047aa7b0993875cfc92200211051582503a68a72288",
	//"742ece35a1929acb183c793729ef76eeffba9ea23360d70b872a06701c8cf62a",
	//"0d5fa5429e3eb12655d9a343a3b1157377175b7308a7a1e64326c515310c85b5",
	//"5078666fb015a2142bd4fc38013b60fcba92275f042e97fd946ce804a3fdfd14",
	//"734dd67f03442f792bd167bbbb2aef7c33ea65bef802bfe268b0a6d7c4b6695a",
	//"2e768d4eb3dcb0d3c06c5fbed97e17be3fdf3f7c50f6a297607ac3e17fb09f20",
	//"1097f9697148f93412496481b2cf8c85be6b68fc137838144961612bcca48641",
	//"21210e4cf7cdb22e72841113de16c4698fae14848b15c7ecf676512f1d6391b1",
	//"1a52f6a5b1f7a167bbb3a24f38375d090a538905827d93f4513174408404ed52",
	//"54c472fa29b6409e1ab7151e1c92745137745a9a2e2354267b1da3363bad1058",
	//"22c3bbb27d76c637c9d0dd02c76c5929f835448a6c4161ff8daa55e0d6242439",
	//"7b9643050747769f4d01de1bcc72638ac3d9ec2146253a931a21dea064e041bf",
	//"77b7db9acb340717bfca40914c5ef768a7ef87b4d3ca01f5c510425fa9ccc335",
	//"6b5cffddb4604ac3be4fb71b1077c78673e6fe607bd37b24e74ce97dd5789eac",
	//"0e26c73db801c26e294516cc4537edacb1539984606ea3676c30b4fbc4619723",
	//"70d2a73250554a895291034b9cc6359b15f68e5eb15e209d4bcbc139f2a93b78",
	//"293201d23c00d325687d20e8d50f7d2dcb78268c99c755e0bba939dbea23e8a5",
	//"1114aea26d215b926bf7a207c0b331cfee1c37801841de3f6338da9de0995ab4",
	//"4d5eea8af958475b43b56fa28aa69b92fb4b94745cf526807327e0b59ccbc9c7",
	//"5bc1aecd6f4e78cdf42be9689c27d9ac04e872944a8195365d9e729aa08b9e10",
	//"48082d06aea944156312545b6dc5ab1d172fc8b3d775284f33dd57ad477bea40",
	"1c23b0beb3a728940b2573c9274543e8ca454f0aceacd99d60929f80a1d331e4",


}

func main() {
	log.SetFlags(log.Lmicroseconds)

	var wg sync.WaitGroup

	//contract := ggKeccak256ToAddress("8000d109DAef5C81799bC01D4d82B0589dEEDb33:testcc02:1.0.0").String()
	for i, privKey := range privKeys {
		log.Printf("i is = %d", i)
		wg.Add(1)

		go func(j int, key string) {
			defer wg.Done()

			sendTestTx(
				key,
				//"a41368620000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000a68692c2067756f74616f00000000000000000000000000000000000000000000",
				strconv.Itoa(j)+":", j)
		}(i, privKey)
	}

	wg.Wait()

}

func sendTestTx(privKey, flag string, x int) {
	//proxy := "http://10.0.0.241:9999"
	//token := "eyJhbGciOiJFUzI1NiJ9.eyJpYXQiOjE1ODk5NzQ2NzIsIkZSRUUiOiJUUlVFIn0.sWYZ6awd8yRNX9iG5o7Ls4Uop5nfZrUtuprx9hwKxw2fS5zQtxunY11bccJ_h29VfnFMqyvaVvI9Tu3R0USlwQ"
	//if client, err := eth.Connect("http://utopia-chain-1001:8545", proxy, token); err != nil {
	//if client, err := client2.Connect("http://127.0.0.1:8547"); err != nil {
		if client, err := client2.Connect(host1); err != nil {

			//	if client, err := eth.Connect("http://10.0.0.219:33333"); err != nil {
		fmt.Printf(err.Error())
		return
	} else {
		//addr: 0xa2b67b7e4489549b855e423068cbded55f7c5daa
		//r := rand.Intn(len(privKeys))

		pKey, err := crypto.HexToECDSA(privKey)
		if err != nil {
			return
		}

		//sbin, ok := new(big.Int).SetString(srcKey, 16)
		//if !ok {
		//	return
		//}
		//sKey := sm2.InitKey(sbin)

		//pKey1, _ := crypto.HexToECDSA(privKey)
		////pbin, _ := new(big.Int).SetString(privKey, 16)
		//pKey1 := sm2.InitKey(pbin)

		from := crypto.PubkeyToAddress(pKey.PublicKey)
		fmt.Printf("from:%s\n", from.String())
		//privK: d6bf45db5f7e1209cdf58c0cca2f28516bdf4ce07cad211cf748f31874084b5e
		if nonce, err := client.EthClient.NonceAt(context.TODO(), from,nil); err != nil {

			//if nonce, err := client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
			fmt.Printf("nonce err: %s", err.Error())
			//return
		} else {
			fmt.Println("nonce", nonce)

			amount := big.NewInt(0).Mul(big.NewInt(0), big.NewInt(1e18))

			//amount := big.NewInt(0).Mul(big.NewInt(1), big.NewInt(1e18))
			gasLimit := uint64(21000)
			gasPrice := new(big.Int).Mul(big.NewInt(1e9), big.NewInt(4000)) //todo 此处很重要，不可以太低，可能会报underprice错误，增大该值就没有问题了

			timer := time.NewTimer(sendDuration)
			ticker := time.NewTicker(nonceTicker)
			log.Println(flag+"start:", time.Now().String())
			i := 0
			//nonce = 4161
			for {
				select {
				case <-timer.C:
					log.Println(flag + "time is over")
					log.Println(flag+"end:", time.Now().String())
					return
				case <-ticker.C:
					fmt.Println("sleep.........")
					time.Sleep(sleepDuration)
					fmt.Println("get nonce again")
					if nonce, err = client.EthClient.NonceAt(context.TODO(), from, nil); err != nil {
						fmt.Printf("nonce again err: %s", err.Error())
						return
					}

					fmt.Println("nonce again is", nonce)

				default:
					//time.Sleep(5*time.Millisecond)
					//fmt.Println(flag+"nonce is what", ":", nonce, "from:", from.String())
					//rand.Seed(time.Now().Unix())
					//n1 := rand.Int31n(9)
					//n2 := rand.Int31n(9)
					to := common.HexToAddress(fmt.Sprintf("0x08b299d855734914cd7b19eea60c84b%d256840%d", x, x+1))
					//to := common.HexToAddress("0x48c60bdeed69477460127c28b27e43a7ad442b9a")
					//fmt.Printf("to:%s\n", to.String())
					var data []byte
					//for h := 0; h <= 1024; h++ {
					//	data = append(data, 1)
					//}

					tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data)
					//tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, nil)

					//signer := types.HomesteadSigner{}
					//signer := types.NewSm2Signer(big.NewInt(777))
					signer := types.NewEIP155Signer(big.NewInt(738))

					//sm := (*ecdsa.PrivateKey)(pKey1)
					//signedTx, _ := types.SignTx(tx, signer, sm)
					signedTx, err := types.SignTx(tx, signer, pKey)

					if err != nil {
						fmt.Println("types.SignTx", err)
						return
					}
					if txhash, err := client.SendRawTransaction(context.TODO(), signedTx); err != nil {
						fmt.Println(flag+"send raw transaction err:", err.Error())
						nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						return
					} else {
						fmt.Printf("Transaction hash: %s, %d, %s\n", txhash.String(), nonce, from.String())
						nonce++
						//nonce, _ = client.EthClient.NonceAt(context.TODO(), from, nil)
						i++
						//time.Sleep(1*time.Second)
						if txNum != -1 && i >= txNum {
							return
						}
					}
				}
			}
		}
	}
}
