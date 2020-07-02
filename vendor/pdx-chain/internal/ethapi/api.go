// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethapi

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/big"
	"os"
	"os/exec"
	"pdx-chain/contracts/erc20token"
	"pdx-chain/core/publicBC"
	"pdx-chain/core/state"
	"pdx-chain/ethdb"
	"pdx-chain/pdxcc"
	"pdx-chain/pdxcc/conf"
	"pdx-chain/pdxcc/protos"
	util2 "pdx-chain/pdxcc/util"
	"pdx-chain/quorum"
	"pdx-chain/utopia"
	"pdx-chain/utopia/engine/qualification"
	"pdx-chain/utopia/iaasconn"
	utils "pdx-chain/utopia/utils"
	"pdx-chain/utopia/utils/blacklist"
	"pdx-chain/utopia/utils/client"
	"pdx-chain/utopia/utils/frequency"
	"pdx-chain/utopia/utils/whitelist"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"pdx-chain/accounts"
	"pdx-chain/accounts/keystore"
	"pdx-chain/common"
	"pdx-chain/common/hexutil"
	"pdx-chain/common/math"
	"pdx-chain/consensus/ethash"
	"pdx-chain/core"
	"pdx-chain/core/rawdb"
	"pdx-chain/core/types"
	"pdx-chain/core/vm"
	"pdx-chain/crypto"
	"pdx-chain/log"
	"pdx-chain/p2p"
	"pdx-chain/params"
	pdxUtil "pdx-chain/pdxcc/util"
	"pdx-chain/rlp"
	"pdx-chain/rpc"
	utopiaTypes "pdx-chain/utopia/types"
)

const (
	defaultGasPrice                = 50 * params.Shannon
	NODE_OFF_LINE                  = 1 //没质押也不在线
	NODE_STACKED_OFF_LINE          = 2 //质押了但不在线
	NODE_STACKED_ACTIVITYNOTENOUGH = 3 //质押了但是活跃度不够
	NODE_STACKED_AUDIT             = 4 //质押了活跃度也够了只是还没生效
	NODE_INQUERUM                  = 5 //在共识委员会中
	NODE_SHOULD_QUIT_STACKING      = 6 //已退出共识委员会可以退出质押
	NODE_SHOULD_PUNISHED           = 7 //需要被惩罚
	NODE_PUNISHED                  = 8 //已被惩罚
)

// PublicEthereumAPI provides an API to access Ethereum related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicEthereumAPI struct {
	b Backend
}

var chainID2HostsMap sync.Map

// NewPublicEthereumAPI creates a new Ethereum protocol API.
func NewPublicEthereumAPI(b Backend) *PublicEthereumAPI {
	return &PublicEthereumAPI{b}
}

// GasPrice returns a suggestion for a gas price.
func (s *PublicEthereumAPI) GasPrice(ctx context.Context) (*hexutil.Big, error) {
	price, err := s.b.SuggestPrice(ctx)
	return (*hexutil.Big)(price), err
}

// ProtocolVersion returns the current Ethereum protocol version this node supports
func (s *PublicEthereumAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(s.b.ProtocolVersion())
}

// Syncing returns false in case the node is currently not syncing with the network. It can be up to date or has not
// yet received the latest block headers from its pears. In case it is synchronizing:
// - startingBlock: block number this node started to synchronise from
// - currentBlock:  block number this node is currently importing
// - highestBlock:  block number of the highest block header this node has received from peers
// - pulledStates:  number of state entries processed until now
// - knownStates:   number of known state entries that still need to be pulled
func (s *PublicEthereumAPI) Syncing() (interface{}, error) {
	progress := s.b.Downloader().Progress()

	// Return not syncing if the synchronisation already completed
	//if progress.CurrentBlock >= progress.HighestBlock {
	//	return false, nil
	//}
	//当前commit 大于最新commit 同步完成
	if progress.CurrentCommit >= progress.HighestCommit {
		return false, nil
	}

	// Otherwise gather the block sync stats
	return map[string]interface{}{
		//"startingBlock": hexutil.Uint64(progress.StartingBlock),
		//"currentBlock":  hexutil.Uint64(progress.CurrentBlock),
		"highestBlock": progress.HighestBlock,
		//"pulledStates":  hexutil.Uint64(progress.PulledStates),
		//"knownStates":   hexutil.Uint64(progress.KnownStates),
		//"heightCommit":  hexutil.Uint64(progress.HighestCommit),
		"currentCommit": progress.CurrentCommit,
	}, nil
}

// PublicTxPoolAPI offers and API for the transaction pool. It only operates on data that is non confidential.
type PublicTxPoolAPI struct {
	b Backend
}

// NewPublicTxPoolAPI creates a new tx pool service that gives information about the transaction pool.
func NewPublicTxPoolAPI(b Backend) *PublicTxPoolAPI {
	return &PublicTxPoolAPI{b}
}

// Content returns the transactions contained within the transaction pool.
func (s *PublicTxPoolAPI) Content() map[string]map[string]map[string]*RPCTransaction {
	content := map[string]map[string]map[string]*RPCTransaction{
		"pending": make(map[string]map[string]*RPCTransaction),
		"queued":  make(map[string]map[string]*RPCTransaction),
	}
	pending, queue := s.b.TxPoolContent()

	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]*RPCTransaction)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = newRPCPendingTransaction(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// Status returns the number of pending and queued transaction in the pool.
func (s *PublicTxPoolAPI) Status() map[string]hexutil.Uint {
	pending, queue := s.b.Stats()
	return map[string]hexutil.Uint{
		"pending": hexutil.Uint(pending),
		"queued":  hexutil.Uint(queue),
	}
}

// Inspect retrieves the content of the transaction pool and flattens it into an
// easily inspectable list.
func (s *PublicTxPoolAPI) Inspect() map[string]map[string]map[string]string {
	content := map[string]map[string]map[string]string{
		"pending": make(map[string]map[string]string),
		"queued":  make(map[string]map[string]string),
	}
	pending, queue := s.b.TxPoolContent()

	// Define a formatter to flatten a transaction into a string
	var format = func(tx *types.Transaction) string {
		if to := tx.To(); to != nil {
			return fmt.Sprintf("%s: %v wei + %v gas × %v wei", tx.To().Hex(), tx.Value(), tx.Gas(), tx.GasPrice())
		}
		return fmt.Sprintf("contract creation: %v wei + %v gas × %v wei", tx.Value(), tx.Gas(), tx.GasPrice())
	}
	// Flatten the pending transactions
	for account, txs := range pending {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["pending"][account.Hex()] = dump
	}
	// Flatten the queued transactions
	for account, txs := range queue {
		dump := make(map[string]string)
		for _, tx := range txs {
			dump[fmt.Sprintf("%d", tx.Nonce())] = format(tx)
		}
		content["queued"][account.Hex()] = dump
	}
	return content
}

// PublicAccountAPI provides an API to access accounts managed by this node.
// It offers only methods that can retrieve accounts.
type PublicAccountAPI struct {
	am *accounts.Manager
}

// NewPublicAccountAPI creates a new PublicAccountAPI.
func NewPublicAccountAPI(am *accounts.Manager) *PublicAccountAPI {
	return &PublicAccountAPI{am: am}
}

// Accounts returns the collection of accounts this node manages
func (s *PublicAccountAPI) Accounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// PrivateAccountAPI provides an API to access accounts managed by this node.
// It offers methods to create, (un)lock en list accounts. Some methods accept
// passwords and are therefore considered private by default.
type PrivateAccountAPI struct {
	am        *accounts.Manager
	nonceLock *AddrLocker
	b         Backend
}

// NewPrivateAccountAPI create a new PrivateAccountAPI.
func NewPrivateAccountAPI(b Backend, nonceLock *AddrLocker) *PrivateAccountAPI {
	return &PrivateAccountAPI{
		am:        b.AccountManager(),
		nonceLock: nonceLock,
		b:         b,
	}
}

// ListAccounts will return a list of addresses for accounts this node manages.
func (s *PrivateAccountAPI) ListAccounts() []common.Address {
	addresses := make([]common.Address, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		for _, account := range wallet.Accounts() {
			addresses = append(addresses, account.Address)
		}
	}
	return addresses
}

// rawWallet is a JSON representation of an accounts.Wallet interface, with its
// data contents extracted into plain fields.
type rawWallet struct {
	URL      string             `json:"url"`
	Status   string             `json:"status"`
	Failure  string             `json:"failure,omitempty"`
	Accounts []accounts.Account `json:"accounts,omitempty"`
}

// ListWallets will return a list of wallets this node manages.
func (s *PrivateAccountAPI) ListWallets() []rawWallet {
	wallets := make([]rawWallet, 0) // return [] instead of nil if empty
	for _, wallet := range s.am.Wallets() {
		status, failure := wallet.Status()

		raw := rawWallet{
			URL:      wallet.URL().String(),
			Status:   status,
			Accounts: wallet.Accounts(),
		}
		if failure != nil {
			raw.Failure = failure.Error()
		}
		wallets = append(wallets, raw)
	}
	return wallets
}

// OpenWallet initiates a hardware wallet opening procedure, establishing a USB
// connection and attempting to authenticate via the provided passphrase. Note,
// the method may return an extra challenge requiring a second open (e.g. the
// Trezor PIN matrix challenge).
func (s *PrivateAccountAPI) OpenWallet(url string, passphrase *string) error {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return err
	}
	pass := ""
	if passphrase != nil {
		pass = *passphrase
	}
	return wallet.Open(pass)
}

// DeriveAccount requests a HD wallet to derive a new account, optionally pinning
// it for later reuse.
func (s *PrivateAccountAPI) DeriveAccount(url string, path string, pin *bool) (accounts.Account, error) {
	wallet, err := s.am.Wallet(url)
	if err != nil {
		return accounts.Account{}, err
	}
	derivPath, err := accounts.ParseDerivationPath(path)
	if err != nil {
		return accounts.Account{}, err
	}
	if pin == nil {
		pin = new(bool)
	}
	return wallet.Derive(derivPath, *pin)
}

// NewAccount will create a new account and returns the address for the new account.
func (s *PrivateAccountAPI) NewAccount(password string) (common.Address, error) {
	acc, err := fetchKeystore(s.am).NewAccount(password)
	if err == nil {
		return acc.Address, nil
	}
	return common.Address{}, err
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

// ImportRawKey stores the given hex encoded ECDSA key into the key directory,
// encrypting it with the passphrase.
func (s *PrivateAccountAPI) ImportRawKey(privkey string, password string) (common.Address, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return common.Address{}, err
	}
	acc, err := fetchKeystore(s.am).ImportECDSA(key, password)
	return acc.Address, err
}

// UnlockAccount will unlock the account associated with the given address with
// the given password for duration seconds. If duration is nil it will use a
// default of 300 seconds. It returns an indication if the account was unlocked.
func (s *PrivateAccountAPI) UnlockAccount(addr common.Address, password string, duration *uint64) (bool, error) {
	const max = uint64(time.Duration(math.MaxInt64) / time.Second)
	var d time.Duration
	if duration == nil {
		d = 300 * time.Second
	} else if *duration > max {
		return false, errors.New("unlock duration too large")
	} else {
		d = time.Duration(*duration) * time.Second
	}
	err := fetchKeystore(s.am).TimedUnlock(accounts.Account{Address: addr}, password, d)
	return err == nil, err
}

