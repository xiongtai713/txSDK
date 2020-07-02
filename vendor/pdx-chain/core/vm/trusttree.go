package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/big"
	"pdx-chain/common"
	"pdx-chain/log"
	"pdx-chain/pdxcc/conf"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/pdxcc/util"
)

type TrustTree struct{}

type TrustTxData struct {
	PreTrustChainID     string      `json:"preTrustChainID"` // pre under layer trust chain ID
	CurTrustChainID     string      `json:"curTrustChainID"` // cur under layer trust chain ID
	ChainID             string      `json:"chainID"`         // service chain ID
	CommitBlockNo       uint64      `json:"commitBlockNo"`
	CommitBlockHash     common.Hash `json:"commitBlockHash"`
	PrevCommitBlockHash common.Hash `json:"prevCommitBlockHash"`
	NodeAddress         string      `json:"nodeAddress"`
}

func (x *TrustTree) RequiredGas(input []byte) uint64 {
	return 0
}

func (x *TrustTree) Run(ctx *PrecompiledContractContext, input []byte, extra map[string][]byte) ([]byte, error) {
	log.Info("trust tree precompile contract run")

	ptx := &protos.Transaction{}
	err := proto.Unmarshal(input, ptx)
	if err != nil {
		log.Error("proto unmarshal input", "err", err)
		return nil, err
	}

	inv := &protos.Invocation{}
	err = proto.Unmarshal(ptx.Payload, inv)
	if err != nil {
		log.Error("proto unmarshal payload", "err", err)
		return nil, err
	}

	//ctx.Evm.StateDB.CreateAccount(ctx.Contract.Address())
	if ctx.Evm.ChainConfig().IsEIP158(ctx.Evm.BlockNumber) {
		ctx.Evm.StateDB.SetNonce(ctx.Contract.Address(), 1)
	}

	//if !ctx.Evm.StateDB.Exist(ctx.Contract.Address()) {
	//	log.Info("iiiiiiiiiiiiiiiiiiii")
	//	ctx.Evm.StateDB.CreateAccount(ctx.Contract.Address())
	//	if ctx.Evm.ChainConfig().IsEIP158(ctx.Evm.BlockNumber) {
	//		log.Info("oooooooooooo")
	//		ctx.Evm.StateDB.SetNonce(ctx.Contract.Address(), 1)
	//	}
	//}

	txData := &TrustTxData{}
	for _, v := range inv.Args {
		err = json.Unmarshal(v, txData)
		if err != nil {
			log.Error("json unmarshal arg", "err", err)
			return nil, err
		}

		//maxNum
		var maxNum *big.Int
		maxNumKey := fmt.Sprintf("%s-trustTransactionHeight", txData.ChainID)
		maxNumKeyHash := util.EthHash([]byte(maxNumKey + conf.PDXKeyFlag))
		maxNumSt := ctx.Evm.StateDB.GetPDXState(ctx.Contract.Address(), maxNumKeyHash)

		log.Info("trust tx data info", "chainID", txData.ChainID, "pre TC chainID", txData.PreTrustChainID, "cur TC chainID", txData.CurTrustChainID, "block hash", txData.CommitBlockHash, "block num", txData.CommitBlockNo)
		if len(maxNumSt) > 0 {
			maxNum = big.NewInt(0).SetBytes(maxNumSt)
			log.Info("block num", "chainID", txData.ChainID, "maxNum", maxNum, "payload", txData.CommitBlockNo)
			//判断块号是否连续
			if new(big.Int).Add(maxNum, big.NewInt(1)).Cmp(new(big.Int).SetUint64(txData.CommitBlockNo)) == 0 {
				curCommitBlockKey := fmt.Sprintf("%s:%s", txData.ChainID, maxNum.String())
				curCommitBlockKeyHash := util.EthHash([]byte(curCommitBlockKey + conf.PDXKeyFlag))
				curCommitBlockHashSt := ctx.Evm.StateDB.GetPDXState(ctx.Contract.Address(), curCommitBlockKeyHash)
				log.Info("compare pre commit block hash", "pre", txData.PrevCommitBlockHash, "cur", common.BytesToHash(curCommitBlockHashSt))
				if bytes.Equal(txData.PrevCommitBlockHash.Bytes(), curCommitBlockHashSt) {
					setHashAndNumber(ctx, txData.ChainID, txData.CommitBlockNo, txData.CommitBlockHash, maxNumKeyHash)
				} else {
					//hash 不连续
					log.Warn("block hash not continuous")
				}

			} else {
				if new(big.Int).Add(maxNum, big.NewInt(1)).Cmp(new(big.Int).SetUint64(txData.CommitBlockNo)) > 0 {
					//回滚
					priorCurCommitBlockKey := fmt.Sprintf("%s:%d", txData.ChainID, txData.CommitBlockNo-1)
					priorCurCommitBlockKeyHash := util.EthHash([]byte(priorCurCommitBlockKey + conf.PDXKeyFlag))
					priorCurCommitBlockHashSt := ctx.Evm.StateDB.GetPDXState(ctx.Contract.Address(), priorCurCommitBlockKeyHash)
					if len(priorCurCommitBlockHashSt) == 0 {
						log.Info("roll-back to origin", "num", txData.CommitBlockNo)
						setHashAndNumber(ctx, txData.ChainID, txData.CommitBlockNo, txData.CommitBlockHash, maxNumKeyHash)
						return nil, nil
					}
					log.Info("roll-back compare pre commit block hash", "pre", txData.PrevCommitBlockHash, "cur", common.BytesToHash(priorCurCommitBlockHashSt))
					if bytes.Equal(txData.PrevCommitBlockHash.Bytes(), priorCurCommitBlockHashSt) {
						backCommitBlockKey := fmt.Sprintf("%s:%d", txData.ChainID, txData.CommitBlockNo)
						backCommitBlockKeyHash := util.EthHash([]byte(backCommitBlockKey + conf.PDXKeyFlag))
						backCommitBlockHashSt := ctx.Evm.StateDB.GetPDXState(ctx.Contract.Address(), backCommitBlockKeyHash)
						log.Info("roll-back compare back commit block hash", "payload", txData.CommitBlockHash.Bytes(), "backCommitBlockHashSt", backCommitBlockHashSt)
						if bytes.Equal(txData.CommitBlockHash.Bytes(), backCommitBlockHashSt) {
							log.Warn("roll-back block hash no change, so do not rewrite")
							return nil, nil
						}
						setHashAndNumber(ctx, txData.ChainID, txData.CommitBlockNo, txData.CommitBlockHash, maxNumKeyHash)
					} else {
						log.Warn("roll-back: block hash not continuous")
					}

				} else {
					//块号过大
					log.Warn("block number too big")
				}
			}

		} else {
			//第一次存储
			log.Info("max num empty", "num", maxNumSt)
			setHashAndNumber(ctx, txData.ChainID, txData.CommitBlockNo, txData.CommitBlockHash, maxNumKeyHash)
		}

	}

	return nil, nil
}

func setHashAndNumber(ctx *PrecompiledContractContext, chainID string, commitBlockNumber uint64, commitBlockHash common.Hash, maxNumKeyHash common.Hash) {
	//commit block hash
	commitBlockKey := fmt.Sprintf("%s:%d", chainID, commitBlockNumber)
	commitBlockKeyHash := util.EthHash([]byte(commitBlockKey + conf.PDXKeyFlag))
	//ctx.Evm.StateDB.SetState(ctx.Contract.Address(), commitBlockKeyHash, util.EthHash(commitBlockHash.Bytes()))
	ctx.Evm.StateDB.SetPDXState(ctx.Contract.Address(), commitBlockKeyHash, commitBlockHash.Bytes())
	//maxNum
	//ctx.Evm.StateDB.SetState(ctx.Contract.Address(), maxNumKeyHash, util.EthHash(new(big.Int).SetUint64(commitBlockNumber).Bytes()))
	ctx.Evm.StateDB.SetPDXState(ctx.Contract.Address(), maxNumKeyHash, new(big.Int).SetUint64(commitBlockNumber).Bytes())
}
