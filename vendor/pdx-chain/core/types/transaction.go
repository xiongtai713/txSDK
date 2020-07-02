// Copyright 2014 The go-ethereum Authors
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

package types

import (
	"container/heap"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"math/big"
	"os"
	"os/exec"
	"pdx-chain/pdxcc/protos"
	"pdx-chain/utopia/utils"
	"strings"
	"sync/atomic"

	"pdx-chain/common"
	"pdx-chain/common/hexutil"
	"pdx-chain/crypto"
	"pdx-chain/log"
	"pdx-chain/rlp"
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

const (
	Transaction_invoke = 1
	Transaction_deploy = 2
)

var (
	ErrInvalidSig = errors.New("invalid transaction v, r, s values")
	BaapHome      = os.Getenv("PDX_BAAP_HOME")
)

type Transaction struct {
	data txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	GasLimit     hexutil.Uint64
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
}

func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return newTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

func newTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &Transaction{data: d}
}

// add by liangc : 当交易需要广播时，如果是 wasm 合约需要重新对 payload 封包
func (tx *Transaction) CloneAndResetPayload(payload []byte) (*Transaction, error) {
	buff, err := rlp.EncodeToBytes(tx)
	if err != nil {
		log.Error("encode-tx-error", "err", err)
		return nil, err
	}
	ntx := new(Transaction)
	err = rlp.DecodeBytes(buff, &ntx.data)
	if err != nil {
		log.Error("decode-tx-error", "err", err)
		return nil, err
	}
	ntx.data.Payload = payload
	return ntx, nil
}

func (tx *Transaction) getCertFromTcupdate() (tcAdmCrt string, err error) {
	type TrustChainUpdTxModel struct {
		ChainId        string   `json:"ChainId"`
		ChainOwner     string   `json:"chainOwner"`
		Timestamp      int64    `json:"timestamp"`
		Random         string   `json:"random"`
		TcAdmCrt       string   `json:"tcadmCrt"`
		Signature      string   `json:"signature"`
		EnodeList      []string `json:"enodeList"`
		HostList       []string `json:"hostList"`
		PrevChainID    string   `json:"prevChainID"`
		PrevChainOwner string   `json:"prevChainOwner"`
		SelfHost       string   `json:"selfHost"`
	}
	ptx := &protos.Transaction{}
	err = proto.Unmarshal(tx.data.Payload, ptx)
	if err != nil {
		log.Error("unmarshal finalTx error", "err", err)
		return
	}
	if ptx.Type != Transaction_invoke {
		log.Error("tx type not invoke error")
		return
	}
	invoke := &protos.Invocation{}
	err = proto.Unmarshal(ptx.Payload, invoke)
	if err != nil {
		log.Error("unmarshal payload error", "err", err)
		return
	}

	if len(invoke.Args) != 1 {
		log.Error("params is empty")
		return
	}
	var pl TrustChainUpdTxModel
	err = json.Unmarshal(invoke.Args[0], &pl)
	if err != nil {
		log.Error("unmarshal payload error", "err", err)
		return
	}
	//log.Info("TrustChainUpdTxModel", "is", pl)
	//log.Info("cert is", ":", pl.TcAdmCrt)
	return pl.TcAdmCrt, nil
}

func (tx *Transaction) getCertFromWhiteList() (tcAdmCrt string, err error) {
	type WhiteList struct {
		Timestamp int64    `json:"timestamp"`
		Random    string   `json:"random"`
		TcAdmCrt  string   `json:"tcadmCrt"`
		Nodes     []string `json:"nodes"`
		Type      int      `json:"type"`
	}
	whitelist := &WhiteList{}
	err = json.Unmarshal(tx.data.Payload, whitelist)
	if err != nil {
		log.Error("unmarshal Payload error", "err", err)
		return
	}

	return whitelist.TcAdmCrt, nil
}

func (tx *Transaction) getCertFromNodeUpdate() (tcAdmCrt string, err error) {
	type ConsensusNodeUpdate struct {
		NodeType     int              `json:"nodeType"`
		FromChainID  uint64           `json:"fromChainID"`
		ToChainID    uint64           `json:"toChainID"`
		Cert         string           `json:"cert"`
		Address      []common.Address `json:"address"`
		CommitHeight uint64           `json:"commitHeight"`
		NuAdmCrt     string           `json:"nuAdmCrt"`
	}
	NodeUpdate := &ConsensusNodeUpdate{}
	err = json.Unmarshal(tx.data.Payload, NodeUpdate)
	if err != nil {
		log.Error("unmarshal Payload error", "err", err)
		return
	}
	return NodeUpdate.Cert, nil
}