// LockAccount will lock the account associated with the given address when it's unlocked.
func (s *PrivateAccountAPI) LockAccount(addr common.Address) bool {
	return fetchKeystore(s.am).Lock(addr) == nil
}

// signTransactions sets defaults and signs the given transaction
// NOTE: the caller needs to ensure that the nonceLock is held, if applicable,
// and release it after the transaction has been submitted to the tx pool
func (s *PrivateAccountAPI) signTransaction(ctx context.Context, args SendTxArgs, passwd string) (*types.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}
	wallet, err := s.am.Find(account)
	if err != nil {
		return nil, err
	}
	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainID
	}
	return wallet.SignTxWithPassphrase(account, passwd, tx, chainID)
}

// SendTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails.
//func (s *PrivateAccountAPI) SendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
//	log.Info("into SendTransaction !!!!!!!!!!!")
//	if args.Nonce == nil {
//		// Hold the addresse's mutex around signing to prevent concurrent assignment of
//		// the same nonce to multiple accounts.
//		s.nonceLock.LockAddr(args.From)
//		defer s.nonceLock.UnlockAddr(args.From)
//	}
//	signed, err := s.signTransaction(ctx, args, passwd)
//	if err != nil {
//		return common.Hash{}, err
//	}
//
//	//filter tx
//	if err := Filter(signed, ctx, s.b); err != nil {
//		return common.Hash{}, err
//	}
//
//	return submitTransaction(ctx, s.b, signed)
//}

// SignTransaction will create a transaction from the given arguments and
// tries to sign it with the key associated with args.To. If the given passwd isn't
// able to decrypt the key it fails. The transaction is returned in RLP-form, not broadcast
// to other nodes
func (s *PrivateAccountAPI) SignTransaction(ctx context.Context, args SendTxArgs, passwd string) (*SignTransactionResult, error) {
	// No need to obtain the noncelock mutex, since we won't be sending this
	// tx into the transaction pool, but right back to the user
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	signed, err := s.signTransaction(ctx, args, passwd)
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(signed)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, signed}, nil
}

// signHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calulcated as
//   keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

// Sign calculates an Ethereum ECDSA signature for:
// keccack256("\x19Ethereum Signed Message:\n" + len(message) + message))
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The key used to calculate the signature is decrypted with the given password.
//
// https://pdx-chain/wiki/Management-APIs#personal_sign
func (s *PrivateAccountAPI) Sign(ctx context.Context, data hexutil.Bytes, addr common.Address, passwd string) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Assemble sign the data with the wallet
	signature, err := wallet.SignHashWithPassphrase(account, passwd, signHash(data))
	if err != nil {
		return nil, err
	}
	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// EcRecover returns the address for the account that was used to create the signature.
// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Ethereum Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be be 27 or 28 for legacy reasons.
//
// https://pdx-chain/wiki/Management-APIs#personal_ecRecover
func (s *PrivateAccountAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	rpk, err := crypto.SigToPub(signHash(data), sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

// SignAndSendTransaction was renamed to SendTransaction. This method is deprecated
// and will be removed in the future. It primary goal is to give clients time to update.
//func (s *PrivateAccountAPI) SignAndSendTransaction(ctx context.Context, args SendTxArgs, passwd string) (common.Hash, error) {
//	return s.SendTransaction(ctx, args, passwd)
//}

// PublicBlockChainAPI provides an API to access the Ethereum blockchain.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBlockChainAPI struct {
	b Backend
}

// NewPublicBlockChainAPI creates a new Ethereum blockchain API.
func NewPublicBlockChainAPI(b Backend) *PublicBlockChainAPI {
	// modify by liangc
	p := &PublicBlockChainAPI{b}
	go p.loop()
	return p
}

func (s *PublicBlockChainAPI) GetConsensusQuorum(ctx context.Context) []string {
	currentCommit, _ := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	number := currentCommit.NumberU64()
	log.Info("取出的commit高度", "高度", number)
	address, _ := quorum.CommitHeightToConsensusQuorum.Get(number, nil)
	return address.Keys()
}

func (s *PublicBlockChainAPI) GetAmountofHypothecation(ctx context.Context) (*big.Int, error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return nil, err
	}

	quorum, ok := quorum.CommitHeightToConsensusQuorum.Get(block.NumberU64(), s.b.ChainDb())
	if !ok {
		return nil, errors.New("didnot get querom on this height")
	}

	quorumSize := len(quorum.Keys())
	nodeType := utopia.Config.Int64("nodeType")
	var num uint64
	switch nodeType {
	case 0:
		//TRUST_CHAIN
		num = 16
	case 1:
		//SERVICE_CHAIN
		num = 128
	case 2:
		//BIZ_CHAIN
		num = 128
	case 3:
		//
		num = 128
	}

	count := uint64(quorumSize) / num
	if count == 0 {
		count = 1
	}

	promitedPrice := new(big.Int).Mul(vm.HypothecationLimit, big.NewInt(int64(count)))
	return promitedPrice, nil
}

func (s *PublicBlockChainAPI) GetTotalAmountOfHypothecationWithAddress(ctx context.Context, address common.Address) (*big.Int, error) {
	newHeight := s.BlockNumber()
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.BlockNumber(newHeight))
	if err != nil {
		return big.NewInt(0), err
	}
	contractAddress := common.HexToAddress("0x04D2FBC88555e90F457F814819224460519e73A1")
	keyHash := hypothecationInfoKeyHash(address.Hex())
	hypothInfoByte := stateDB.GetPDXState(contractAddress, keyHash)
	if len(hypothInfoByte) == 0 {
		//可能已调用过退出委员会的接口
		recaContractAddr := common.HexToAddress("0x60e1e88eadf16458dacDBAe83D472CA17c960aCA")
		recaKeyHash := quiteQuorumInfoKeyHash(address.Hex())
		recaData := stateDB.GetPDXState(recaContractAddr, recaKeyHash)
		var quitInfo vm.QuitStack
		err := rlp.DecodeBytes(recaData, &quitInfo)
		if err != nil {
			return big.NewInt(0), err
		}
		return quitInfo.TotalPledgeAmount, nil
	}
	var hypothInfo vm.HypothecationInfo
	err = rlp.DecodeBytes(hypothInfoByte, &hypothInfo)
	if err != nil {
		return big.NewInt(0), err
	}
	return hypothInfo.TotalPledgeAmount, nil
}

func hypothecationInfoKeyHash(address string) common.Hash {
	return pdxUtil.EthHash([]byte(fmt.Sprintf("%s:%s:%s", conf.PDXKeyFlag, "hypothecationInfo", address)))
}

func (s *PublicBlockChainAPI) GetRunTime(ctx context.Context) string {

	return fmt.Sprintf("Running Time %.2f Hour", time.Since(utopia.StartTime).Hours())
}

func (s *PublicBlockChainAPI) GetLocalMinerRewardAddress(ctx context.Context) string {
	return utopia.Config.String("rewardAddress")
}

func (s *PublicBlockChainAPI) ChangeMinerRewardAddress(ctx context.Context, address string) (flag bool, err error) {
	var isAddress bool
	isAddress = common.IsHexAddress(address)
	if !isAddress {
		return false, errors.New("input address is err")
	}
	//return utopia.Config.UpdataGlobalFlag("rewardAddress", address)
	return true, utopia.Config.GlobalSet("rewardAddress", address)
}

func (s *PublicBlockChainAPI) GetNodeActivityInfoForAddress(ctx context.Context, address common.Address) (info uint64, err error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return 0, err
	}

	currentHeight := block.NumberU64()

	nodeDetails, ok := qualification.CommitHeight2NodeDetailSetCache.Get(currentHeight, s.b.ChainDb())
	if !ok {
		return 0, errors.New("no nodeDetails on this height")
	}

	nodeDetail := nodeDetails.Get(address.Hex())
	if nodeDetail == nil {
		return 0, errors.New("no nodeDetail record of given address")
	}

	return nodeDetail.NumAssertionsTotal, nil
}

func (s *PublicBlockChainAPI) GetNodeActiveInfo(ctx context.Context, address common.Address) (info uint64, err error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return 0, err
	}

	currentHeight := block.NumberU64()

	nodeDetails, ok := qualification.CommitHeight2NodeDetailSetCache.Get(currentHeight, s.b.ChainDb())
	if !ok {
		return 0, errors.New("no nodeDetails on this height")
	}

	nodeDetail := nodeDetails.Get(address.Hex())
	if nodeDetail == nil {
		return 0, errors.New("no nodeDetail record of given address")
	}

	return nodeDetail.UselessAssertions, nil
}

//获取质押成功的地址
func (s *PublicBlockChainAPI) GetNodeActivityInfo(ctx context.Context) ([]string, error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return nil, err
	}

	currentHeight := block.NumberU64()

	nodeDetails, ok := qualification.CommitHeight2NodeDetailSetCache.Get(currentHeight, s.b.ChainDb())
	if !ok {
		return nil, errors.New("no nodeDetails on this height")
	}

	activityCache := make([]string, 0)
	for key, node := range nodeDetails.NodeMap {
		if node.CanBeMaster == qualification.CanBeMaster {
			activityCache = append(activityCache, key)
		}
	}
	return activityCache, nil
}

func (s *PublicBlockChainAPI) GetAddressIsInQuorum(ctx context.Context, address common.Address) (flag bool, err error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return false, err
	}

	quorum, ok := quorum.CommitHeightToConsensusQuorum.Get(block.NumberU64(), s.b.ChainDb())
	if !ok {
		return false, errors.New("didnot get querom on this height")
	}

	value := quorum.Get(address.Hex())
	if value == [20]byte{} {
		return false, errors.New("this address is not in the quorum")
	}
	return true, nil
}

func (s *PublicBlockChainAPI) GetCurrentQuorumSize(ctx context.Context) (*big.Int, error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return nil, err
	}

	quorum, ok := quorum.CommitHeightToConsensusQuorum.Get(block.NumberU64(), s.b.ChainDb())
	if !ok {
		return nil, errors.New("didnot get querom on this height")
	}

	return big.NewInt(int64(len(quorum.Keys()))), nil
}

func (s *PublicBlockChainAPI) GetCurrentCommitHeight(ctx context.Context) (*big.Int, error) {
	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return nil, err
	}
	return big.NewInt(block.Number().Int64()), nil
}

func (s *PublicBlockChainAPI) GetCommitBlockByNormalBlockHash(ctx context.Context, hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, hash)
	if err != nil {
		return nil, err
	}

	blockNr := rpc.BlockNumber(block.Number().Int64())

	blockExtra := utopiaTypes.BlockExtraDecode(block)
	commitNum := blockExtra.CNumber.Uint64()
	commitBlock, err := s.b.GetCommitBlockByNum(ctx, commitNum)
	if commitBlock != nil {
		response, err := s.rpcOutputBlock(commitBlock, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// GetCommitBlockByNormalBlockNumber returns the commitBlock for the given normal block number
func (s *PublicBlockChainAPI) GetCommitBlockByNormalBlockNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if err != nil {
		return nil, err
	}

	blockExtra := utopiaTypes.BlockExtraDecode(block)
	commitNum := blockExtra.CNumber.Uint64()

	commitBlock, err := s.b.GetCommitBlockByNum(ctx, commitNum)
	if commitBlock != nil {
		response, err := s.rpcOutputBlock(commitBlock, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

// BlockNumber returns the block number of the chain head.
func (s *PublicBlockChainAPI) BlockNumber() hexutil.Uint64 {
	header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return hexutil.Uint64(header.Number.Uint64())
}

func (s *PublicBlockChainAPI) BlockCommitNumber() hexutil.Uint64 {
	header, _ := s.b.HeaderByCommitNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
	return hexutil.Uint64(header.Number.Uint64())
}

// GetBalance returns the amount of wei for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicBlockChainAPI) GetBalance(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Big, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	return (*hexutil.Big)(state.GetBalance(address)), state.Error()
}

// GetBlockByNumber returns the requested block. When blockNr is -1 the chain head is returned. When fullTx is true all
// transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	//if err := CheckQueryToken(ctx); err != nil {
	//	return nil, err
	//}
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}
		return response, err
	}
	return nil, err
}

