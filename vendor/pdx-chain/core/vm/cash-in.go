package vm

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/log"
	"pdx-chain/pdxcc/conf"
	"pdx-chain/pdxcc/util"
	"pdx-chain/utopia/utils"
)

type CashIn struct{}

type CashInData struct {
	BankReceipt string `json:"bank_receipt"` //银行收据
	Amount *big.Int `json:"amount"` //充值钱数
	To common.Address `json:"to"` //给谁充值
}

type CashInRecord struct {
	BankReceipt string         `json:"bank_receipt"` // 银行收据
	Amount      *big.Int       `json:"amount"` // 充值数
	To          common.Address `json:"to"` // 充值给谁
	PublicAccount common.Address `json:"public_account"` // 公账户地址
	txHash common.Hash `json:"tx_hash"` // 充值时间
}

func (x *CashIn) RequiredGas(input []byte) uint64 {
	return 0
}

func (x *CashIn) Run(ctx *PrecompiledContractContext, input []byte, extra map[string][]byte) ([]byte, error) {
	//避免记录的累次加钱数量有出入。单独记录
	//if ctx.Contract.value.Cmp(big.NewInt(0)) == +1 {
	//	log.Error("amount in transaction should equal to zero")
	//	return nil, errors.New("amount in transaction should equal to zero")
	//}

	cashInData := new(CashInData)
	err := json.Unmarshal(input, cashInData)
	if err != nil {
		log.Error("json unmarshal cashInData", "err", err)
		return nil, err
	}
	log.Debug("^^^^^^^^^^^^^^^cash in data", "receipt", cashInData.BankReceipt, "amount", cashInData.Amount, "to", cashInData.To)

	contractAddr := ctx.Contract.Address()

	if ctx.Evm.ChainConfig().IsEIP158(ctx.Evm.BlockNumber) {
		ctx.Evm.StateDB.SetNonce(contractAddr, 1)
	}

	// 获取公账户
	publicAccount := getPublicAccount(ctx)
	if publicAccount == (common.Address{}) {
		log.Error("public account empty")
		return nil, err
	}
	log.Debug("^^^^^^^^^^^^^^public account", "addr", publicAccount.String())

	// 验证交易来自公账户
	if ctx.Contract.CallerAddress != publicAccount {
		log.Error("illegal public Account", "sender", ctx.Contract.CallerAddress)
		return nil, errors.New("illegal public Account")
	}

	// 给相应的账户加钱
	ctx.Evm.StateDB.AddBalance(cashInData.To, cashInData.Amount)

	// 记录用户本次加钱
	tokeyHash := getCashInRecordKeyHash(cashInData.To)
	byts := ctx.Evm.StateDB.GetPDXState(contractAddr, tokeyHash)
	cashInRecords := make([]CashInRecord, 0)
	if len(byts) > 0 {
		err = json.Unmarshal(byts, &cashInRecords)
		if err != nil {
			log.Error("json unmarshal cashInRecords", "err", err)
			return nil, err
		}
	}

	txHash := common.BytesToHash(extra[conf.BaapTxid][:])
	record := CashInRecord{
		BankReceipt:cashInData.BankReceipt,
		Amount:cashInData.Amount,
		To:cashInData.To,
		PublicAccount:ctx.Contract.CallerAddress,
		txHash:txHash,
	}
	cashInRecords = append(cashInRecords, record)

	for k, r := range cashInRecords {
		log.Debug("^^^^^^^^^^^^^^^^^^^^cash in record",
			"----------->k", k,
			"receipt", r.BankReceipt,
			"amount", r.Amount,
			"to", r.To.String(),
			"pA", r.PublicAccount.String(),
			"txHash", r.txHash.String(),
		)
	}


	value, err := json.Marshal(cashInRecords)
	ctx.Evm.StateDB.SetPDXState(contractAddr, tokeyHash, value)

	// 记录累次加钱数额
	totalKeyHash := getCashInTotalKeyHash()
	byts = ctx.Evm.StateDB.GetPDXState(contractAddr, totalKeyHash)
	total := new(big.Int).Add(new(big.Int).SetBytes(byts), cashInData.Amount)
	ctx.Evm.StateDB.SetPDXState(contractAddr, totalKeyHash, total.Bytes())
	log.Debug("^^^^^^^^^^^^^^^^^^^^cash in", "total", total.String())

	return nil, nil
}

func getCashInTotalKeyHash() common.Hash {
	return util.EthHash([]byte("pdx:cashIn:total"))
}

func getCashInRecordKeyHash(caller common.Address) common.Hash {
	key := fmt.Sprintf("pdx:cashIn:%s", caller.String())
	return util.EthHash([]byte(key))
}

func getPublicAccount(ctx *PrecompiledContractContext) common.Address {
	var publicAccount common.Address
	// 首先去库里获取
	publicAccountKeyHash := getPublicAccountKeyHash()
	byts := ctx.Evm.StateDB.GetPDXState(utils.PublicAccountVote, publicAccountKeyHash)
	if len(byts) > 0 {
		return common.BytesToAddress(byts)
	}
	// 否则从初始化配置中获取
	publicAccount = ctx.Evm.chainConfig.Utopia.PublicAccount
	return publicAccount
}