//verify cert
func (tx *Transaction) VerifyCert(category int, chainId *big.Int) (ok bool) {
	//get cert
	var (
		tcAdmCrt string
		err      error
	)

	switch category {
	case 1:
		tcAdmCrt, err = tx.getCertFromTcupdate()
	case 2:
		tcAdmCrt, err = tx.getCertFromWhiteList()
	case 3:
		tcAdmCrt, err = tx.getCertFromNodeUpdate()

	}
	if err != nil {
		return
	}

	//check publicKey
	toPub := GenPubKeyFromIAASCert(tcAdmCrt)
	if toPub == nil {
		log.Error(" gen pubkey from cert fail")
		return
	}
	pubkey := crypto.CompressPubkey(toPub)

	signer := NewEIP155Signer(chainId)
	txPubkey, err := signer.SenderPubkey(tx)
	if err != nil {
		log.Error("sender public key", "err", err)
		return
	}

	txPubkeyHex := hex.EncodeToString(crypto.CompressPubkey(txPubkey))
	certPubkeyHex := hex.EncodeToString(pubkey)
	if txPubkeyHex != certPubkeyHex {
		log.Error("cert public key no match", "tx pubkey", txPubkeyHex, "certPubkey", certPubkeyHex)
		return
	}

	//save cert for check
	//BaapHome := os.Getenv("PDX_BAAP_HOME")
	if !CheckIAASCert(tcAdmCrt, "tcadmUnity.crt") {
		log.Error("check IAAS cert fail")
		return
	}

	return true
}

func CheckIAASCert(cert string, tmpCertName string) (ok bool) {
	if BaapHome == "" {
		log.Error("BaapHome is empty")
		BaapHome = "/pdx"
	}
	conf := BaapHome + "/conf/"
	userCert := conf + tmpCertName
	f, err := os.OpenFile(userCert, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer f.Close()
	if err != nil {
		log.Error("open error", "err", err)
		return
	}
	_, err = f.Write([]byte(cert))
	if err != nil {
		log.Error("write error", "err", err)
		return
	}

	//check certificate chain
	rootCert := conf + "ca.crt"
	interCert := conf + "trust-signer.crt"
	shell := fmt.Sprintf("openssl verify -CAfile %s -untrusted %s %s", rootCert, interCert, userCert)
	out, err := exec.Command("/bin/bash", "-c", shell).CombinedOutput()
	if err != nil {
		log.Error("command error", "err", err, "out", string(out))
		return
	}
	log.Info("exec command", "out", string(out))
	if string(out) == userCert+": OK\n" {
		log.Info("cert chain verify success")
		return true
	}

	return
}

func GenPubKeyFromIAASCert(cert string) (pubKey *ecdsa.PublicKey) {
	s := strings.Split(cert, "pub:")
	if len(s) != 2 {
		log.Error("TcAdmCrt pub format err")
		return
	}
	news := strings.Split(s[1], "ASN1 OID: secp256k1")
	if len(news) != 2 {
		log.Error("TcAdmCrt secp256k1 format err")
		return
	}
	noEmpty := strings.Replace(news[0], " ", "", -1)
	noReturn := strings.Replace(noEmpty, "\n", "", -1)
	noColon := strings.Replace(noReturn, ":", "", -1)

	b, err := hex.DecodeString(noColon)
	if err != nil {
		log.Error("decode hex err:", err.Error())
		return
	}
	toPub, err := crypto.UnmarshalPubkey(b)
	if err != nil {
		log.Error("unmarshal pubkey error", "err", err)
		return
	}

	return toPub
}

// ChainId returns which chain id this transaction was signed for (if at all)
func (tx *Transaction) ChainId() *big.Int {
	return deriveChainId(tx.data.V)
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *Transaction) Protected() bool {
	return isProtectedV(tx.data.V)
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 is considered protected
	return true
}

func (tx *Transaction) IsSM2() bool {
	return tx.data.V.BitLen() > 64
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		log.Error("err", "err", err)
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		chainID := deriveChainId(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return ErrInvalidSig
	}
	*tx = Transaction{data: dec}
	return nil
}

func (tx *Transaction) Data() []byte       { return common.CopyBytes(tx.data.Payload) }
func (tx *Transaction) Gas() uint64        { return tx.data.GasLimit }
func (tx *Transaction) GasPrice() *big.Int { return new(big.Int).Set(tx.data.Price) }
func (tx *Transaction) Value() *big.Int    { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Nonce() uint64      { return tx.data.AccountNonce }
func (tx *Transaction) CheckNonce() bool   { return true }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

//note:此方法在eip155签的交易，解出来的公钥不正确，改用eip155Signer.senderPubkey()
//func (tx *Transaction) PublicKey() []byte {
//
//	var V byte
//	if isProtectedV(tx.data.V) {
//		chainID := deriveChainId(tx.data.V).Uint64()
//		V = byte(tx.data.V.Uint64() - 35 - 2*chainID)
//	} else {
//		V = byte(tx.data.V.Uint64() - 27)
//	}
//
//	/*var sig []byte
//	sig = append(append(append(sig, tx.data.R.Bytes()...), tx.data.S.Bytes()...), V)*/
//
//	// encode the snature in uncompressed format
//	R, S := tx.data.R.Bytes(), tx.data.S.Bytes()
//	sig := make([]byte, 65)
//	copy(sig[32-len(R):32], R)
//	copy(sig[64-len(S):64], S)
//	sig[64] = V
//
//	bytes, err := rlp.EncodeToBytes([]interface{}{
//		tx.data.AccountNonce,
//		tx.data.Price, tx.data.GasLimit, tx.data.Recipient,
//		tx.data.Amount,
//		tx.data.Payload})
//
//	if err != nil {
//		log.Error(err.Error())
//		return nil
//	}
//	pub, err := crypto.Ecrecover(crypto.Keccak256(bytes), sig)
//	if err != nil {
//		log.Error(err.Error())
//		return nil
//	}
//
//	toPub, err := crypto.UnmarshalPubkey(pub)
//	if err != nil {
//		log.Error(err.Error())
//		return nil
//	}
//	pubkey := crypto.CompressPubkey(toPub)
//	//fmt.Println(hexutil.Encode(pubkey))
//
//	return pubkey
//}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *Transaction) AsMessage(s Signer) (Message, error) {
	msg := Message{
		nonce:      tx.data.AccountNonce,
		gasLimit:   tx.data.GasLimit,
		gasPrice:   new(big.Int).Set(tx.data.Price),
		to:         tx.data.Recipient,
		amount:     tx.data.Amount,
		data:       tx.data.Payload,
		checkNonce: true,
	}

	var err error
	msg.from, err = Sender(s, tx)
	return msg, err
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	if utils.IsFreeTx(tx.To()) {
		total = big.NewInt(0)
		return total
	}
	if utils.IsFreeGas() || utils.IsFreeTx(tx.To()) {
		total = big.NewInt(0)
	}
	total.Add(total, tx.data.Amount)
	return total
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set which is the difference between a and b.
func TxDifference(a, b Transactions) Transactions {
	keep := make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].data.AccountNonce < s[j].data.AccountNonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice Transactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].data.Price.Cmp(s[j].data.Price) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]Transactions // Per account nonce-sorted list of transactions
	heads  TxByPrice                       // Next transaction for each unique account (price heap)
	signer Signer                          // Signer for the set of transactions
}

// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]Transactions) *TransactionsByPriceAndNonce {
	// Initialize a price based heap with the head transactions
	heads := make(TxByPrice, 0, len(txs))
	for from, accTxs := range txs {
		if len(accTxs) == 0 {
			continue
		}
		heads = append(heads, accTxs[0])
		// Ensure the sender address is from the signer
		acc, _ := Sender(signer, accTxs[0])
		txs[acc] = accTxs[1:]
		if from != acc {
			delete(txs, from)
		}
	}
	heap.Init(&heads)

	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

func NewTransactionsByRecipient(txs []*Transaction) map[common.Address]Transactions {
	txsRecipient := make(map[common.Address]Transactions)
	for _, tx := range txs {
		to := tx.To()
		txsRecipient[*to] = append(txsRecipient[*to], tx)
	}
	return txsRecipient
}

func NewTransactionsByRecipientFromMap(maptxs map[common.Address]Transactions) map[common.Address]Transactions {
	txsRecipient := make(map[common.Address]Transactions)
	for _, txs := range maptxs {
		for _, tx := range txs {
			to := tx.To()
			txsRecipient[*to] = append(txsRecipient[*to], tx)
		}
	}
	return txsRecipient
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

func (t *TransactionsByPriceAndNonce) HeadSize() int {
	return len(t.heads)
}

func (t *TransactionsByPriceAndNonce) PopAccountTxs() Transactions {
	if len(t.heads) == 0 {
		return nil
	}
	accTxs := make(Transactions, 0)
	accTxs = append(accTxs, t.heads[0])

	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		accTxs = append(accTxs, txs...)

		heap.Pop(&t.heads)
		return accTxs
	}
	return nil
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to         *common.Address
	from       common.Address
	nonce      uint64
	amount     *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func NewMessage(from common.Address, to *common.Address, nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:       from,
		to:         to,
		nonce:      nonce,
		amount:     amount,
		gasLimit:   gasLimit,
		gasPrice:   gasPrice,
		data:       data,
		checkNonce: checkNonce,
	}
}

func (m Message) From() common.Address { return m.from }
func (m Message) To() *common.Address  { return m.to }
func (m Message) GasPrice() *big.Int   { return m.gasPrice }
func (m Message) Value() *big.Int      { return m.amount }
func (m Message) Gas() uint64          { return m.gasLimit }
func (m Message) Nonce() uint64        { return m.nonce }
func (m Message) Data() []byte         { return m.data }
func (m Message) CheckNonce() bool     { return m.checkNonce }