func (s *PublicBlockChainAPI) GetNodeStatus(ctx context.Context, address common.Address) (status int, err error) {

	block, err := s.b.GetCurrentCommitBlock(ctx, common.Hash{})
	if err != nil {
		return NODE_OFF_LINE, err
	}

	nodeDetails, ok := qualification.CommitHeight2NodeDetailSetCache.Get(block.NumberU64(), s.b.ChainDb())
	if !ok {
		return NODE_OFF_LINE, errors.New("no nodetails on this height")
	}

	queru, ok := quorum.CommitHeightToConsensusQuorum.Get(block.NumberU64(), s.b.ChainDb())
	if !ok {
		return NODE_OFF_LINE, errors.New("no querom on this height")
	}

	nodeDetail := nodeDetails.Get(address.Hex())
	if nodeDetail == nil {
		return NODE_OFF_LINE, errors.New("no nodeDetail by given address")
	}

	switch nodeDetail.CanBeMaster {

	case qualification.CanBeMaster:
		if nodeDetail.NumAssertionsTotal == 0 {
			return NODE_STACKED_OFF_LINE, nil
		}

		if nodeDetail.QualifiedAt == 0 && nodeDetail.NumAssertionsTotal > 0 {
			return NODE_STACKED_ACTIVITYNOTENOUGH, nil
		}

		add := queru.Get(address.String())
		if add != quorum.EmptyAddress {
			return NODE_INQUERUM, nil
		}

		if nodeDetail.QualifiedAt != 0 && nodeDetail.NumAssertionsTotal > 0 {
			return NODE_STACKED_AUDIT, nil
		}

	case qualification.CantBeMaster:

		_, commitExtra := utopiaTypes.CommitExtraDecode(block)
		num := rpc.BlockNumber(commitExtra.NewBlockHeight.Int64())
		state, _, err := s.b.StateAndHeaderByNumber(ctx, num)
		if err != nil {
			return NODE_OFF_LINE, err
		}
		recaptionAddr := pdxUtil.EthAddress("redamption")
		data := state.GetPDXState(recaptionAddr, recaptionAddr.Hash())
		if len(data) == 0 {
			return NODE_OFF_LINE, nil
		}
		var cache1 vm.CacheMap1
		cache1.Decode(data)
		ok, _ := cache1.Get(address.Hex())
		if ok {
			return NODE_SHOULD_QUIT_STACKING, nil
		}
		return NODE_OFF_LINE, nil
	case qualification.ShouldBePunished:
		if nodeDetail.NumAssertionsTotal != 0 {
			return NODE_SHOULD_PUNISHED, nil
		} else {
			return NODE_PUNISHED, nil
		}
	}
	return NODE_OFF_LINE, errors.New("")
}

func quiteQuorumInfoKeyHash(address string) common.Hash {
	return util2.EthHash([]byte(fmt.Sprintf("%s:%s:%s", conf.PDXKeyFlag, "recaption", address)))
}

func (s *PublicBlockChainAPI) GetCommitBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	block, err := s.b.CommitBlockByNumber(ctx, blockNr)
	if block != nil {
		response, err := s.rpcOutputBlock(block, true, fullTx)
		if err == nil && blockNr == rpc.PendingBlockNumber {
			// Pending blocks need to nil out a few fields
			for _, field := range []string{"hash", "nonce", "miner"} {
				response[field] = nil
			}
		}

		println(response)
		return response, err
	}
	return nil, err
}

// GetBlockByHash returns the requested block. When fullTx is true all transactions in the block are returned in full
// detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetBlockByHash(ctx context.Context, blockHash common.Hash, fullTx bool) (map[string]interface{}, error) {
	//if err := CheckQueryToken(ctx); err != nil {
	//	return nil, err
	//}
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		return s.rpcOutputBlock(block, true, fullTx)
	}
	return nil, err
}

// GetUncleByBlockNumberAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.BlockByNumber(ctx, blockNr)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", blockNr, "hash", block.Hash(), "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleByBlockHashAndIndex returns the uncle block for the given block hash and index. When fullTx is true
// all transactions in the block are returned in full detail, otherwise only the transaction hash is returned.
func (s *PublicBlockChainAPI) GetUncleByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) (map[string]interface{}, error) {
	block, err := s.b.GetBlock(ctx, blockHash)
	if block != nil {
		uncles := block.Uncles()
		if index >= hexutil.Uint(len(uncles)) {
			log.Debug("Requested uncle not found", "number", block.Number(), "hash", blockHash, "index", index)
			return nil, nil
		}
		block = types.NewBlockWithHeader(uncles[index])
		return s.rpcOutputBlock(block, false, false)
	}
	return nil, err
}

// GetUncleCountByBlockNumber returns number of uncles in the block for the given block number
func (s *PublicBlockChainAPI) GetUncleCountByBlockNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetUncleCountByBlockHash returns number of uncles in the block for the given block hash
func (s *PublicBlockChainAPI) GetUncleCountByBlockHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Uncles()))
		return &n
	}
	return nil
}

// GetCode returns the code stored at the given address in the state for the given block number.
func (s *PublicBlockChainAPI) GetCode(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	code := state.GetCode(address)
	return code, state.Error()
}

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (s *PublicBlockChainAPI) GetStorageAt(ctx context.Context, address common.Address, key string, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, common.HexToHash(key))
	return res[:], state.Error()
}

//chainID service chain ID
func (s *PublicBlockChainAPI) GetMaxCommitBlockHash(ctx context.Context, chainID string) (commitBlockHah string) {
	newHeight := s.BlockNumber()
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.BlockNumber(newHeight))
	if stateDB == nil || err != nil {
		return ""
	}

	key := fmt.Sprintf("%s-trustTransactionHeight", chainID)
	keyHash := pdxUtil.EthHash([]byte(key + conf.PDXKeyFlag))
	st := stateDB.GetPDXState(pdxUtil.EthAddress(utils.BaapTrustTree), keyHash)
	if len(st) > 0 {
		maxNum := big.NewInt(0).SetBytes(st)
		log.Info("maxNum", "is", maxNum)

		commitBlockKey := fmt.Sprintf("%s:%s", chainID, maxNum.String())
		commitBlockKeyHash := pdxUtil.EthHash([]byte(commitBlockKey + conf.PDXKeyFlag))
		commitBlockSt := stateDB.GetPDXState(pdxUtil.EthAddress(utils.BaapTrustTree), commitBlockKeyHash)
		if len(commitBlockSt) > 0 {
			return hex.EncodeToString(commitBlockSt)
		}
	}
	return ""
}

// CallArgs represents the arguments for a call.
type CallArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      hexutil.Uint64  `json:"gas"`
	GasPrice hexutil.Big     `json:"gasPrice"`
	Value    hexutil.Big     `json:"value"`
	Data     hexutil.Bytes   `json:"data"`
	estimate bool
}

func (s *PublicBlockChainAPI) doCall(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) ([]byte, uint64, bool, error) {
	_, a, b, c, d := s.doCallRtnState(ctx, common.HexToHash("0x"), args, blockNr, vmCfg, timeout)
	return a, b, c, d
}
func (s *PublicBlockChainAPI) doCallRtnState(ctx context.Context, txHash common.Hash, args CallArgs, blockNr rpc.BlockNumber, vmCfg vm.Config, timeout time.Duration) (*state.StateDB, []byte, uint64, bool, error) {
	//defer func(start time.Time) { fmt.Println("PRECREATE : doCallRtnState ", "timeused", time.Since(start)) }(time.Now())
	state, header, err := s.b.StateAndHeaderByNumber(ctx, blockNr)

	if state == nil || err != nil {
		return nil, nil, 0, false, err
	}
	if txHash != common.HexToHash("0x") {
		state.Prepare(txHash, header.Hash(), 0)
	}

	// Set sender address or use a default if none specified
	addr := args.From
	if addr == (common.Address{}) {
		if wallets := s.b.AccountManager().Wallets(); len(wallets) > 0 {
			if accounts := wallets[0].Accounts(); len(accounts) > 0 {
				addr = accounts[0].Address
			}
		}
	}
	// Set default gas & gas price if none were set
	gas, gasPrice := uint64(args.Gas), args.GasPrice.ToInt()
	if gas == 0 {
		gas = math.MaxUint64 / 2
	}
	if gasPrice.Sign() == 0 {
		gasPrice = new(big.Int).SetUint64(defaultGasPrice)
	}

	// Create new call message
	msg := types.NewMessage(addr, args.To, 0, args.Value.ToInt(), gas, gasPrice, args.Data, false)

	// Setup context so it may be cancelled the call has completed
	// or, in case of unmetered gas, setup a context with a timeout.
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	// Make sure the context is cancelled when the call has completed
	// this makes sure resources are cleaned up.
	defer cancel()

	// Get a new instance of the EVM.
	evm, vmError, err := s.b.GetEVM(ctx, msg, state, header, vmCfg)
	if err != nil {
		return nil, nil, 0, false, err
	}
	// Wait for the context to be done and cancel the evm. Even if the
	// EVM has finished, cancelling may be done (repeatedly)
	go func() {
		<-ctx.Done()
		evm.Cancel()
	}()

	extra := make(map[string][]byte)
	if args.estimate {
		extra[conf.TxEstimateGas] = []byte(conf.TxEstimateGas)
	}

	// Setup the gas pool (also for unmetered requests)
	// and apply the message.
	gp := new(core.GasPool).AddGas(math.MaxUint64)
	res, gas, failed, err := core.ApplyMessage(evm, msg, gp, extra)
	if err := vmError(); err != nil {
		return nil, nil, 0, false, err
	}
	return state, res, gas, failed, err
}

// Call executes the given transaction on the state for the given block number.
// It doesn't make and changes in the state/blockchain and is useful to execute and retrieve values.
func (s *PublicBlockChainAPI) Call(ctx context.Context, args CallArgs, blockNr rpc.BlockNumber) (hexutil.Bytes, error) {
	result, _, _, err := s.doCall(ctx, args, blockNr, vm.Config{}, 5*time.Second)
	return (hexutil.Bytes)(result), err
}

