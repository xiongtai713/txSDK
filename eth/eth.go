package eth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"

	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

// Client defines typed wrappers for the Ethereum RPC API.
type Client struct {
	rpcClient *rpc.Client
	EthClient *ethclient.Client
}

func toHexInt(n *big.Int) string {
	return fmt.Sprintf("0x%x", n)
}

// Connect creates a client that uses the given host.
func Connect(host string, proxy ... string) (*Client, error) {
	rpcClient, err := rpc.Dial(host)
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)

	return &Client{rpcClient, ethClient}, nil
}

// GetBlockNumber returns the block number.
func (ec *Client) GetBlockNumber(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := ec.rpcClient.CallContext(ctx, &result, "eth_blockNumber")
	return (*big.Int)(&result), err
}

// Message is a fully derived transaction and implements core.Message
type Message struct {
	To       *common.Address `json:"to"`
	From     *common.Address `json:"from"`
	Value    string          `json:"value"`
	GasLimit string          `json:"gas"`
	GasPrice string          `json:"gasPrice"`
	Data     []byte          `json:"data"`
}

// String
func (msg *Message) String() string {
	if str, err := json.Marshal(msg); err != nil {
		panic(err)
	} else {
		return string(str)
	}
}

// NewMessage returns the message.
func NewMessage(from *common.Address, to *common.Address, value *big.Int, gasLimit *big.Int, gasPrice *big.Int, data []byte) Message {
	return Message{
		From:     from,
		To:       to,
		Value:    toHexInt(value),
		GasLimit: toHexInt(gasLimit),
		GasPrice: toHexInt(gasPrice),
		Data:     data,
	}
}

// SendTransaction injects a transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendTransaction(ctx context.Context, tx *Message) (common.Hash, error) {
	var txHash common.Hash
	err := ec.rpcClient.CallContext(ctx, &txHash, "eth_sendTransaction", tx)
	return txHash, err
}

// SendRawTransaction injects a transaction into the pending pool for execution.
//
// If the transaction was a contract creation use the TransactionReceipt method to get the
// contract address after the transaction has been mined.
func (ec *Client) SendRawTransaction(ctx context.Context, tx *types.Transaction) (common.Hash, error) {
	var txHash common.Hash
	if data, err := rlp.EncodeToBytes(tx); err != nil {
		return txHash, err
	} else {
		err := ec.rpcClient.CallContext(ctx, &txHash, "eth_sendRawTransaction", common.ToHex(data))
		return txHash, err
	}
}

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        common.Hash     `json:"blockHash"`
	BlockNumber      *hexutil.Big    `json:"blockNumber"`
	From             common.Address  `json:"from"`
	Gas              hexutil.Uint64  `json:"gas"`
	GasPrice         *hexutil.Big    `json:"gasPrice"`
	Hash             common.Hash     `json:"hash"`
	Input            hexutil.Bytes   `json:"input"`
	Nonce            hexutil.Uint64  `json:"nonce"`
	To               *common.Address `json:"to"`
	TransactionIndex hexutil.Uint    `json:"transactionIndex"`
	Value            *hexutil.Big    `json:"value"`
	V                *hexutil.Big    `json:"v"`
	R                *hexutil.Big    `json:"r"`
	S                *hexutil.Big    `json:"s"`
}

type WithdrawRPCTransaction struct {
	RPCTx         *RPCTransaction `json:"rpc_tx"`
	Status        int             `json:"status"` //0 pending 1 success 2 commitSuccess
	CommitNumber  *hexutil.Big    `json:"commitNumber"`
	CurrentCommit *hexutil.Big    `json:"currentCommit"`
	TxMsg         string          `json:"txMsg"` //交易报文
	DstChainID    string          `json:"dst_chain_id"`
}

func (ec *Client) QueryTxByHash(ctx context.Context, txHash common.Hash) (withDrawTx WithdrawRPCTransaction, err error) {
	err = ec.rpcClient.CallContext(ctx, &withDrawTx, "eth_getWithDrawTransactionByHash", txHash)
	return

}

type BaapQ struct {
	Result string `json:"result"`
}

func (ec *Client) BaapQuery(ctx context.Context, txData hexutil.Bytes) (result string, err error) {
	err = ec.rpcClient.CallContext(ctx, &result, "eth_baapQuery", txData)
	return

}

func (ec *Client) QueryUncompletedTxByAddr(ctx context.Context, myaddr common.Address) (mytxhash []common.Hash) {
	ec.rpcClient.CallContext(ctx, &mytxhash, "eth_getUncompletedTransactionByAddress", myaddr)
	return

}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}
func (ec *Client) GetCommitBlockByNum(ctx context.Context, number *big.Int) (block map[string]interface{}, err error) {
	err = ec.rpcClient.CallContext(ctx, &block, "eth_getCommitBlockByNumber", toBlockNumArg(number), true)
	return

}

// 
func (ec *Client) CheckTransaction(ctx context.Context, receiptChan chan *types.Receipt, txHash common.Hash, retrySeconds time.Duration) {
	// check transaction receipt
	go func() {
		fmt.Printf("Check transaction: %s\n", txHash.String())
		for {
			receipt, _ := ec.EthClient.TransactionReceipt(ctx, txHash)
			if receipt != nil {
				receiptChan <- receipt
				break
			} else {
				fmt.Printf("Retry after %d second\n", retrySeconds)
				time.Sleep(retrySeconds * time.Second)
			}
		}
	}()
}
