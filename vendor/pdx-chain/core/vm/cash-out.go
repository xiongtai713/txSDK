package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/log"
	"pdx-chain/pdxcc/conf"
	"pdx-chain/pdxcc/util"
)

type CashOut struct{}

func (x *CashOut) RequiredGas(input []byte) uint64 {
	return 0
}

//type CashOutData struct {
//	Amount *big.Int `json:"amount"` // 提现数
//}

type CashOutRecord struct {
	Amount    *big.Int       `json:"amount"` // 提现数
	From      common.Address `json:"to"` // 提现用户
	TxHash common.Hash `json:"tx_hash"` // 提现时间
}

func (x *CashOut) Run(ctx *PrecompiledContractContext, input []byte, extra map[string][]byte) ([]byte, error) {
	// 解析payload
	//cashOutData := new(CashOutData)
	//err := json.Unmarshal(input, cashOutData)
	//if err != nil {
	//	log.Error("json unmarshal cashData", "err", err)
	//	return nil, err
	//}

	contractAddr := ctx.Contract.Address()
	if ctx.Evm.ChainConfig().IsEIP158(ctx.Evm.BlockNumber) {
		ctx.Evm.StateDB.SetNonce(contractAddr, 1)
	}

	// 判断余额是否充足
	//callerBalance := ctx.Evm.StateDB.GetBalance(ctx.Contract.CallerAddress)
	//if callerBalance.Cmp(ctx.Contract.value) < 0 {
	//	log.Error("balance not enough", "callerBalance", callerBalance.String(), "out", ctx.Contract.value.String())
	//	return nil, ErrInsufficientBalance
	//}
	//ctx.Evm.StateDB.SubBalance(ctx.Contract.CallerAddress, cashOutData.Amount) 在transfer中subbalance

	// 记录每个人的提现记录
	keyHash := getCashOutRecordKeyHash(ctx.Contract.CallerAddress)
	cashOutRecords := make([]CashOutRecord, 0)
	byts := ctx.Evm.StateDB.GetPDXState(contractAddr, keyHash)
	if len(byts) > 0 {
		err := json.Unmarshal(byts, &cashOutRecords)
		if err != nil {
			log.Error("json unmarshal cashoutRecords", "err", err)
			return nil, err
		}
	}

	txHash := common.BytesToHash(extra[conf.BaapTxid][:])
	record := CashOutRecord{
		Amount:ctx.Contract.value,
		From:ctx.Contract.CallerAddress,
		TxHash:txHash,
	}
	cashOutRecords = append(cashOutRecords, record)
	value, err := json.Marshal(cashOutRecords)
	if err != nil {
		log.Error("json marshal cashOutRecords", "err", err)
		return nil, err
	}

	ctx.Evm.StateDB.SetPDXState(contractAddr, keyHash, value)

	for k, v := range cashOutRecords {
		log.Debug("^^^^^^^^^^^^cash out record",
			"------------>k", k,
			"amount", v.Amount,
			"from", v.From.String(),
			"txHash", v.TxHash.String(),
		)
	}

	// 记录总的转出数
	totalKeyHash := getCashOutTotalKeyHash()
	byts = ctx.Evm.StateDB.GetPDXState(contractAddr, totalKeyHash)
	total := new(big.Int).Add(new(big.Int).SetBytes(byts), ctx.Contract.value)
	ctx.Evm.StateDB.SetPDXState(contractAddr, totalKeyHash, total.Bytes())
	log.Debug("^^^^^^^^^^^^^^^^^^^^cash out", "total", total.String())

	return nil, nil
}

func getCashOutTotalKeyHash() common.Hash {
	return util.EthHash([]byte("pdx:cashOut:total"))
}

func getCashOutRecordKeyHash(caller common.Address) common.Hash {
	key := fmt.Sprintf("pdx:cashOut:%s", caller.String())
	return util.EthHash([]byte(key))
}