// EstimateGas returns an estimate of the amount of gas needed to execute the
// given transaction against the current pending block.
func (s *PublicBlockChainAPI) EstimateGas(ctx context.Context, args CallArgs) (hexutil.Uint64, error) {
	// Binary search the gas requirement, as it may be higher than the amount used
	var (
		lo  uint64 = params.TxGas - 1
		hi  uint64
		cap uint64
	)
	if uint64(args.Gas) >= params.TxGas {
		hi = uint64(args.Gas)
	} else {
		// Retrieve the current pending block to act as the gas ceiling
		block, err := s.b.BlockByNumber(ctx, rpc.PendingBlockNumber)
		if err != nil {
			return 0, err
		}
		hi = block.GasLimit()
	}
	cap = hi

	if args.To == nil &&
		utils.EwasmToolImpl.IsWASM(args.Data) &&
		!utils.EwasmToolImpl.IsFinalcode(args.Data) {
		code := args.Data
		final, err := utils.EwasmToolImpl.Sentinel(code)
		if err != nil {
			log.Error("tx.data format error 1", "err", err)
			return hexutil.Uint64(0), err
		}
		// 处理 meta
		code, _, err = utopia.ParseData(code)
		if err != nil {
			panic(err)
		}
		codekey := crypto.Keccak256(code)
		finalkey := crypto.Keccak256(final)
		utils.EwasmToolImpl.PutCode(finalkey, final)
		utils.EwasmToolImpl.PutCode(codekey, finalkey)
	}

	// Create a helper to check if a gas allowance results in an executable transaction
	executable := func(gas uint64) bool {
		args.Gas = hexutil.Uint64(gas)
		args.estimate = true

		_, _, failed, err := s.doCall(ctx, args, rpc.PendingBlockNumber, vm.Config{}, 0)
		if err != nil || failed {
			return false
		}
		return true
	}
	// Execute the binary search and hone in on an executable gas limit
	for lo+1 < hi {
		mid := (hi + lo) / 2
		if !executable(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	// Reject the transaction as invalid if it still fails at the highest allowance
	if hi == cap {
		log.Error("estimate gaslimit", "hi", hi)
		if !executable(hi) {
			return 0, fmt.Errorf("gas required exceeds allowance or always failing transaction")
		}
	}
	return hexutil.Uint64(hi), nil
}

// ExecutionResult groups all structured logs emitted by the EVM
// while replaying a transaction in debug mode as well as transaction
// execution status, the amount of gas used and the return value
type ExecutionResult struct {
	Gas         uint64         `json:"gas"`
	Failed      bool           `json:"failed"`
	ReturnValue string         `json:"returnValue"`
	StructLogs  []StructLogRes `json:"structLogs"`
}

// StructLogRes stores a structured log emitted by the EVM while replaying a
// transaction in debug mode
type StructLogRes struct {
	Pc      uint64             `json:"pc"`
	Op      string             `json:"op"`
	Gas     uint64             `json:"gas"`
	GasCost uint64             `json:"gasCost"`
	Depth   int                `json:"depth"`
	Error   error              `json:"error,omitempty"`
	Stack   *[]string          `json:"stack,omitempty"`
	Memory  *[]string          `json:"memory,omitempty"`
	Storage *map[string]string `json:"storage,omitempty"`
}

// formatLogs formats EVM returned structured logs for json output
func FormatLogs(logs []vm.StructLog) []StructLogRes {
	formatted := make([]StructLogRes, len(logs))
	for index, trace := range logs {
		formatted[index] = StructLogRes{
			Pc:      trace.Pc,
			Op:      trace.Op.String(),
			Gas:     trace.Gas,
			GasCost: trace.GasCost,
			Depth:   trace.Depth,
			Error:   trace.Err,
		}
		if trace.Stack != nil {
			stack := make([]string, len(trace.Stack))
			for i, stackValue := range trace.Stack {
				stack[i] = fmt.Sprintf("%x", math.PaddedBigBytes(stackValue, 32))
			}
			formatted[index].Stack = &stack
		}
		if trace.Memory != nil {
			memory := make([]string, 0, (len(trace.Memory)+31)/32)
			for i := 0; i+32 <= len(trace.Memory); i += 32 {
				memory = append(memory, fmt.Sprintf("%x", trace.Memory[i:i+32]))
			}
			formatted[index].Memory = &memory
		}
		if trace.Storage != nil {
			storage := make(map[string]string)
			for i, storageValue := range trace.Storage {
				storage[fmt.Sprintf("%x", i)] = fmt.Sprintf("%x", storageValue)
			}
			formatted[index].Storage = &storage
		}
	}
	return formatted
}

// RPCMarshalBlock converts the given block to the RPC output which depends on fullTx. If inclTx is true transactions are
// returned. When fullTx is true the returned block contains full transaction details, otherwise it will only contain
// transaction hashes.
func RPCMarshalBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	head := b.Header() // copies the header once
	fields := map[string]interface{}{
		"number":           (*hexutil.Big)(head.Number),
		"hash":             b.Hash(),
		"parentHash":       head.ParentHash,
		"nonce":            head.Nonce,
		"mixHash":          head.MixDigest,
		"sha3Uncles":       head.UncleHash,
		"logsBloom":        head.Bloom,
		"stateRoot":        head.Root,
		"miner":            head.Coinbase,
		"rewardAddr":       head.RewardAddress,
		"difficulty":       (*hexutil.Big)(head.Difficulty),
		"extraData":        hexutil.Bytes(head.Extra),
		"size":             hexutil.Uint64(b.Size()),
		"gasLimit":         hexutil.Uint64(head.GasLimit),
		"gasUsed":          hexutil.Uint64(head.GasUsed),
		"timestamp":        (*hexutil.Big)(head.Time),
		"transactionsRoot": head.TxHash,
		"receiptsRoot":     head.ReceiptHash,
	}

	if inclTx {
		formatTx := func(tx *types.Transaction) (interface{}, error) {
			return tx.Hash(), nil
		}
		if fullTx {
			formatTx = func(tx *types.Transaction) (interface{}, error) {
				return newRPCTransactionFromBlockHash(b, tx.Hash()), nil
			}
		}
		txs := b.Transactions()
		transactions := make([]interface{}, len(txs))
		var err error
		for i, tx := range txs {
			if transactions[i], err = formatTx(tx); err != nil {
				return nil, err
			}
		}
		fields["transactions"] = transactions
	}

	uncles := b.Uncles()
	uncleHashes := make([]common.Hash, len(uncles))
	for i, uncle := range uncles {
		uncleHashes[i] = uncle.Hash()
	}
	fields["uncles"] = uncleHashes

	return fields, nil
}

// rpcOutputBlock uses the generalized output filler, then adds the total difficulty field, which requires
// a `PublicBlockchainAPI`.
func (s *PublicBlockChainAPI) rpcOutputBlock(b *types.Block, inclTx bool, fullTx bool) (map[string]interface{}, error) {
	fields, err := RPCMarshalBlock(b, inclTx, fullTx)
	if err != nil {
		return nil, err
	}
	fields["totalDifficulty"] = (*hexutil.Big)(s.b.GetTd(b.Hash()))
	return fields, err
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
	Status        int             `json:"status"` //0 pending 1 success 2 commitSuccess -1 withdrawFail
	CommitNumber  *hexutil.Big    `json:"commitNumber"`
	CurrentCommit *hexutil.Big    `json:"currentCommit"`
	TxMsg         string          `json:"txMsg"` //交易报文
	DstChainID    string          `json:"dst_chain_id"`
}

// newRPCTransaction returns a transaction that will serialize to the RPC
// representation, with the given location metadata set (if available).
func newRPCTransaction(tx *types.Transaction, blockHash common.Hash, blockNumber uint64, index uint64) *RPCTransaction {
	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	if tx.IsSM2() {
		signer = types.NewSm2Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)
	v, r, s := tx.RawSignatureValues()

	result := &RPCTransaction{
		From:     from,
		Gas:      hexutil.Uint64(tx.Gas()),
		GasPrice: (*hexutil.Big)(tx.GasPrice()),
		Hash:     tx.Hash(),
		Input:    hexutil.Bytes(tx.Data()),
		Nonce:    hexutil.Uint64(tx.Nonce()),
		To:       tx.To(),
		Value:    (*hexutil.Big)(tx.Value()),
		V:        (*hexutil.Big)(v),
		R:        (*hexutil.Big)(r),
		S:        (*hexutil.Big)(s),
	}
	if blockHash != (common.Hash{}) {
		result.BlockHash = blockHash
		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
		result.TransactionIndex = hexutil.Uint(index)
	}
	return result
}

func GenTxMsg(tx *types.Transaction) (txMsg string) {
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Error("rlp encode tx error", "err", err)
	} else {
		txMsg = common.ToHex(data)
	}
	return
}

// newRPCPendingTransaction returns a pending transaction that will serialize to the RPC representation
func newRPCPendingTransaction(tx *types.Transaction) *RPCTransaction {
	return newRPCTransaction(tx, common.Hash{}, 0, 0)
}

// newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockIndex(b *types.Block, index uint64) *RPCTransaction {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), index)
}

// newRPCRawTransactionFromBlockIndex returns the bytes of a transaction given a block and a transaction index.
func newRPCRawTransactionFromBlockIndex(b *types.Block, index uint64) hexutil.Bytes {
	txs := b.Transactions()
	if index >= uint64(len(txs)) {
		return nil
	}
	blob, _ := rlp.EncodeToBytes(txs[index])
	return blob
}

// newRPCTransactionFromBlockHash returns a transaction that will serialize to the RPC representation.
func newRPCTransactionFromBlockHash(b *types.Block, hash common.Hash) *RPCTransaction {
	for idx, tx := range b.Transactions() {
		if tx.Hash() == hash {
			return newRPCTransactionFromBlockIndex(b, uint64(idx))
		}
	}
	return nil
}

// PublicTransactionPoolAPI exposes methods for the RPC interface
type PublicTransactionPoolAPI struct {
	b         Backend
	nonceLock *AddrLocker
}

// NewPublicTransactionPoolAPI creates a new RPC service with methods specific for the transaction pool.
func NewPublicTransactionPoolAPI(b Backend, nonceLock *AddrLocker) *PublicTransactionPoolAPI {
	return &PublicTransactionPoolAPI{b, nonceLock}
}

// GetBlockTransactionCountByNumber returns the number of transactions in the block with the given block number.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByNumber(ctx context.Context, blockNr rpc.BlockNumber) *hexutil.Uint {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetBlockTransactionCountByHash returns the number of transactions in the block with the given hash.
func (s *PublicTransactionPoolAPI) GetBlockTransactionCountByHash(ctx context.Context, blockHash common.Hash) *hexutil.Uint {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		n := hexutil.Uint(len(block.Transactions()))
		return &n
	}
	return nil
}

// GetTransactionByBlockNumberAndIndex returns the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) *RPCTransaction {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockNumberAndIndex returns the bytes of the transaction for the given block number and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockNumberAndIndex(ctx context.Context, blockNr rpc.BlockNumber, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.BlockByNumber(ctx, blockNr); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetRawTransactionByBlockHashAndIndex returns the bytes of the transaction for the given block hash and index.
func (s *PublicTransactionPoolAPI) GetRawTransactionByBlockHashAndIndex(ctx context.Context, blockHash common.Hash, index hexutil.Uint) hexutil.Bytes {
	if block, _ := s.b.GetBlock(ctx, blockHash); block != nil {
		return newRPCRawTransactionFromBlockIndex(block, uint64(index))
	}
	return nil
}

// GetTransactionCount returns the number of transactions the given address has sent for the given block number
func (s *PublicTransactionPoolAPI) GetTransactionCount(ctx context.Context, address common.Address, blockNr rpc.BlockNumber) (*hexutil.Uint64, error) {
	state, _, err := s.b.StateAndHeaderByNumber(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	nonce := state.GetNonce(address)
	// 如果此处不校准，导致第一次获取PublicAccountVote合约的nonce是0，第二次则成了1，
	// 从而第一个人投了第0轮的票，第二个人就投了第1轮。
	if address == utils.PublicAccountVote && nonce == 0 {
		nonce = 1
	}
	return (*hexutil.Uint64)(&nonce), state.Error()
}

// GetTransactionByHash returns the transaction for the given hash
func (s *PublicTransactionPoolAPI) GetTransactionByHash(ctx context.Context, hash common.Hash) *RPCTransaction {
	//if err := CheckQueryToken(ctx); err != nil {
	//	return nil
	//}
	// Try to return an already finalized transaction
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		return newRPCTransaction(tx, blockHash, blockNumber, index)
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		return newRPCPendingTransaction(tx)
	}
	// Transaction unknown, return as such
	return nil
}

//calculate solidity contract addr
func (s *PublicTransactionPoolAPI) GetContractAddrByHash(ctx context.Context, txHash common.Hash) common.Address {
	// Try to return an already finalized transaction
	tx, _, _, _ := rawdb.ReadTransaction(s.b.ChainDb(), txHash)
	if tx == nil {
		tx = s.b.GetPoolTransaction(txHash)
	}

	if tx != nil && tx.To() == nil {
		//pub := tx.PublicKey()
		//key, err := crypto.DecompressPubkey(pub)
		//if err != nil {
		//    log.Error("decompress public key error", "err", err)
		//    return common.Address{}
		//}
		//from := crypto.PubkeyToAddress(*key)
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(s.b.ChainConfig().ChainID)
		}
		if tx.IsSM2() {
			signer = types.NewSm2Signer(s.b.ChainConfig().ChainID)
		}
		from, err := types.Sender(signer, tx)
		if err != nil {
			log.Error("invalide sender")
			return common.Address{}
		}

		addr := crypto.CreateAddress(from, tx.Nonce())
		log.Info("contract addr", "addr", addr)
		return addr
	}
	// No finalized transaction, try to retrieve it from the pool

	// Transaction unknown, return as such
	return common.Address{}
}

func (s *PublicTransactionPoolAPI) CheckTxCompletedByHash(ctx context.Context, from common.Address, txHash common.Hash) bool {
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		log.Error("get stateDB error", "err", err)
		return true
	}

	keyHash := utils.DepositKeyHash(from, txHash, conf.PDXKeyFlag)
	res := stateDB.GetPDXState(utils.DepositContractAddr(), keyHash)
	if len(res) != 0 {
		log.Info("has completed tx", "txhash", txHash.Hex())
		return true
	}
	log.Info("not completed tx", "txhash", txHash.Hex())
	return false
}

func (s *PublicTransactionPoolAPI) GetUncompletedTransactionByAddress(ctx context.Context, from common.Address) []common.Hash {
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		log.Error("get stateDB error", "err", err)
		return nil
	}

	var myTxids []common.Hash

	contractAddr := pdxUtil.EthAddress(utils.WithdrawContract)
	keyHash := utils.WithdrawTxKeyHash(from)
	value := stateDB.GetPDXState(contractAddr, keyHash)
	if len(value) == 0 {
		return myTxids
	}

	var txs []vm.WithdrawTx
	err = json.Unmarshal(value, &txs)
	if err != nil {
		log.Error("unmarshal withdrawTx error", "err", err)
		return myTxids
	}

	//blockFreeCloud url
	proxyServer := utopia.Config.String("blockFreeCloud")
	iaasServer := utopia.Config.String("iaas")

	var privateKey *ecdsa.PrivateKey
	var srcChainid string
	if proxyServer != "" && iaasServer != "" {
		ks := s.b.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
		etherBase := utopia.Config.String("etherbase")
		unlocked := ks.GetUnlocked(common.HexToAddress(etherBase))
		privateKey = unlocked.PrivateKey
		//current chain id
		srcChainid = conf.ChainId.String()
	}

	for _, v := range txs {
		log.Info("################### range withdraw info", "txid", v.Txid)
		var hosts []string
		//如果设置了代理blockFreeCloud,就去走代理
		token := utopia.GetToken(privateKey, srcChainid, v.Payload.DstChainID, utopia.XChainTokenType)
		if token != "" {
			hosts = []string{fmt.Sprintf(utopia.ProxyRPC, v.Payload.DstChainID)}
		} else {
			//根据destChainID放到全局map变量中，如果获取失败，就尝试5次。
			//尝试10个host，如果错误的host数量大于5个，就再次根据destChainID获取一遍，更新全局map
			//如果该笔交易在链2上已经成功了，记录到本地数据库，下次再查询时，先从本地数据库查询，如果已经成功的交易，就直接返回，不用再次请求链2
			if value, ok := chainID2HostsMap.Load(v.Payload.DstChainID); ok {
				hosts, ok = value.([]string)
				if !ok {
					log.Error("value assertion fail")
					return myTxids
				}

				log.Trace("##################### load success", "hosts", hosts)
			} else {
				hosts, err = iaasconn.GetNodeFromIaas(v.Payload.DstChainID)
				if err != nil || len(hosts) == 0 {
					log.Error("get node from iaas", "err", err, "hosts", hosts)
					return myTxids
				}
				log.Trace("###################### get nodes from iaas", "hosts", hosts)
				chainID2HostsMap.Store(v.Payload.DstChainID, hosts)
			}
		}

		//whether has succeed
		db := *ethdb.ChainDb
		completedTxKey := pdxUtil.EthHash([]byte(fmt.Sprintf("utopia:completed:withdraw:tx:%s", v.Txid)))
		succeed, err := db.Has(completedTxKey.Bytes())
		if err != nil {
			log.Error("db has completed tx", "err", err)
			return myTxids
		}
		if succeed {
			log.Trace("################## tx has completed", "txid", v.Txid)
			continue
		}
		var badHostN int
		for _, host := range hosts {
			cli, err := client.Connect(host, proxyServer, token)
			if err != nil {
				badHostN++
				log.Error("client connect error", "err", err, "host", host)
				continue
			}

			txHash := common.HexToHash(v.Txid)
			completed, err := cli.CheckTxCompletedByHash(from, txHash)
			if err != nil {
				badHostN++
				log.Error("cli CheckTxCompletedByHash", "err", err)
				continue
			}
			if !completed {
				myTxids = append(myTxids, txHash)
			} else {
				log.Trace("############### completed tx, so put db", "txid", txHash.String())
				err = db.Put(completedTxKey.Bytes(), []byte{})
				if err != nil {
					log.Error("db put completed tx", "err", err)
					return myTxids
				}
			}
			break
		}

		if badHostN > 3 && token == "" { //非代理模式走这里
			log.Trace("############### bad host num more than 3")
			hosts, err = iaasconn.GetNodeFromIaas(v.Payload.DstChainID)
			if err != nil {
				log.Error("get node from iaas", "err", err)
				return myTxids
			}

			chainID2HostsMap.Store(v.Payload.DstChainID, hosts)
		}
	}

	return myTxids
}

//chainID service chain
func (s *PublicTransactionPoolAPI) GetMaxNumFromTrustChain(ctx context.Context, chainID string) (maxNum *big.Int) {
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		log.Error("get stateDB error", "err", err)
		return nil
	}

	key := fmt.Sprintf("%s-trustTransactionHeight", chainID)
	keyHash := pdxUtil.EthHash([]byte(key + conf.PDXKeyFlag))
	st := stateDB.GetPDXState(pdxUtil.EthAddress(utils.BaapTrustTree), keyHash)
	if len(st) > 0 {
		maxNum = big.NewInt(0).SetBytes(st)
		log.Info("maxNum", "is", maxNum)
		return maxNum
	}

	return nil

}

func (s *PublicTransactionPoolAPI) GetWithDrawTransactionByHash(ctx context.Context, hash common.Hash) (*WithdrawRPCTransaction, error) {
	// Try to return an already finalized transaction
	withdrawTx := new(WithdrawRPCTransaction)
	if tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash); tx != nil {
		withdrawTx.TxMsg = GenTxMsg(tx)
		withdrawTx.RPCTx = newRPCTransaction(tx, blockHash, blockNumber, index)
		withdraw := &vm.Withdraw{}
		err := json.Unmarshal(tx.Data(), withdraw)
		if err != nil {
			log.Error("unmarshal withdraw error", "err", err)
			return nil, fmt.Errorf("unmarshal withdraw error:%v", err)
		}
		withdrawTx.DstChainID = withdraw.DstChainID
		//判断是否达到了10*cfd个块的确认
		currentCommit, err := s.b.GetCurrentCommitBlock(ctx, hash)
		if err != nil {
			log.Error("get current Commitblock error", "err", err)
			return nil, fmt.Errorf("get current Commitblock error:%v", err)
		}
		withdrawTx.CurrentCommit = (*hexutil.Big)(currentCommit.Number())
		block, err := s.b.BlockByNumber(ctx, rpc.BlockNumber(blockNumber))
		if err != nil {
			log.Error("get block by number error", "err", err)
			return nil, fmt.Errorf("get block by number error:%v", err)
		}

		var extra utopiaTypes.BlockExtra
		extra.Decode(block.Extra())
		log.Info("extra decode", "extra", extra)
		withdrawTx.CommitNumber = (*hexutil.Big)(extra.CNumber)

		res := new(big.Int)
		res.Sub(currentCommit.Number(), extra.CNumber)

		//check withdraw success
		stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
		if err != nil {
			log.Error("get stateDB error", "err", err)
			return nil, fmt.Errorf("get stateDB error:%v", err)
		}

		//pub := tx.PublicKey()
		//key, _ := crypto.DecompressPubkey(pub)
		//from := crypto.PubkeyToAddress(*key)
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(s.b.ChainConfig().ChainID)
		}
		if tx.IsSM2() {
			signer = types.NewSm2Signer(s.b.ChainConfig().ChainID)
		}
		from, err := types.Sender(signer, tx)
		if err != nil {
			log.Error("invalid sender")
			return nil, fmt.Errorf("invalid sender")
		}

		withdrawTx.Status = -1

		contractAddr := pdxUtil.EthAddress(utils.WithdrawContract)
		keyHash := utils.WithdrawTxKeyHash(from)
		value := stateDB.GetPDXState(contractAddr, keyHash)
		if len(value) == 0 {
			return nil, fmt.Errorf("tx1 fail")
		}

		var txs []vm.WithdrawTx
		err = json.Unmarshal(value, &txs)
		if err != nil {
			log.Error("unmarshal withdrawTx error", "err", err)
			return nil, fmt.Errorf("unmarshal withdrawTx error:%v", err)
		}

		for _, v := range txs {
			log.Info("range withdraw info", "txid", v.Txid)
			if v.Txid == hash.String() {
				withdrawTx.Status = 1
				if res.Cmp(big.NewInt(utopia.XChainTransferCommitNum)) >= 0 {
					withdrawTx.Status = 2
				}
			}
		}

		return withdrawTx, nil
	}
	// No finalized transaction, try to retrieve it from the pool
	if tx := s.b.GetPoolTransaction(hash); tx != nil {
		withdrawTx.TxMsg = GenTxMsg(tx)

		withdrawTx.Status = 0
		withdrawTx.RPCTx = newRPCPendingTransaction(tx)
		return withdrawTx, nil
	}
	// Transaction unknown, return as such
	return nil, fmt.Errorf("Transaction unknown")
}

// GetRawTransactionByHash returns the bytes of the transaction for the given hash.
func (s *PublicTransactionPoolAPI) GetRawTransactionByHash(ctx context.Context, hash common.Hash) (hexutil.Bytes, error) {
	var tx *types.Transaction

	// Retrieve a finalized transaction, or a pooled otherwise
	if tx, _, _, _ = rawdb.ReadTransaction(s.b.ChainDb(), hash); tx == nil {
		if tx = s.b.GetPoolTransaction(hash); tx == nil {
			// Transaction not found anywhere, abort
			return nil, nil
		}
	}
	// Serialize to RLP and return
	return rlp.EncodeToBytes(tx)
}

// add by liangc
func (s *PublicTransactionPoolAPI) PeepEwasm(txHash string) (map[string]map[string]string, error) {
	if tx, _, num, _ := rawdb.ReadTransaction(s.b.ChainDb(), common.HexToHash(txHash)); tx != nil && utils.EwasmToolImpl.IsWASM(tx.Data()) {
		code := tx.Data()
		//codekey := crypto.Keccak256(code)
		codekey := utils.EwasmToolImpl.GenCodekey(code)
		final, ok := utils.EwasmToolImpl.GetCode(codekey)
		if ok && len(final) == 32 {
			final, _ = utils.EwasmToolImpl.GetCode(final)
		}
		finalcode, err := utils.EwasmToolImpl.JoinTxdata(code, final)
		if err != nil {
			return nil, err
		}
		// TODO get status
		m := map[string]map[string]string{
			"codekey":   {"size": strconv.Itoa(len(codekey)), "hex": hex.EncodeToString(codekey)},
			"code":      {"size": strconv.Itoa(len(code)), "hex": hex.EncodeToString(code)},
			"final":     {"size": strconv.Itoa(len(final)), "hex": hex.EncodeToString(final)},
			"finalcode": {"size": strconv.Itoa(len(finalcode)), "hex": hex.EncodeToString(finalcode)},
			"status":    {"block": strconv.Itoa(int(num)), "target": strconv.Itoa(int(num) + 100), "status": "TODO"},
		}
		return m, nil
	}
	return nil, nil
}

/*
	_initialSupply, _ = new(big.Int).SetString("10000000000000000000", 10)
	_name             = "HelloWorld"
	_symbol           = "HWT"
	_decimals         = big.NewInt(18)
*/
func (s *PublicTransactionPoolAPI) CompileToken(initialSupply string, name string, symbol string, decimals *big.Int) (hexutil.Bytes, error) {
	supply, ok := new(big.Int).SetString(initialSupply, 10)
	if !ok {
		return nil, errors.New("initial supply error")
	}
	fmt.Println("-------------------------------------------------------------")
	fmt.Println("initialSupply", initialSupply)
	fmt.Println("name", name)
	fmt.Println("symbol", symbol)
	fmt.Println("decimals", decimals)
	fmt.Println("-------------------------------------------------------------")
	return erc20token.Template.Compile(supply, name, symbol, decimals)
}

func (s *PublicTransactionPoolAPI) CurrentCommitHash() (string, error) {
	cb, err := s.b.GetCurrentCommitBlock(context.Background(), common.Hash{})
	if err != nil {
		return "", err
	}
	fmt.Println("CurrentCommitBlockHash", cb.Hash().Hex())
	return cb.Hash().Hex(), nil
}

//check withdraw tx status
func (s *PublicTransactionPoolAPI) GetWithDrawTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]

	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	if tx.IsSM2() {
		signer = types.NewSm2Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(blockNumber),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(index),
		"from":              from,
		"to":                tx.To(),
		"value":             (*hexutil.Big)(tx.Value()),
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
		"status":            types.ReceiptStatusFailed,
	}

	// Assign receipt status or post state.
	log.Trace("receipt status", "status", receipt.Status)
	if receipt.Status == types.ReceiptStatusSuccessful {
		//判断是否达到了10*cfd个块的确认
		header, _ := s.b.HeaderByNumber(context.Background(), rpc.LatestBlockNumber) // latest header should always be available
		if header.Number.Uint64()-blockNumber >= utopia.XChainTransferCommitNum {
			fields["status"] = types.ReceiptStatusSuccessful
		}
	}

	if receipt.Logs == nil {
		fields["logs"] = [][]*types.Log{}
	}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = receipt.ContractAddress
	}
	return fields, nil
}

func CheckQueryToken(ctx context.Context) error {
	if utopia.Consortium {
		//verify jwt token
		token, ok := ctx.Value("jwt").(string)
		if !ok {
			return errors.New("jwt token illegal")
		}
		if !checkJWT(token, "a") {
			return errors.New("jwt token check fail")
		}
	}

	return nil
}

// GetTransactionReceipt returns the transaction receipt for the given transaction hash.
func (s *PublicTransactionPoolAPI) GetTransactionReceipt(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	//if err := CheckQueryToken(ctx); err != nil {
	//	return nil, err
	//}

	tx, blockHash, blockNumber, index := rawdb.ReadTransaction(s.b.ChainDb(), hash)
	if tx == nil {
		return nil, nil
	}
	receipts, err := s.b.GetReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	if len(receipts) <= int(index) {
		return nil, nil
	}
	receipt := receipts[index]

	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	if tx.IsSM2() {
		signer = types.NewSm2Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)

	fields := map[string]interface{}{
		"blockHash":         blockHash,
		"blockNumber":       hexutil.Uint64(blockNumber),
		"transactionHash":   hash,
		"transactionIndex":  hexutil.Uint64(index),
		"from":              from,
		"to":                tx.To(),
		"gasUsed":           hexutil.Uint64(receipt.GasUsed),
		"cumulativeGasUsed": hexutil.Uint64(receipt.CumulativeGasUsed),
		"contractAddress":   nil,
		"logs":              receipt.Logs,
		"logsBloom":         receipt.Bloom,
	}

	// Assign receipt status or post state.
	if len(receipt.PostState) > 0 {
		fields["root"] = hexutil.Bytes(receipt.PostState)
	} else {
		fields["status"] = hexutil.Uint(receipt.Status)
	}
	if receipt.Logs == nil {
		fields["logs"] = [][]*types.Log{}
	}
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if receipt.ContractAddress != (common.Address{}) {
		fields["contractAddress"] = receipt.ContractAddress
	}
	return fields, nil
}

// sign is a helper function that signs a transaction with the private key of the given address.
func (s *PublicTransactionPoolAPI) sign(addr common.Address, tx *types.Transaction) (*types.Transaction, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Request the wallet to sign the transaction
	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainID
	}
	return wallet.SignTx(account, tx, chainID)
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Gas      *hexutil.Uint64 `json:"gas"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Value    *hexutil.Big    `json:"value"`
	Nonce    *hexutil.Uint64 `json:"nonce"`
	// We accept "data" and "input" for backwards-compatibility reasons. "input" is the
	// newer name and should be preferred by clients.
	Data  *hexutil.Bytes `json:"data"`
	Input *hexutil.Bytes `json:"input"`
}

// setDefaults is a helper function that fills in default values for unspecified tx fields.
func (args *SendTxArgs) setDefaults(ctx context.Context, b Backend) error {
	if args.Gas == nil {
		args.Gas = new(hexutil.Uint64)
		*(*uint64)(args.Gas) = 90000
	}
	if args.GasPrice == nil {
		price, err := b.SuggestPrice(ctx)
		if err != nil {
			return err
		}
		args.GasPrice = (*hexutil.Big)(price)
	}
	if args.Value == nil {
		args.Value = new(hexutil.Big)
	}
	if args.Nonce == nil {
		nonce, err := b.GetPoolNonce(ctx, args.From)
		if err != nil {
			return err
		}
		args.Nonce = (*hexutil.Uint64)(&nonce)
	}
	if args.Data != nil && args.Input != nil && !bytes.Equal(*args.Data, *args.Input) {
		return errors.New(`Both "data" and "input" are set and not equal. Please use "input" to pass transaction call data.`)
	}
	if args.To == nil {
		// Contract creation
		var input []byte
		if args.Data != nil {
			input = *args.Data
		} else if args.Input != nil {
			input = *args.Input
		}
		if len(input) == 0 {
			return errors.New(`contract creation without any data provided`)
		}
	}
	return nil
}

func (args *SendTxArgs) toTransaction() *types.Transaction {
	var input []byte
	if args.Data != nil {
		input = *args.Data
	} else if args.Input != nil {
		input = *args.Input
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

// submitTransaction is a helper function that submits tx to txPool and logs a message.
func submitTransaction(ctx context.Context, b Backend, tx *types.Transaction) (common.Hash, error) {
	if err := b.SendTx(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	if tx.To() == nil {
		signer := types.MakeSigner(b.ChainConfig(), b.CurrentBlock().Number())
		//from, err := types.Sender(signer, tx)
		_, err := types.Sender(signer, tx)
		if err != nil {
			return common.Hash{}, err
		}
		//addr := crypto.CreateAddress(from, tx.Nonce())
		//log.Info("Submitted contract creation", "fullhash", tx.Hash().Hex(), "contract", addr.Hex())
	} else {
		//log.Info("Submitted transaction", "fullhash", tx.Hash().Hex(), "recipient", tx.To())
	}
	return tx.Hash(), nil
}

// SendTransaction creates a transaction for the given argument, sign it and submit it to the
// transaction pool.
func (s *PublicTransactionPoolAPI) SendTransaction(ctx context.Context, args SendTxArgs) (common.Hash, error) {
	if !utopia.Config.Bool("debug") {
		return common.Hash{}, errors.New("forbid when not debug mode")
	}
	log.Debug("send transaction", "args", args)
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: args.From}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return common.Hash{}, err
	}

	if args.Nonce == nil {
		// Hold the addresse's mutex around signing to prevent concurrent assignment of
		// the same nonce to multiple accounts.
		s.nonceLock.LockAddr(args.From)
		defer s.nonceLock.UnlockAddr(args.From)
	}

	// Set some sanity defaults and terminate on failure
	if err := args.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	// Assemble the transaction and sign with the wallet
	tx := args.toTransaction()

	var chainID *big.Int
	if config := s.b.ChainConfig(); config.IsEIP155(s.b.CurrentBlock().Number()) {
		chainID = config.ChainID
	}
	signed, err := wallet.SignTx(account, tx, chainID)
	if err != nil {
		return common.Hash{}, err
	}

	//filter tx
	if err := Filter(signed, ctx, s.b); err != nil {
		return common.Hash{}, err
	}

	return submitTransaction(ctx, s.b, signed)
}

//get all deployed contract desc
func (s *PublicTransactionPoolAPI) GetAllContractDesc(ctx context.Context) (list []vm.ContractDesc) {
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		log.Error("get state db", "err", err)
		return
	}
	list, err = vm.GetAllContract(stateDB)
	if err != nil {
		log.Error("get all contract", "err", err)
		return
	}

	return
}

//set node crt to consortium dir
func (s *PublicTransactionPoolAPI) SetCrt(ctx context.Context, jwt string, cert string) (success string, err error) {
	if !checkJWT(jwt, "a") {
		return "", errors.New("authorization failed")
	}

	//check cert
	_, _, _, _, _, err = utopia.ConsortiumSet(cert)
	if err != nil {
		return "", fmt.Errorf("check cert: %v", err)
	}

	//save cert
	localhostCrt := utopia.DataDir + "/consortium/localhost.crt"
	crtFile, err := os.OpenFile(localhostCrt, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer crtFile.Close()
	if err != nil {
		return "", fmt.Errorf("open file: %v", err)
	}
	_, err = crtFile.Write([]byte(cert))
	if err != nil {
		return "", fmt.Errorf("write file: %v", err)
	}

	return localhostCrt, nil
}

//generate node csr for node cert
func (s *PublicTransactionPoolAPI) GenCsr(ctx context.Context, jwt string, days int, subj string) (csr string, err error) {
	log.Info("gen csr")
	if !checkJWT(jwt, "a") {
		return "", errors.New("authorization failed")
	}

	if days == 0 {
		days = 365
	}

	etherBase := utopia.Config.String("etherbase")

	//a preString : 30740201010420
	preString := "30740201010420"
	//a midString : a00706052b8104000aa144034200 (identifies secp256k1)
	midString := "a00706052b8104000aa144034200"

	ks := s.b.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	unlocked := ks.GetUnlocked(common.HexToAddress(etherBase))

	//the privKey  : (32 bytes as 64 hexits)
	privKey := fmt.Sprintf("%x", crypto.FromECDSA(unlocked.PrivateKey))

	pubKeyObj := unlocked.PrivateKey.PublicKey
	//the pubKey   : (65 bytes as 130 hexits)
	pubKey := fmt.Sprintf("%x", crypto.FromECDSAPub(&pubKeyObj))

	pemFile := utopia.DataDir + "/node.pem"
	cmdPem := fmt.Sprintf("echo %s %s %s %s | xxd -r -p | openssl ec -inform d > %s", preString, privKey, midString, pubKey, pemFile)
	result, err := exec.Command("/bin/bash", "-c", cmdPem).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("generate pem [fail]: %v [reason]: %s [cmd]: %s", err, result, cmdPem)
	}

	if subj == "" {
		subj = "/miner=" + pubKey
	}

	cmdCsr := fmt.Sprintf(`openssl req -new -days %d -key %s -subj "%s"`, days, pemFile, subj)
	result, err = exec.Command("/bin/bash", "-c", cmdCsr).Output()
	if err != nil {
		result, err = exec.Command("/bin/bash", "-c", cmdCsr).CombinedOutput() // debug error
		return "", fmt.Errorf("generate node csr [fail]: %v [reson]: %s [cmd]: %s", err, result, cmdCsr)
	}
	return string(result), nil
}

// SendRawTransaction will add the signed transaction to the transaction pool.
// The sender is responsible for signing the transaction and using the correct nonce.
//extensions include name, version, jwt
func (s *PublicTransactionPoolAPI) SendRawTransaction(ctx context.Context, encodedTx hexutil.Bytes) (common.Hash, error) {
	//log.Info("into sendRawTransaction")
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(encodedTx, tx); err != nil {
		log.Error("rlp decode bytes error", "err", err)
		return common.Hash{}, err
	}

	//check token
	if utopia.Consortium {
		stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
		if err != nil {
			log.Error("get state db", "err", err)
			return common.Hash{}, fmt.Errorf("get state db:%v", err)
		}
		//check freeze
		var signer types.Signer

		if tx.IsSM2() {
			signer = types.NewSm2Signer(s.b.ChainConfig().ChainID)
		} else {
			signer = types.NewEIP155Signer(s.b.ChainConfig().ChainID)
		}
		sender, err := types.Sender(signer, tx)
		if err != nil {
			return common.Hash{}, err
		}
		if CheckFreeze(tx, stateDB, sender) {
			return common.Hash{}, errors.New("account has been frozen")
		}

		//parse token from tx payload
		ok := CheckToken(tx, stateDB)
		if !ok {
			return common.Hash{}, errors.New("authorization fail")
		}
	}
	//log.Debug("tx", "amount", tx.Value().Uint64(), "price", tx.GasPrice().Uint64(), "limit", tx.Gas())

	//filter tx
	if err := Filter(tx, ctx, s.b); err != nil {
		return common.Hash{}, err
	}

	return submitTransaction(ctx, s.b, tx)
}

func (s *PublicTransactionPoolAPI) WaitComplete(ctx context.Context) error {
	//WaitTime<-1
	log.Info("合约启动完毕")
	return nil
}

func (s *PublicTransactionPoolAPI) BaapQuery(ctx context.Context, encodedTx hexutil.Bytes) (string, error) {
	stateDB, _, err := s.b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		log.Error("get stateDB error", "err", err)
		return "", errors.New("get stateDB error")
	}
	return pdxcc.BaapQuery(stateDB, encodedTx)
}

// Sign calculates an ECDSA signature for:
// keccack256("\x19Ethereum Signed Message:\n" + len(message) + message).
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
//
// The account associated with addr must be unlocked.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
func (s *PublicTransactionPoolAPI) Sign(addr common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr}

	wallet, err := s.b.AccountManager().Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the requested hash with the wallet
	signature, err := wallet.SignHash(account, signHash(data))
	if err == nil {
		signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	}
	return signature, err
}

// SignTransactionResult represents a RLP encoded signed transaction.
type SignTransactionResult struct {
	Raw hexutil.Bytes      `json:"raw"`
	Tx  *types.Transaction `json:"tx"`
}

// SignTransaction will sign the given transaction with the from account.
// The node needs to have the private key of the account corresponding with
// the given from address and it needs to be unlocked.
func (s *PublicTransactionPoolAPI) SignTransaction(ctx context.Context, args SendTxArgs) (*SignTransactionResult, error) {
	if args.Gas == nil {
		return nil, fmt.Errorf("gas not specified")
	}
	if args.GasPrice == nil {
		return nil, fmt.Errorf("gasPrice not specified")
	}
	if args.Nonce == nil {
		return nil, fmt.Errorf("nonce not specified")
	}
	if err := args.setDefaults(ctx, s.b); err != nil {
		return nil, err
	}
	tx, err := s.sign(args.From, args.toTransaction())
	if err != nil {
		return nil, err
	}
	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	return &SignTransactionResult{data, tx}, nil
}

// PendingTransactions returns the transactions that are in the transaction pool
// and have a from address that is one of the accounts this node manages.
func (s *PublicTransactionPoolAPI) PendingTransactions() ([]*RPCTransaction, error) {
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return nil, err
	}
	accounts := make(map[common.Address]struct{})
	for _, wallet := range s.b.AccountManager().Wallets() {
		for _, account := range wallet.Accounts() {
			accounts[account.Address] = struct{}{}
		}
	}
	transactions := make([]*RPCTransaction, 0, len(pending))
	for _, tx := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(s.b.ChainConfig().ChainID)
		}
		if s.b.ChainConfig().Sm2Crypto {
			signer = types.NewSm2Signer(s.b.ChainConfig().ChainID)
		}
		from, _ := types.Sender(signer, tx)
		if _, exists := accounts[from]; exists {
			transactions = append(transactions, newRPCPendingTransaction(tx))
		}
	}
	return transactions, nil
}

// Resend accepts an existing transaction and a new gas price and limit. It will remove
// the given transaction from the pool and reinsert it with the new gas price and limit.
func (s *PublicTransactionPoolAPI) Resend(ctx context.Context, sendArgs SendTxArgs, gasPrice *hexutil.Big, gasLimit *hexutil.Uint64) (common.Hash, error) {
	if sendArgs.Nonce == nil {
		return common.Hash{}, fmt.Errorf("missing transaction nonce in transaction spec")
	}
	if err := sendArgs.setDefaults(ctx, s.b); err != nil {
		return common.Hash{}, err
	}
	matchTx := sendArgs.toTransaction()
	pending, err := s.b.GetPoolTransactions()
	if err != nil {
		return common.Hash{}, err
	}

	for _, p := range pending {
		var signer types.Signer = types.HomesteadSigner{}
		if p.Protected() {
			signer = types.NewEIP155Signer(p.ChainId())
		}
		if s.b.ChainConfig().Sm2Crypto {
			signer = types.NewSm2Signer(p.ChainId())
		}
		wantSigHash := signer.Hash(matchTx)

		if pFrom, err := types.Sender(signer, p); err == nil && pFrom == sendArgs.From && signer.Hash(p) == wantSigHash {
			// Match. Re-sign and send the transaction.
			if gasPrice != nil && (*big.Int)(gasPrice).Sign() != 0 {
				sendArgs.GasPrice = gasPrice
			}
			if gasLimit != nil && *gasLimit != 0 {
				sendArgs.Gas = gasLimit
			}
			signedTx, err := s.sign(sendArgs.From, sendArgs.toTransaction())
			if err != nil {
				return common.Hash{}, err
			}
			if err = s.b.SendTx(ctx, signedTx); err != nil {
				return common.Hash{}, err
			}
			return signedTx.Hash(), nil
		}
	}

	return common.Hash{}, fmt.Errorf("Transaction %#x not found", matchTx.Hash())
}

// PublicDebugAPI is the collection of Ethereum APIs exposed over the public
// debugging endpoint.
type PublicDebugAPI struct {
	b Backend
}

// NewPublicDebugAPI creates a new API definition for the public debug methods
// of the Ethereum service.
func NewPublicDebugAPI(b Backend) *PublicDebugAPI {
	return &PublicDebugAPI{b: b}
}

// GetBlockRlp retrieves the RLP encoded for of a single block.
func (api *PublicDebugAPI) GetBlockRlp(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	encoded, err := rlp.EncodeToBytes(block)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", encoded), nil
}

// PrintBlock retrieves a block and returns its pretty printed form.
func (api *PublicDebugAPI) PrintBlock(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return spew.Sdump(block), nil
}

// SeedHash retrieves the seed hash of a block.
func (api *PublicDebugAPI) SeedHash(ctx context.Context, number uint64) (string, error) {
	block, _ := api.b.BlockByNumber(ctx, rpc.BlockNumber(number))
	if block == nil {
		return "", fmt.Errorf("block #%d not found", number)
	}
	return fmt.Sprintf("0x%x", ethash.SeedHash(number)), nil
}

// PrivateDebugAPI is the collection of Ethereum APIs exposed over the private
// debugging endpoint.
type PrivateDebugAPI struct {
	b Backend
}

// NewPrivateDebugAPI creates a new API definition for the private debug methods
// of the Ethereum service.
func NewPrivateDebugAPI(b Backend) *PrivateDebugAPI {
	return &PrivateDebugAPI{b: b}
}

// ChaindbProperty returns leveldb properties of the chain database.
func (api *PrivateDebugAPI) ChaindbProperty(property string) (string, error) {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return "", fmt.Errorf("chaindbProperty does not work for memory databases")
	}
	if property == "" {
		property = "leveldb.stats"
	} else if !strings.HasPrefix(property, "leveldb.") {
		property = "leveldb." + property
	}
	return ldb.LDB().GetProperty(property)
}

func (api *PrivateDebugAPI) ChaindbCompact() error {
	ldb, ok := api.b.ChainDb().(interface {
		LDB() *leveldb.DB
	})
	if !ok {
		return fmt.Errorf("chaindbCompact does not work for memory databases")
	}
	for b := byte(0); b < 255; b++ {
		log.Info("Compacting chain database", "range", fmt.Sprintf("0x%0.2X-0x%0.2X", b, b+1))
		err := ldb.LDB().CompactRange(util.Range{Start: []byte{b}, Limit: []byte{b + 1}})
		if err != nil {
			log.Error("Database compaction failed", "err", err)
			return err
		}
	}
	return nil
}

// SetHead rewinds the head of the blockchain to a previous block.
func (api *PrivateDebugAPI) SetHead(number hexutil.Uint64) {
	api.b.SetHead(uint64(number))
}

// PublicNetAPI offers network related RPC methods
type PublicNetAPI struct {
	net            *p2p.Server
	networkVersion uint64
}

// NewPublicNetAPI creates a new net API instance.
func NewPublicNetAPI(net *p2p.Server, networkVersion uint64) *PublicNetAPI {
	return &PublicNetAPI{net, networkVersion}
}

// Listening returns an indication if the node is listening for network connections.
func (s *PublicNetAPI) Listening() bool {
	return true // always listening
}

// PeerCount returns the number of connected peers
func (s *PublicNetAPI) PeerCount() hexutil.Uint {
	return hexutil.Uint(s.net.PeerCount())
}

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", s.networkVersion)
}

func Filter(tx *types.Transaction, ctx context.Context, b Backend) error {
	if tx.GasPrice().Cmp(big.NewInt(0)) <= 0 {
		return errors.New("gas price too low")
	}
	to := tx.To()
	switch {
	case utils.IsBaapTrustTree(to): //trust tx white list
		//from
		var signer types.Signer
		signer = types.NewEIP155Signer(b.ChainConfig().ChainID)
		if tx.IsSM2() {
			signer = types.NewSm2Signer(b.ChainConfig().ChainID)
		}
		from, err := types.Sender(signer, tx)
		if err != nil {
			return err
		}
		//chainid
		ptx := &protos.Transaction{}
		err = proto.Unmarshal(tx.Data(), ptx)
		if err != nil {
			return fmt.Errorf("tx data not right")
		}
		invocation := &protos.Invocation{}
		err = proto.Unmarshal(ptx.Payload, invocation)
		if err != nil {
			return fmt.Errorf("proto unmarshal invocation err:%v", err)
		}
		chainid, ok := invocation.Meta["chain-id"]
		if !ok {
			return fmt.Errorf("need chain-id in meta")
		}
		//control key
		controlKey := "trustTree:" + string(chainid) + ":" + from.String()
		//黑名单控制
		if err := trustTreeBlacklistControl(controlKey); err != nil {
			return err
		}

		//白名单控制
		if err := whitelist.Verify(from); err != nil {
			return err
		}
	case utils.IsDeposit(to): //blacklist and frequency filter
		//get ip
		var controlKey = "deposit:"
		ip, ok := ctx.Value("ip").(string)
		if !ok || ip == "" {
			log.Warn("get ip", "ok", ok, "ip", ip)
			//var signer types.Signer = types.HomesteadSigner{}
			//if tx.Protected() {
			//	signer = types.NewEIP155Signer(tx.ChainId())
			//}
			//from, err := types.Sender(signer, tx)
			//if err != nil {
			//	return fmt.Errorf("deposit get sender:%v", err)
			//}
			var signer types.Signer
			signer = types.NewEIP155Signer(b.ChainConfig().ChainID)
			if tx.IsSM2() {
				signer = types.NewSm2Signer(b.ChainConfig().ChainID)
			}

			from, err := types.Sender(signer, tx)
			if err != nil {
				return err
			}
			controlKey += from.Hex()
		} else {
			controlKey += ip
		}

		//控制频率
		if err := frequency.FrequencyControl(controlKey, frequency.ReqIntervalLimit); err != nil {
			return fmt.Errorf("frequency control:%v", err)
		}

		//黑名单过滤
		if err := depositBlacklistControl(ctx, tx, b); err != nil {
			return fmt.Errorf("blacklist control:%v", err)
		}

	case utils.IsTcUpdater(to): //verify cert
		if !tx.VerifyCert(1, b.ChainConfig().ChainID) {
			return errors.New("IsTcUpdater cert illegal")
		}
	case utils.IsTrustTxWhiteList(to): //verify cert
		if !tx.VerifyCert(2, b.ChainConfig().ChainID) {
			return errors.New("IsTrustTxWhiteList cert illegal")
		}
	case utils.IsNodeUpdate(to): //verify cert
		if !tx.VerifyCert(3, b.ChainConfig().ChainID) {
			return errors.New("IsNodeUpdate cert illegal")
		}
	}
	return nil
}

func depositBlacklistControl(ctx context.Context, tx *types.Transaction, b Backend) error {
	//是否在本地黑名单中
	//var signer types.Signer = types.HomesteadSigner{}
	//if tx.Protected() {
	//	signer = types.NewEIP155Signer(tx.ChainId())
	//}
	//sender, err := types.Sender(signer, tx)
	//if err != nil {
	//	return fmt.Errorf("get sender:%v", err)
	//}
	var signer types.Signer
	signer = types.NewEIP155Signer(b.ChainConfig().ChainID)
	if tx.IsSM2() {
		signer = types.NewSm2Signer(b.ChainConfig().ChainID)
	}
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return err
	}
	v, _ := blacklist.BlacklistControlMap.Load(sender)
	if v != nil {
		expiredBlockNumber, _ := v.(*big.Int)
		if expiredBlockNumber.Cmp(b.CurrentBlock().Number()) > 0 {
			return fmt.Errorf("in the blacklist util blockNumber: %s", expiredBlockNumber.String())
		}
	}

	//是否在状态黑名单中
	stateDB, _, err := b.StateAndHeaderByNumber(ctx, rpc.LatestBlockNumber)
	if err != nil {
		return fmt.Errorf("get state db:%v", err)
	}
	contractAddr := tx.To()
	blacklistKeyHash := utils.DepositBlacklistKeyHash(sender)
	result := stateDB.GetPDXState(*contractAddr, blacklistKeyHash)
	if result != nil {
		expiredBlockNumber := big.NewInt(0).SetBytes(result)
		if expiredBlockNumber.Cmp(b.CurrentBlock().Number()) > 0 {
			return fmt.Errorf("in blacklist util blockNumber: %s", expiredBlockNumber.String())
		}
	}

	//两次交易的发送者是否一致
	var payload vm.Deposit
	err = json.Unmarshal(tx.Data(), &payload)
	if err != nil {
		return fmt.Errorf("payload unmarshal:%v", err)
	}

	tx1From, _, _, tx1Hash, err := vm.GenMsgFromTxMsg(payload.TxMsg)
	if err != nil {
		return fmt.Errorf("get msg tx1From txMsg:%v", err)
	}

	//expiredBlockNumber := big.NewInt(0).Add(currentBlockNum, big.NewInt(int64(blacklist.BadExpiredBlockNum)))
	if sender != tx1From {
		//blacklist.BlacklistControlMap.Store(sender, expiredBlockNumber)
		log.Error("different sender between two tx", "expired", blacklist.BadExpired)
		//return fmt.Errorf("different sender between two tx, put in blacklist for %d millisecond", blacklist.BadExpired)
		return nil
	}

	//确认是否已经存款成功
	keyHash := utils.DepositKeyHash(sender, tx1Hash, conf.PDXKeyFlag)
	state := stateDB.GetPDXState(*contractAddr, keyHash)
	if len(state) != 0 {
		//blacklist.BlacklistControlMap.Store(sender, expiredBlockNumber)
		log.Error("deposit has completed", "expired", blacklist.BadExpired)
		//return fmt.Errorf("deposit has completed, put in blacklist for %d millisecond", blacklist.BadExpired)
		return nil
	}

	dstChains := b.ChainConfig().Utopia.DstChain
	_, err = vm.CheckWithDrawTxStatus(sender, dstChains, tx1Hash, payload.SrcChainID)
	if err != nil {
		return fmt.Errorf("check withdraw tx status:%v", err)
	}

	return nil
}

func trustTreeBlacklistControl(controlKey string) (err error) {
	//是否在黑名单中
	v, _ := blacklist.BlacklistControlMap.Load(controlKey)
	if v != nil {
		expiredBlockNumber, _ := v.(*big.Int)
		if expiredBlockNumber.Cmp(public.BC.CurrentBlock().Number()) > 0 {
			return fmt.Errorf("in the blacklist util blockNumber: %s", expiredBlockNumber.String())
		}
	}

	//是否加入黑名单
	if err := frequency.FrequencyControl(controlKey, frequency.ReqIntervalLimit); err != nil {
		//put in blacklist
		blacklist.PutIn(controlKey)
		return fmt.Errorf("frequency control:%v", err)
	}

	return nil
}

// add by liangc >>>>
// 躲避循环引用，我们使用一个 channel 来直接调用 call
func (s *PublicBlockChainAPI) loop() {
	utils.Precreate.Start()
	for request := range utils.Precreate.TaskCh {
		go func(req utils.PrecreateRequest) {
			if req.RespCh == nil {
				panic("TODO: precreate_contract_error : response_channel_nil")
			}
			select {
			case req.RespCh <- s.CallFromTxPool(request):
			case <-time.After(15 * time.Second):
				panic("TODO: precreate_contract_timeout : debug_info")
			}
		}(request)
	}
}

func (s *PublicBlockChainAPI) CallFromTxPool(request utils.PrecreateRequest) *utils.PrecreateResponse {
	args := CallArgs{request.From, request.To, request.Gas, request.GasPrice, request.Value, request.Data, false}
	state, result, gas, failed, err := s.doCallRtnState(context.Background(), request.Hash, args, rpc.PendingBlockNumber, vm.Config{}, 10*time.Second)
	logs := make([]*utils.Log, 0)
	stateMap := make(map[common.Address]map[common.Hash]common.Hash)
	statePDXMap := make(map[common.Address]map[common.Hash][]byte)
	// 在执行 call 时已经计算过一次 nonce，不想取下一个 nonce ，所以这里要减 1
	// 不需要了，要深度复制这个临时 state 中的全部状态
	// contractAddr := crypto.CreateAddress(request.From, state.GetNonce(request.From)-1)
	_logs := state.GetLogs(request.Hash)

	for _, _log := range _logs {
		// 这么手写，比 rlp 速度快 哈哈
		logs = append(logs, &utils.Log{_log.Address, _log.Topics, _log.Data, _log.BlockNumber, _log.TxHash, _log.TxIndex, _log.BlockHash, _log.Index, _log.Removed})
	}

	for addr, _ := range state.StateObjects() {
		// 复制状态
		_stateMap := make(map[common.Hash]common.Hash)
		_statePDXMap := make(map[common.Hash][]byte)
		// 复制 pdx 状态
		if vm.DeployedContractAccountAddr != addr {
			state.ForEachStorage(addr, func(k, v common.Hash) bool {
				_stateMap[k] = v
				return true
			})
			state.ForEachPDXStorage(addr, func(k common.Hash, v []byte) bool {
				_statePDXMap[k] = v
				delete(stateMap, addr)
				return true
			})
		}
		stateMap[addr] = _stateMap
		statePDXMap[addr] = _statePDXMap
	}
	pendingNumber := s.b.CurrentBlock().Number().Uint64() + 1
	return &utils.PrecreateResponse{pendingNumber, logs, stateMap, statePDXMap, result, gas, failed, err}
}

// add by liangc <<<<