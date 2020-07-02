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

// Package downloader contains the manual full chain synchronisation.
package downloader

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"pdx-chain/cacheBlock"
	"pdx-chain/core"
	"pdx-chain/eth/fetcher"
	"pdx-chain/examineSync"
	"pdx-chain/p2p"
	"pdx-chain/quorum"
	"pdx-chain/utopia"
	"pdx-chain/utopia/engine"
	"pdx-chain/utopia/engine/qualification"
	types2 "pdx-chain/utopia/types"
	"sync"
	"sync/atomic"
	"time"

	"pdx-chain"
	"pdx-chain/common"
	"pdx-chain/core/rawdb"
	"pdx-chain/core/types"
	"pdx-chain/ethdb"
	"pdx-chain/event"
	"pdx-chain/log"
	"pdx-chain/metrics"
	"pdx-chain/params"
)

var (
	MaxHashFetch         = 512 // Amount of hashes to be fetched per retrieval request
	MaxBlockFetch        = 128 // Amount of blocks to be fetched per retrieval request
	MaxHeaderFetch       = 64  // Amount of block headers to be fetched per retrieval request
	MaxSkeletonSize      = 128 // Number of header fetches to need for a skeleton assembly
	MaxBodyFetch         = 128 // Amount of block bodies to be fetched per retrieval request
	MaxReceiptFetch      = 256 // Amount of transaction receipts to allow fetching per request
	MaxStateFetch        = 384 // Amount of node state values to allow fetching per request
	MaxMaxHeaderFetchPDX = 0

	MaxForkAncestry  = 3 * params.EpochDuration // Maximum chain reorganisation
	rttMinEstimate   = 2 * time.Second          // Minimum round-trip time to target for download requests
	rttMaxEstimate   = 20 * time.Second         // Maximum round-trip time to target for download requests
	rttMinConfidence = 0.1                      // Worse confidence factor in our estimated RTT value
	ttlScaling       = 3                        // Constant scaling factor for RTT -> TTL conversion
	ttlLimit         = time.Minute              // Maximum TTL allowance to prevent reaching crazy timeouts

	qosTuningPeers   = 5    // Number of peers to tune based on (best peers)
	qosConfidenceCap = 10   // Number of peers above which not to modify RTT confidence
	qosTuningImpact  = 0.25 // Impact that a new tuning target has on the previous value

	maxQueuedHeaders  = 32 * 1024 // [eth/62] Maximum number of headers to queue for import (DOS protection)
	maxHeadersProcess = 2048      // Number of header download results to import at once into the chain
	maxResultsProcess = 2048      // Number of content download results to import at once into the chain

	fsHeaderCheckFrequency = 100             // Verification frequency of the downloaded headers during fast sync
	fsHeaderSafetyNet      = 2048            // Number of headers to discard in case a chain violation is detected
	fsHeaderForceVerify    = 24              // Number of headers to verify before and after the pivot to accept it
	fsHeaderContCheck      = 3 * time.Second // Time interval to check for header continuations during state download
	fsMinFullBlocks        = 64              // Number of blocks to retrieve fully even in fast sync

)

var (
	ErrCommitBusy              = errors.New("CommitBusy")
	errBusy                    = errors.New("busy")
	errUnknownPeer             = errors.New("peer is unknown or unhealthy")
	errBadPeer                 = errors.New("action from bad peer ignored")
	errStallingPeer            = errors.New("peer is stalling")
	errNoPeers                 = errors.New("no peers to keep download active")
	errTimeout                 = errors.New("timeout")
	errEmptyHeaderSet          = errors.New("empty header set by peer")
	errPeersUnavailable        = errors.New("no peers available or all tried for download")
	errInvalidAncestor         = errors.New("retrieved ancestor is invalid")
	errInvalidChain            = errors.New("retrieved hash chain is invalid")
	errInvalidBlock            = errors.New("retrieved block is invalid")
	errInvalidBody             = errors.New("retrieved block body is invalid")
	errInvalidReceipt          = errors.New("retrieved receipt is invalid")
	errCancelBlockFetch        = errors.New("block download canceled (requested)")
	errCancelHeaderFetch       = errors.New("block header download canceled (requested)")
	errCancelBodyFetch         = errors.New("block body download canceled (requested)")
	errCancelReceiptFetch      = errors.New("receipt download canceled (requested)")
	errCancelStateFetch        = errors.New("state data download canceled (requested)")
	errCancelHeaderProcessing  = errors.New("header processing canceled (requested)")
	errCancelContentProcessing = errors.New("content processing canceled (requested)")
	errNoSyncActive            = errors.New("no sync active")
	errTooOld                  = errors.New("peer doesn't speak recent enough protocol version (need version >= 62)")
)

const syncOrigin = 10

var IslandDel int32

type Downloader struct {
	mode SyncMode       // Synchronisation mode defining the strategy used (per sync cycle)
	Mux  *event.TypeMux // Event multiplexer to announce sync operation events

	queue   *queue   // Scheduler for selecting the hashes to download
	peers   *peerSet // Set of active peers from which download can proceed
	stateDB ethdb.Database

	rttEstimate   uint64 // Round trip time to target for download requests
	rttConfidence uint64 // Confidence in the estimated RTT (unit: millionths to allow atomic ops)

	// Statistics
	syncStatsChainOrigin  uint64 // Origin block number where syncing started at
	syncStatsChainHeight  uint64 // Highest block number known when syncing started
	syncStatsCommitHeight uint64 // 最新commit高度
	syncStatsState        stateSyncStats
	syncStatsLock         sync.RWMutex // RW protecting the sync stats fields

	lightchain  LightChain
	blockchain  BlockChain
	Commitchain CommitChain

	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving

	// Status
	synchroniseMock         func(id string, hash common.Hash) error // Replacement for synchronise during testing
	Synchronis              int32
	AssociatedSynchronising int32
	commitSynchronising     int32
	notified                int32
	committed               int32

	// Channels
	headerCh                  chan dataPack        // [eth/62] Channel receiving inbound block headers
	bodyCh                    chan dataPack        // [eth/62] Channel receiving inbound block bodies
	receiptCh                 chan dataPack        // [eth/63] Channel receiving inbound receipts
	bodyWakeCh                chan bool            // [eth/62] Channel to signal the block body fetcher of new tasks
	receiptWakeCh             chan bool            // [eth/63] Channel to signal the receipt fetcher of new tasks
	headerProcCh              chan []*types.Header // [eth/62] Channel to feed the header processor new tasks
	commitCh                  chan dataPack        // [eth/63] Channel receiving inbound commit blocks PDX
	processCommitCh           chan []*types.Block  // [eth/63] Channel receiving inbound commit blocks PDX
	ProcessCommitChAssociated chan []*types.Block
	SyncInBlockPathCh         chan []*types.Block
	IslandCH                  chan []*types.Block
	// for stateFetcher
	stateSyncStart chan *stateSync
	trackStateReq  chan *stateReq
	stateCh        chan dataPack // [eth/63] Channel receiving inbound node state data

	// Cancellation and termination
	cancelPeer string         // Identifier of the peer currently being used as the master (cancel on drop)
	cancelCh   chan struct{}  // Channel to cancel mid-flight syncs
	cancelLock sync.RWMutex   // RW to protect the cancel channel and peer in delivers
	cancelWg   sync.WaitGroup // Make sure all fetcher goroutines have exited.

	quitCh   chan struct{} // Quit channel to signal termination
	quitLock sync.RWMutex  // RW to prevent double closes

	// Testing hooks
	syncInitHook     func(uint64, uint64)  // Method to call upon initiating a new sync run
	bodyFetchHook    func([]*types.Header) // Method to call upon starting a block body fetch
	receiptFetchHook func([]*types.Header) // Method to call upon starting a receipt fetch
	chainInsertHook  func([]*fetchResult)  // Method to call upon inserting a chain of blocks (possibly in multiple invocations)
	Worker           Worker                // PDX
	sync             bool                  // PDX
	CbFetcher        *fetcher.CbFetcher    // PDX
	AssociatedCH     chan bool             // PDX
	//IsSyncNormalDeleteCommit int32                 //PDX
}

type CommitAssertionSum struct {
	AssertionSum *big.Int
	CommitHeight *big.Int
}

type originAssertionSum struct {
	originAssertion *big.Int
	originNum       *big.Int
}

// LightChain encapsulates functions required to synchronise a light chain.
type LightChain interface {
	// HasHeader verifies a header's presence in the local chain.
	HasHeader(common.Hash, uint64) bool

	// GetHeaderByHash retrieves a header from the local chain.
	GetHeaderByHash(common.Hash) *types.Header

	// CurrentHeader retrieves the head header from the local chain.
	CurrentHeader() *types.Header

	// GetTd returns the total difficulty of a local block.
	GetTd(common.Hash, uint64) *big.Int

	// InsertHeaderChain inserts a batch of headers into the local chain.
	InsertHeaderChain([]*types.Header, int) (int, error)

	// Rollback removes a few recently added elements from the local chain.
	Rollback([]common.Hash)
}

// BlockChain encapsulates functions required to sync a (full or fast) blockchain.
type BlockChain interface {
	LightChain

	// HasBlock verifies a block's presence in the local chain.
	HasBlock(common.Hash, uint64) bool

	// GetBlockByHash retrieves a block from the local chain.
	GetBlockByHash(common.Hash) *types.Block

	// CurrentBlock retrieves the head block from the local chain.
	CurrentBlock() *types.Block

	// CurrentFastBlock retrieves the head fast block from the local chain.
	CurrentFastBlock() *types.Block

	// FastSyncCommitHead directly commits the head block to a certain entity.
	FastSyncCommitHead(common.Hash) error

	// InsertChain inserts a batch of blocks into the local chain.
	InsertChain(types.Blocks) (int, error)

	// InsertReceiptChain inserts a batch of receipts into the local chain.
	InsertReceiptChain(types.Blocks, []types.Receipts) (int, error)

	SetHead(head uint64) error

	GetBlockByNumber(number uint64) *types.Block
}

type CommitChain interface {
	CurrentBlock() *types.Block
	HasBlock(common.Hash, uint64) bool
	GetBlockByNum(number uint64) *types.Block
	CommitChain() *core.CommitChain
	DeleteHeightMap(uint64)
	SearchingBlock(hash common.Hash, number uint64) *types.Block
}

type Worker interface {
	ProcessCommitBlock(block *types.Block, sync bool) error
	StopWaitSync()
	SyncJoin(block types.Block)
	Signer() common.Address
	Utopia() *engine.Utopia
	VerifyCommitBlock(block *types.Block, land engine.LocalLand, sync bool) error
	RollBackCommit(commitBlock *types.Block)
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(mode SyncMode, stateDb ethdb.Database, mux *event.TypeMux, chain BlockChain, lightchain LightChain, dropPeer peerDropFn) *Downloader {
	if lightchain == nil {
		lightchain = chain
	}

	dl := &Downloader{
		mode:                      mode,
		stateDB:                   stateDb,
		Mux:                       mux,
		queue:                     newQueue(),
		peers:                     newPeerSet(),
		rttEstimate:               uint64(rttMaxEstimate),
		rttConfidence:             uint64(1000000),
		blockchain:                chain,
		lightchain:                lightchain,
		dropPeer:                  dropPeer,
		headerCh:                  make(chan dataPack, 1),
		bodyCh:                    make(chan dataPack, 1),
		receiptCh:                 make(chan dataPack, 1),
		commitCh:                  make(chan dataPack, 1000),
		processCommitCh:           make(chan []*types.Block, 1000),
		ProcessCommitChAssociated: make(chan []*types.Block, 1000),
		bodyWakeCh:                make(chan bool, 1),
		receiptWakeCh:             make(chan bool, 1),
		headerProcCh:              make(chan []*types.Header, 1),
		quitCh:                    make(chan struct{}),
		stateCh:                   make(chan dataPack),
		stateSyncStart:            make(chan *stateSync),
		AssociatedCH:              make(chan bool, 1),
		SyncInBlockPathCh:         make(chan []*types.Block, 1),
		IslandCH:                  make(chan []*types.Block, 10),
		syncStatsState: stateSyncStats{
			processed: rawdb.ReadFastTrieProgress(stateDb),
		},
		trackStateReq: make(chan *stateReq),
	}

	go dl.qosTuner()
	go dl.stateFetcher()
	//go dl.ProcessCommitsAssociated()
	return dl
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of fast synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
func (d *Downloader) Progress() ethereum.SyncProgress {
	// RW the current stats and return the progress
	d.syncStatsLock.RLock()
	defer d.syncStatsLock.RUnlock()

	current := uint64(0)
	switch d.mode {
	case FullSync:
		current = d.blockchain.CurrentBlock().NumberU64()
	case FastSync:
		current = d.blockchain.CurrentFastBlock().NumberU64()
	case LightSync:
		current = d.lightchain.CurrentHeader().Number.Uint64()
	}
	return ethereum.SyncProgress{
		StartingBlock: d.syncStatsChainOrigin,
		CurrentBlock:  current,
		HighestBlock:  d.syncStatsChainHeight,
		PulledStates:  d.syncStatsState.processed,
		KnownStates:   d.syncStatsState.processed + d.syncStatsState.pending,
		CurrentCommit: d.Commitchain.CurrentBlock().NumberU64(), //PDX
		HighestCommit: d.syncStatsCommitHeight,                  //PDX
	}
}

// Synchronising returns whether the downloader is currently retrieving blocks.
func (d *Downloader) Synchronising() bool {
	return atomic.LoadInt32(&d.Synchronis) > 0
}
func (d *Downloader) PeerSet() *peerSet {
	return d.peers
}

func (d *Downloader) CommitSynchronising() bool {
	return atomic.LoadInt32(&d.commitSynchronising) > 0
}

// RegisterPeer injects a new download peer into the set of block source to be
// used for fetching hashes and blocks from.
func (d *Downloader) RegisterPeer(id string, version int, peer Peer, p2pPeer *p2p.Peer) error {
	logger := log.New("peer", id)
	logger.Trace("Registering sync peer")
	if err := d.peers.Register(newPeerConnection(id, version, peer, logger), p2pPeer); err != nil {
		logger.Error("Failed to register sync peer", "err", err)
		return err
	}

	d.qosReduceConfidence()

	return nil
}

// RegisterLightPeer injects a light client peer, wrapping it so it appears as a regular peer.
func (d *Downloader) RegisterLightPeer(id string, version int, peer LightPeer) error {
	return d.RegisterPeer(id, version, &lightPeerWrapper{peer}, nil)
}

// UnregisterPeer remove a peer from the known list, preventing any action from
// the specified peer. An effort is also made to return any pending fetches into
// the queue.
func (d *Downloader) UnregisterPeer(id string) error {
	// Unregister the peer from the active peer set and revoke any fetch tasks
	logger := log.New("peer", id)
	logger.Trace("Unregistering sync peer")
	if err := d.peers.Unregister(id); err != nil {
		logger.Error("Failed to unregister sync peer", "err", err)
		return err
	}
	d.queue.Revoke(id)

	// If this peer was the master peer, abort sync immediately
	d.cancelLock.RLock()
	master := id == d.cancelPeer
	d.cancelLock.RUnlock()

	if master {
		d.cancel()
	}
	return nil
}

// Synchronise tries to sync up our local block chain with a remote peer, both
// adding various sanity checks as well as wrapping it with various log entries.
func (d *Downloader) Synchronise(id string, head common.Hash, td *big.Int, mode SyncMode) error {
	err := d.synchronise(id, head, td, mode)
	switch err {

	case nil:
		//恢复状态
		//atomic.StoreInt32(&d.IsSyncNormalDeleteCommit, 0)
	case errBusy:

	case errTimeout, errBadPeer, errStallingPeer,
		errEmptyHeaderSet, errPeersUnavailable, errTooOld,
		errInvalidAncestor, errInvalidChain:
		//atomic.StoreInt32(&d.IsSyncNormalDeleteCommit, 0)
		log.Warn("Synchronisation failed, dropping peer", "peer", id, "err", err)
		if d.dropPeer == nil {
			// The dropPeer method is nil when `--copydb` is used for a local copy.
			// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
			log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", id)
		} else {
			d.dropPeer(id)
		}
		//d.Worker.SyncJoin(*d.blockchain.CurrentBlock())
		//
	default:
		//atomic.StoreInt32(&d.IsSyncNormalDeleteCommit, 0)
		log.Warn("Synchronisation failed, retrying", "err", err)
	}
	d.Worker.StopWaitSync()
	return err
}

//PDX
//func (d *Downloader) FindCommitAncestorp(id string) (uint64, error) {
//
//	p := d.peers.Peer(id)
//	if p == nil {
//		return 0, errUnknownPeer
//	}
//
//	latest, err := d.fetchCommitHeight(p)
//	if err != nil {
//		return 0, err
//	}
//	height := latest.NumberU64()
//
//	origin, _, err := d.findCommitAncestor(p, height)
//	if err != nil {
//		return 0, err
//	}
//	return origin, nil
//}

func (d *Downloader) SynchroniseCommit(id string, syncs int, assertionSum CommitAssertionSum) (err error) {
	//sync 递归次数
	//if syncs == 0 {
	//	if !atomic.CompareAndSwapInt32(&d.Synchronis, 0, 1) {
	//		return ErrCommitBusy
	//	}
	//
	//}

	d.peers.Reset()

	for empty := false; !empty; {
		select {
		case <-d.commitCh:
		case <-d.processCommitCh:
		default:
			empty = true
		}
	}
	cacheBlock.CacheBlocks.ClearBlockHash() //清空hash
	// Retrieve the origin peer and initiate the downloading process
	p := d.peers.Peer(id)
	if p == nil {
		log.Error("p == nil")
		syncs = -1
		return errUnknownPeer
	}
	wait := &sync.WaitGroup{}
	log.Debug("Synchronising with the network commit", "peer", p.id, "eth", p.version)
	errProcessCh := make(chan error, 10) //processCommits错误接收
	errFetchCh := make(chan error, 10)   //fetchCommits 错误接收
	defer func(start time.Time) {
		log.Info("同步完成", "len", len(errProcessCh))
		wait.Wait()
		if err == nil {
			//接收协程中的错误
			select {
			case err = <-errProcessCh:
				log.Error("收到了error", "err", err)
			default:
				err = nil
			}
		}
		if syncs > 0 && err == nil {
			//递归调用不验证assertion
			err = d.SynchroniseCommit(id, syncs, CommitAssertionSum{big.NewInt(0), nil})
		}

		log.Debug("Synchronisation commit terminated", "elapsed", time.Since(start))
	}(time.Now())

	// Look up the sync boundaries: the common ancestor and the target block
	latest, err := d.fetchCommitHeight(p)
	if err != nil {
		syncs = -1
		return err
	}
	height := latest.NumberU64()
	commitBlock := d.Commitchain.CurrentBlock()
	if height <= commitBlock.NumberU64() && syncs != 0 {
		log.Info("本地commit高度已经达到对方peer节点的高度", "本地高度", d.Commitchain.CurrentBlock().NumberU64(), "peer节点高度", height)
		syncs = -1
		return nil
	}
	origin, verifyCommitBlock, err := d.findCommitAncestor(p, height)
	if err != nil {
		syncs = -1
		return err
	}
	if origin != 0 && verifyCommitBlock.NumberU64() > 1 && syncs == 0 {
		//先验证一个祖先块,避免是作恶节点
		if !d.syncOrigin(origin) {
			return errors.New("verify origin fail")
		}
		err := d.Worker.VerifyCommitBlock(verifyCommitBlock, engine.LocalLand{}, true)
		if err != nil {
			log.Error("verifyAncestorCommitBlock error", "err", err)
			return err
		}
	}

	// Ensure our origin point is below any fast sync pivot point
	pivot := uint64(0)
	if origin >= height && syncs > 0 {
		log.Info("commit高度相同,不需要再次同步commit", "height", height, "origin", origin)
		return nil
	}
	syncs++ //可以再次同步 -1不同步 0第一次正常同步 >0后续同步commit

	wait.Add(1)
	_, commitExtra := types2.CommitExtraDecode(verifyCommitBlock)
	//更新commit同步最新高度
	d.syncStatsLock.Lock()
	d.syncStatsCommitHeight = height
	d.syncStatsLock.Unlock()

	log.Info("最新的高度和祖先", "最新的高度", height, "commitExtra.AssertionSum", commitExtra.AssertionSum, "origin", origin)

	var fetch int32

	go d.processCommits(p, wait, errFetchCh, errProcessCh, originAssertionSum{commitExtra.AssertionSum, big.NewInt(int64(origin))}, assertionSum, height, fetch)
	return d.fetchCommits(p, origin, pivot, errFetchCh, errProcessCh, fetch)
}

//查看祖先是否是在规定的大陆块之内
func (d *Downloader) syncOrigin(origin uint64) bool {
	//获取最新的区块,如果是岛屿,寻找最新的大陆块
	current := d.Commitchain.CurrentBlock()
	_, commitExtra := types2.CommitExtraDecode(current)
	switch commitExtra.Island {
	case true:
		//最新块是岛屿,查询最后的大陆块
		if commitExtra.CNum-syncOrigin <= origin {
			log.Info("岛屿情况祖先区块合法可以进行同步", "祖先", origin, "current", current.NumberU64(), "rang", current.NumberU64()-syncOrigin)
			return true
		}

	case false:
		//最新块是大陆块,查看祖先是否在规定的范围内
		if int(current.NumberU64())-syncOrigin <= int(origin) {
			log.Info(" 大陆情况祖先区块合法可以进行同步", "祖先", origin, "current", current.NumberU64(), "rang", current.NumberU64()-syncOrigin)
			return true
		}

	}
	log.Error("祖先高度不合法,停止同步", "祖先", origin, "current", current.NumberU64(), "rang", current.NumberU64()-syncOrigin)
	return false
}

// synchronise will select the peer and use it for Synchronis. If an empty string is given
// it will use the best peer possible and synchronize if its TD is higher than our own. If any of the
// checks fail an error will be returned. This method is synchronous
func (d *Downloader) synchronise(id string, hash common.Hash, td *big.Int, mode SyncMode) error {
	// Mock out the synchronisation if testing
	if d.synchroniseMock != nil {
		return d.synchroniseMock(id, hash)
	}
	// Make sure only one goroutine is ever allowed past this point at once
	//if !atomic.CompareAndSwapInt32(&d.Synchronis, 0, 1) {
	//	return errBusy
	//}
	//defer atomic.StoreInt32(&d.Synchronis, 0)

	// Post a user notification of the sync (only once per session)
	if atomic.CompareAndSwapInt32(&d.notified, 0, 1) {
		log.Info("Block synchronisation started")
	}
	// Reset the queue, peer set and wake channels to clean any internal leftover state
	d.queue.Reset()
	d.peers.Reset()

	for _, ch := range []chan bool{d.bodyWakeCh, d.receiptWakeCh} {
		select {
		case <-ch:
		default:
		}
	}
	for _, ch := range []chan dataPack{d.headerCh, d.bodyCh, d.receiptCh} {
		for empty := false; !empty; {
			select {
			case <-ch:
			default:
				empty = true
			}
		}
	}
	for empty := false; !empty; {
		select {
		case <-d.headerProcCh:
		default:
			empty = true
		}
	}
	cacheBlock.CacheBlocks.ClearBlockHash() //清空hash
	// Create cancel channel for aborting mid-flight and mark the master peer
	d.cancelLock.Lock()
	d.cancelCh = make(chan struct{})
	d.cancelPeer = id
	d.cancelLock.Unlock()

	defer d.Cancel() // No matter what, we can't leave the cancel channel open

	// Set the requested sync mode, unless it's forbidden
	d.mode = mode

	// Retrieve the origin peer and initiate the downloading process
	p := d.peers.Peer(id)
	if p == nil {
		return errUnknownPeer
	}
	return d.syncWithPeer(p, hash, td)
}

//func (d *Downloader) SyncNormalDeleteCommit(origin *types.Block) bool {
//	//origin 是取得本地的 不是要找对方同步的
//	//originNormal := d.blockchain.GetBlockByNumber(origin)
//	if atomic.LoadInt32(&d.AssociatedSynchronising) == 1 {
//		log.Info("SyncNormalDeleteCommit半生commit正在同步")
//		return false
//	}
//	if atomic.LoadInt32(&IslandDel) == 1 {
//		atomic.StoreInt32(&IslandDel, 0)
//		return false
//	}
//
//	if origin == nil || origin.NumberU64() == 0 {
//		log.Debug("origin == nil || origin.NumberU64() == 0")
//		return false
//	}
//	//直接删除对应的高度
//	originCommitNum := types2.BlockExtraDecode(origin).CNumber.Uint64()
//	currentCommitNum := d.Commitchain.CurrentBlock().NumberU64()
//
//	if originCommitNum == 1 {
//		log.Debug("需要发送信息")
//		return true
//	}
//
//	if commit := d.Commitchain.GetBlockByNum(currentCommitNum + 1); commit != nil {
//		//在insert的时候有可能是没有更新current,所以查找下
//		currentCommitNum = commit.NumberU64()
//	}
//
//	if originCommitNum >= currentCommitNum || currentCommitNum-originCommitNum <= 1 {
//		log.Debug("originCommitNum >= currentCommitNum || currentCommitNum-originCommitNum <= 1 ||originCommitNum==0", "originCommitNum", originCommitNum, "currentCommitNum", currentCommitNum)
//		return false
//	}
//	for i := originCommitNum + 1; i <= currentCommitNum; i++ {
//
//		block := d.Commitchain.GetBlockByNum(i)
//		log.Info("要删除的高度", "祖先commit高度", originCommitNum, "当前要删的高度", i, "要删除到哪里", currentCommitNum)
//		//先删除 在同步
//		if block == nil {
//			continue
//		}
//		//先删除 在同步 里面进行-1 所以外面要+1
//		engine.CommitDeRepetition.Del(block.NumberU64())           //commit删除缓存
//		engine.AssociatedCommitDeRepetition.Del(block.NumberU64()) //commit删除缓存
//		core.DeleteCommitBlock(d.stateDB, block.Hash(), block.NumberU64(), *d.Commitchain.CommitChain())
//		engine.CommitBlockQueued.DelToProcessCommitBlockQueueMap(i) //删除toProcessCommit缓存
//		quorum.CommitHeightToConsensusQuorum.Del(block.NumberU64(), d.stateDB)
//	}
//
//	//更新当前高度
//	selfCnum := originCommitNum
//
//	if selfCnum == 0 {
//		selfCnum = 1
//	}
//	selfCommit := d.Commitchain.GetBlockByNum(selfCnum)
//	core.SetCurrentCommitBlock(d.Commitchain.CommitChain(), selfCommit)
//
//	log.Info("删除完毕", "commit高度", d.Commitchain.CurrentBlock().NumberU64(), "normal高度", d.blockchain.CurrentBlock().NumberU64())
//	return true
//}

// syncWithPeer starts a block synchronization based on the hash chain from the
// specified peer and head hash.
func (d *Downloader) syncWithPeer(p *peerConnection, hash common.Hash, td *big.Int) (err error) {
	//if atomic.LoadInt32(&d.AssociatedSynchronising) == 1 {
	//	log.Info("总体同步SyncNormalDeleteCommit半生commit正在同步")
	//	return
	//}
	//
	//d.Mux.Post(StartEvent{})
	//defer func() {
	//	// reset on error
	//	if err != nil {
	//		d.Mux.Post(FailedEvent{err})
	//	} else {
	//		d.Mux.Post(DoneEvent{})
	//	}
	//}()

	if p.version < 62 {
		return errTooOld
	}
	//wait:= &sync.WaitGroup{}
	//wait.Add(1)
	//go d.ProcessCommitsAssociated(wait)

	log.Debug("Synchronising with the network", "peer", p.id, "eth", p.version, "head", hash, "td", td, "mode", d.mode)
	examineSync.PeerExamineSync.AddPeer(p.id)
	defer func(start time.Time) {
		//log.Info("开始等待")
		//wait.Wait()
		examineSync.PeerExamineSync.DelPeer(p.id) //防止同步的节点被剔除
		log.Debug("Synchronisation terminated", "elapsed", time.Since(start))
	}(time.Now())

	// Look up the sync boundaries: the common ancestor and the target block
	latest, err := d.FetchHeight(p)
	if err != nil {
		return err
	}

	log.Info("寻找到的同步的边界信息", "block", latest.Number.Uint64(), "hash", latest.Hash().String())

	height := latest.Number.Uint64()

	origin, err := d.FindAncestor(p, height)
	if err != nil {
		return err
	}
	log.Info("寻找到共同的祖先", "祖先信息", origin, "p id", p.id)

	err = d.blockchain.SetHead(origin)
	if err != nil {
		return err
	}
	for i := origin + 1; i <= height; i++ {
		//d.Commitchain.DeleteHeightMap(i)
		engine.NormalBlockQueued.DelToProcessNormalBlockQueueMap(i) //删除toProcessNormal
	}
	d.syncStatsLock.Lock()
	if d.syncStatsChainHeight <= origin || d.syncStatsChainOrigin > origin {
		d.syncStatsChainOrigin = origin
	}
	d.syncStatsChainHeight = height
	d.syncStatsLock.Unlock()

	// Ensure our origin point is below any fast sync pivot point
	pivot := uint64(0)
	if d.mode == FastSync {
		if height <= uint64(fsMinFullBlocks) {
			origin = 0
		} else {
			pivot = height - uint64(fsMinFullBlocks)
			if pivot <= origin {
				origin = pivot - 1
			}
		}
	}
	d.committed = 1
	if d.mode == FastSync && pivot != 0 {
		d.committed = 0
	}
	// Initiate the sync using a concurrent header and content retrieval algorithm
	d.queue.Prepare(origin+1, d.mode)
	if d.syncInitHook != nil {
		d.syncInitHook(origin, height)
	}
	log.Info("同步的normal", "最新高度", height, "祖先", origin, "pivot", pivot)

	fetchers := []func() error{
		func() error { return d.fetchHeaders(p, origin+1, pivot) }, // Headers are always retrieved
		func() error { return d.fetchBodies(origin + 1) },          // Bodies are retrieved during normal and fast sync
		func() error { return d.fetchReceipts(origin + 1) },        // Receipts are retrieved during fast sync
		func() error { return d.processHeaders(origin+1, pivot, td) },
	}
	if d.mode == FastSync {
		fetchers = append(fetchers, func() error { return d.processFastSyncContent(latest) })
	} else if d.mode == FullSync {
		fetchers = append(fetchers, d.processFullSyncContent)
	}
	return d.spawnSync(fetchers)
}

// spawnSync runs d.process and all given fetcher functions to completion in
// separate goroutines, returning the first error that appears.
func (d *Downloader) spawnSync(fetchers []func() error) error {
	errc := make(chan error, len(fetchers))
	d.cancelWg.Add(len(fetchers))
	for _, fn := range fetchers {
		fn := fn
		go func() { defer d.cancelWg.Done(); errc <- fn() }()
	}
	// Wait for the first error, then terminate the others.
	var err error
	for i := 0; i < len(fetchers); i++ {
		if i == len(fetchers)-1 {
			// Close the queue when all fetchers have exited.
			// This will cause the block processor to end when
			// it has processed the queue.
			d.queue.Close()
		}
		if err = <-errc; err != nil {
			break
		}
	}
	d.queue.Close()
	d.Cancel()
	return err
}

// cancel aborts all of the operations and resets the queue. However, cancel does
// not wait for the running download goroutines to finish. This method should be
// used when cancelling the downloads from inside the downloader.
func (d *Downloader) cancel() {
	// Close the current cancel channel
	d.cancelLock.Lock()
	if d.cancelCh != nil {
		select {
		case <-d.cancelCh:
			// Channel was already closed
		default:
			close(d.cancelCh)
		}
	}
	d.cancelLock.Unlock()
}

// Cancel aborts all of the operations and waits for all download goroutines to
// finish before returning.
func (d *Downloader) Cancel() {
	d.cancel()
	d.cancelWg.Wait()
}

// Terminate interrupts the downloader, canceling all pending operations.
// The downloader cannot be reused after calling Terminate.
func (d *Downloader) Terminate() {
	// Close the termination channel (make sure double close is allowed)
	d.quitLock.Lock()
	select {
	case <-d.quitCh:
	default:
		close(d.quitCh)
	}
	d.quitLock.Unlock()

	// Cancel any pending download requests
	d.Cancel()
}

// FetchHeight retrieves the head header of the remote peer to aid in estimating
// the total time a pending synchronisation would take.
func (d *Downloader) FetchHeight(p *peerConnection) (*types.Header, error) {
	p.log.Debug("Retrieving remote chain height")

	// Request the advertised remote head block and wait for the response
	//head, _ := p.peer.Head()
	go p.peer.RequestNormalLatest()

	ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
	timeout := time.After(ttl)
	for {
		select {
		case <-d.cancelCh:
			return nil, errCancelBlockFetch

		case packet := <-d.headerCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received headers from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*headerPack).headers
			if len(headers) != 1 {
				p.log.Debug("Multiple headers for single request", "headers", len(headers))
				return nil, errBadPeer
			}
			head := headers[0]
			p.log.Debug("Remote head header identified", "number", head.Number, "hash", head.Hash())
			return head, nil

		case <-timeout:
			p.log.Debug("Waiting for head header timed out", "elapsed", ttl)
			return nil, errTimeout

		case <-d.bodyCh:
		case <-d.receiptCh:
			// Out of bounds delivery, ignore
		}
	}
}

func (d *Downloader) fetchCommitHeight(p *peerConnection) (*types.Block, error) {
	p.log.Debug("Retrieving remote commit chain height")

	// Request the advertised remote head block and wait for the response
	go p.peer.RequestLatest()

	ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
	timeout := time.After(ttl)
	for {
		select {
		case packet := <-d.commitCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received commits from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*commitPack).commits
			if len(headers) != 1 {
				p.log.Debug("Multiple commits for single request", "headers", len(headers))
				return nil, errBadPeer
			}
			head := headers[0]
			p.log.Debug("Remote head commit identified", "number", head.NumberU64(), "hash", head.Hash())
			return head, nil

		case <-timeout:
			p.log.Debug("Waiting fetchCommitHeight for head commit timed out", "elapsed", ttl)
			return nil, errTimeout
		}
	}
}

// FindAncestor tries to locate the common ancestor link of the local chain and
// a remote peers blockchain. In the general case when our node was in sync and
// on the correct chain, checking the top N links should already get us a match.
// In the rare scenario when we ended up on a long reorganisation (i.e. none of
// the head links match), we do a binary search to find the common ancestor.
func (d *Downloader) FindAncestor(p *peerConnection, height uint64) (uint64, error) {
	// Figure out the valid ancestor range to prevent rewrite attacks
	floor, ceil := int64(-1), d.lightchain.CurrentHeader().Number.Uint64()

	if d.mode == FullSync {
		ceil = d.blockchain.CurrentBlock().NumberU64()
	} else if d.mode == FastSync {
		ceil = d.blockchain.CurrentFastBlock().NumberU64()
	}
	if ceil >= MaxForkAncestry {
		floor = int64(ceil - MaxForkAncestry)
	}
	p.log.Debug("Looking for common ancestor", "local", ceil, "remote", height)

	// Request the topmost blocks to short circuit binary ancestor lookup
	head := ceil
	if head > height {
		head = height
	}
	from := int64(head) - int64(MaxMaxHeaderFetchPDX)
	if from < 0 {
		from = 0
	}
	// Span out with 15 block gaps into the future to catch bad head reports
	limit := 2 * MaxHeaderFetch / 16
	count := 1 + int((int64(ceil)-from)/16)
	if count > limit {
		count = limit
	}
	log.Info("FindAncestor---", "from", uint64(from), "count", count)
	//PDX
	go p.peer.RequestHeadersByNumber(uint64(from), count, int(engine.Cnfw.Uint64()), false)

	// Wait for the remote response to the head fetch
	number, hash := uint64(0), common.Hash{}

	ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
	timeout := time.After(ttl)

	for finished := false; !finished; {
		select {
		case <-d.cancelCh:
			return 0, errCancelHeaderFetch

		case packet := <-d.headerCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received headers from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*headerPack).headers
			if len(headers) == 0 {
				p.log.Warn("Empty head header set")
				return 0, errEmptyHeaderSet
			}
			// Make sure the peer's reply conforms to the request
			for i := 0; i < len(headers); i++ {
				if number := headers[i].Number.Int64(); number != from+int64(i)*int64(engine.Cnfw.Uint64()+1) { //PDX
					p.log.Warn("Head headers broke chain ordering", "index", i, "requested", from+int64(i)*16, "received", number)
					return 0, errInvalidChain
				}
			}
			// Check if a common ancestor was found
			finished = true
			for i := len(headers) - 1; i >= 0; i-- {
				// Skip any headers that underflow/overflow our requested set
				if headers[i].Number.Int64() < from || headers[i].Number.Uint64() > ceil {
					continue
				}
				// Otherwise check if we already know the header or not
				log.Info("headers里面的高度", "高度", headers[i].Number.Int64(), "hash", headers[i].Hash())
				if (d.mode == FullSync && d.blockchain.HasBlock(headers[i].Hash(), headers[i].Number.Uint64())) || (d.mode != FullSync && d.lightchain.HasHeader(headers[i].Hash(), headers[i].Number.Uint64())) {
					number, hash = headers[i].Number.Uint64(), headers[i].Hash()

					// If every header is known, even future ones, the peer straight out lied about its head
					if number > height && i == limit-1 {
						p.log.Warn("Lied about chain head", "reported", height, "found", number)
						return 0, errStallingPeer
					}
					break
				}
			}

		case <-timeout:
			p.log.Debug("Waiting for head header timed out", "elapsed", ttl)
			return 0, errTimeout

		case <-d.bodyCh:
		case <-d.receiptCh:
			// Out of bounds delivery, ignore
		}
	}
	// If the head fetch already found an ancestor, return
	if hash != (common.Hash{}) {
		if int64(number) <= floor {
			p.log.Warn("Ancestor below allowance", "number", number, "hash", hash, "allowance", floor)
			return 0, errInvalidAncestor
		}
		p.log.Debug("Found common ancestor", "number", number, "hash", hash)
		return number, nil
	}
	// Ancestor not found, we need to binary search over our chain
	start, end := uint64(0), head
	if floor > 0 {
		start = uint64(floor)
	}
	for start+1 < end {
		// Split our chain interval in two, and request the hash to cross check
		check := (start + end) / 2

		ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
		timeout := time.After(ttl)

		go p.peer.RequestHeadersByNumber(check, 1, 0, false)

		// Wait until a reply arrives to this request
		for arrived := false; !arrived; {
			select {
			case <-d.cancelCh:
				return 0, errCancelHeaderFetch

			case packer := <-d.headerCh:
				// Discard anything not from the origin peer
				if packer.PeerId() != p.id {
					log.Debug("Received headers from incorrect peer", "peer", packer.PeerId())
					break
				}
				// Make sure the peer actually gave something valid
				headers := packer.(*headerPack).headers
				if len(headers) != 1 {
					p.log.Debug("Multiple headers for single request", "headers", len(headers))
					return 0, errBadPeer
				}
				arrived = true

				// Modify the search interval based on the response
				if (d.mode == FullSync && !d.blockchain.HasBlock(headers[0].Hash(), headers[0].Number.Uint64())) || (d.mode != FullSync && !d.lightchain.HasHeader(headers[0].Hash(), headers[0].Number.Uint64())) {
					end = check
					break
				}
				header := d.lightchain.GetHeaderByHash(headers[0].Hash()) // Independent of sync mode, header surely exists
				if header.Number.Uint64() != check {
					p.log.Debug("Received non requested header", "number", header.Number, "hash", header.Hash(), "request", check)
					return 0, errBadPeer
				}
				start = check

			case <-timeout:
				p.log.Debug("Waiting for search header timed out", "elapsed", ttl)
				return 0, errTimeout

			case <-d.bodyCh:
			case <-d.receiptCh:
				// Out of bounds delivery, ignore
			}
		}
	}
	// Ensure valid ancestry and return
	if int64(start) <= floor {
		p.log.Warn("Ancestor below allowance", "number", start, "hash", hash, "allowance", floor)
		return 0, errInvalidAncestor
	}
	p.log.Debug("Found common ancestor", "number", start, "hash", hash)
	return start, nil
}

func (d *Downloader) findCommitAncestor(p *peerConnection, height uint64) (uint64, *types.Block, error) {
	// Figure out the valid ancestor range to prevent rewrite attacks
	floor, ceil := int64(-1), d.Commitchain.CurrentBlock().NumberU64()

	if ceil >= MaxForkAncestry {
		floor = int64(ceil - MaxForkAncestry)
	}
	p.log.Debug("Looking for commit common ancestor", "local", ceil, "remote", height)

	// Request the topmost blocks to short circuit binary ancestor lookup
	head := ceil
	if head > height {
		head = height
	}
	from := int64(head) - int64(MaxMaxHeaderFetchPDX)
	if from < 0 {
		from = 0
	}
	// Span out with 15 block gaps into the future to catch bad head reports
	limit := 2 * MaxHeaderFetch / 16
	count := 1 + int((int64(ceil)-from)/16)
	if count > limit {
		count = limit
	}
	log.Info("获取新消息", "count", count, "from", from, "head", head)
	go p.peer.RequestCommitsByNumber(uint64(from), count, 15, false)

	// Wait for the remote response to the head fetch
	number, hash, commitBlock := uint64(0), common.Hash{}, &types.Block{}

	ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
	timeout := time.After(ttl)

	for finished := false; !finished; {
		select {
		case packet := <-d.commitCh:
			// Discard anything not from the origin peer
			if packet.PeerId() != p.id {
				log.Debug("Received commits from incorrect peer", "peer", packet.PeerId())
				break
			}
			// Make sure the peer actually gave something valid
			headers := packet.(*commitPack).commits
			if len(headers) == 0 {
				p.log.Warn("Empty head commit set")
				return 0, nil, errEmptyHeaderSet
			}
			for _, head := range headers {
				log.Info("获取的commit块", "commit", head.Number())

			}
			// Make sure the peer's reply conforms to the request
			for i := 0; i < len(headers); i++ {
				if number := headers[i].Number().Int64(); number != from+int64(i)*16 {
					p.log.Warn("Head commits broke chain ordering", "index", i, "requested", from+int64(i)*16, "received", number)
					return 0, nil, errInvalidChain
				}
			}
			// Check if a common ancestor was found
			finished = true
			for i := len(headers) - 1; i >= 0; i-- {
				// Skip any headers that underflow/overflow our requested set
				if headers[i].Number().Int64() < from || headers[i].NumberU64() > ceil {
					continue
				}
				// Otherwise check if we already know the header or not
				if d.Commitchain.HasBlock(headers[i].Hash(), headers[i].NumberU64()) {
					number, hash, commitBlock = headers[i].NumberU64(), headers[i].Hash(), headers[i]

					// If every header is known, even future ones, the peer straight out lied about its head
					if number > height && i == limit-1 {
						p.log.Warn("Lied about chain head commit", "reported", height, "found", number)
						return 0, nil, errStallingPeer
					}
					break
				}
			}

		case <-timeout:
			p.log.Debug("Waiting findCommitAncestor for head commit timed out", "elapsed", ttl)
			return 0, nil, errTimeout
		}
	}
	// If the head fetch already found an ancestor, return
	if hash != (common.Hash{}) {
		if int64(number) <= floor {
			p.log.Warn("Commit ancestor below allowance", "number", number, "hash", hash, "allowance", floor)
			return 0, nil, errInvalidAncestor
		}
		p.log.Debug("Found common commit ancestor", "number", number, "hash", hash, "commitBlockHeight", commitBlock.NumberU64())
		return number, commitBlock, nil
	}
	// Ancestor not found, we need to binary search over our chain
	start, end := uint64(0), head
	if floor > 0 {
		start = uint64(floor)
	}
	for start+1 < end {
		// Split our chain interval in two, and request the hash to cross check
		check := (start + end) / 2

		ttl := time.Duration(engine.BlockDelay) * time.Millisecond * 4
		timeout := time.After(ttl)

		go p.peer.RequestCommitsByNumber(check, 1, 0, false)

		// Wait until a reply arrives to this request
		for arrived := false; !arrived; {
			select {
			case packer := <-d.commitCh:
				// Discard anything not from the origin peer
				if packer.PeerId() != p.id {
					log.Debug("Received commits from incorrect peer", "peer", packer.PeerId())
					break
				}
				// Make sure the peer actually gave something valid
				headers := packer.(*commitPack).commits
				if len(headers) != 1 {
					p.log.Debug("Multiple commits for single request", "headers", len(headers))
					return 0, nil, errBadPeer
				}
				arrived = true

				// Modify the search interval based on the response
				if !d.Commitchain.HasBlock(headers[0].Hash(), headers[0].NumberU64()) {
					end = check
					break
				}
				start = check
				commitBlock = headers[0]

			case <-timeout:
				p.log.Debug("Waiting for search commit timed out", "elapsed", ttl)
				return 0, nil, errTimeout
			}
		}
	}
	if start == 0 {
		commitBlock = d.Commitchain.GetBlockByNum(0)
	}
	// Ensure valid ancestry and return
	if int64(start) <= floor {
		p.log.Warn("commit ancestor below allowance", "number", start, "hash", hash, "allowance", floor)
		return 0, nil, errInvalidAncestor
	}
	//p.log.Debug("Found common commit ancestor", "number", start, "hash", hash, "commitBlockHeight", commitBlock.NumberU64())
	return start, commitBlock, nil
}

// fetchHeaders keeps retrieving headers concurrently from the number
// requested, until no more are returned, potentially throttling on the way. To
// facilitate concurrency but still protect against malicious nodes sending bad
// headers, we construct a header chain skeleton using the "origin" peer we are
// syncing with, and fill in the missing headers using anyone else. Headers from
// other peers are only accepted if they map cleanly to the skeleton. If no one
// can fill in the skeleton - not even the origin peer - it's assumed invalid and
// the origin is dropped.
func (d *Downloader) fetchHeaders(p *peerConnection, from uint64, pivot uint64) error {
	p.log.Debug("Directing header downloads", "origin", from)
	defer p.log.Debug("Header download terminated")

	// Create a timeout timer, and the associated header fetcher
	skeleton := false           // Skeleton assembly phase or finishing up //PDX true->false
	request := time.Now()       // time of the last skeleton fetch request
	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
	<-timeout.C                 // timeout channel should be initially empty
	defer timeout.Stop()

	var ttl time.Duration
	getHeaders := func(from uint64) {
		request = time.Now()

		ttl = d.requestTTL()
		timeout.Reset(ttl)

		if skeleton {
			p.log.Info("Fetching skeleton headers", "count", MaxHeaderFetch, "from", from)
			go p.peer.RequestHeadersByNumber(from+uint64(MaxHeaderFetch)-1, MaxSkeletonSize, MaxHeaderFetch-1, false)
		} else {
			p.log.Info("Fetching full headers", "count", MaxHeaderFetch, "from", from)
			go p.peer.RequestHeadersByNumber(from, MaxHeaderFetch, 0, false)
		}
	}
	// Start pulling the header chain skeleton until all is done
	getHeaders(from)

	for {
		select {
		case <-d.cancelCh:
			return errCancelHeaderFetch

		case packet := <-d.headerCh:
			// Make sure the active peer is giving us the skeleton headers
			if packet.PeerId() != p.id {
				log.Debug("Received skeleton from incorrect peer", "peer", packet.PeerId())
				break
			}
			headerReqTimer.UpdateSince(request)
			timeout.Stop()

			// If the skeleton's finished, pull any remaining head headers directly from the origin
			if packet.Items() == 0 && skeleton {
				skeleton = false
				getHeaders(from)
				continue
			}
			// If no more headers are inbound, notify the content fetchers and return
			if packet.Items() == 0 {
				// Don't abort header fetches while the pivot is downloading
				if atomic.LoadInt32(&d.committed) == 0 && pivot <= from {
					p.log.Debug("No headers, waiting for pivot commit")
					select {
					case <-time.After(fsHeaderContCheck):
						getHeaders(from)
						continue
					case <-d.cancelCh:
						return errCancelHeaderFetch
					}
				}
				// Pivot done (or not in fast sync) and no more headers, terminate the process
				p.log.Debug("No more headers available")
				select {
				case d.headerProcCh <- nil:
					return nil
				case <-d.cancelCh:
					return errCancelHeaderFetch
				}
			}
			headers := packet.(*headerPack).headers

			// If we received a skeleton batch, resolve internals concurrently
			if skeleton {
				filled, proced, err := d.fillHeaderSkeleton(from, headers)
				if err != nil {
					p.log.Debug("Skeleton chain invalid", "err", err)
					return errInvalidChain
				}
				headers = filled[proced:]
				from += uint64(proced)
			}
			// Insert all the new headers and fetch the next batch
			if len(headers) > 0 {
				p.log.Info("Scheduling new headers", "count", len(headers), "from", from)
				select {
				case d.headerProcCh <- headers:
				case <-d.cancelCh:
					return errCancelHeaderFetch
				}
				from += uint64(len(headers))
			}
			getHeaders(from)

		case <-timeout.C:
			if d.dropPeer == nil {
				// The dropPeer method is nil when `--copydb` is used for a local copy.
				// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
				p.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", p.id)
				break
			}
			// Header retrieval timed out, consider the peer bad and drop
			p.log.Debug("Header request timed out1", "elapsed", ttl)
			headerTimeoutMeter.Mark(1)
			d.dropPeer(p.id)

			// Finish the sync gracefully instead of dumping the gathered data though
			for _, ch := range []chan bool{d.bodyWakeCh, d.receiptWakeCh} {
				select {
				case ch <- false:
				case <-d.cancelCh:
				}
			}
			select {
			case d.headerProcCh <- nil:
			case <-d.cancelCh:
			}
			return errBadPeer
		}
	}
}

//PDX
//func (d *Downloader) FetchIslandCommits(from, height uint64, id string) (commitBlocks []*types.Block, err error) {
//	p := d.peers.Peer(id)
//	p.log.Debug("FetchIslandCommits Directing commit downloads", "origin", from)
//	defer p.log.Debug("FetchIslandCommits commit download terminated")
//
//	// Create a timeout timer, and the associated header fetcher
//	request := time.Now()       // time of the last skeleton fetch request
//	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
//	<-timeout.C                 // timeout channel should be initially empty
//	defer timeout.Stop()
//	var ttl time.Duration
//	getHeaders := func(from uint64) {
//		request = time.Now()
//
//		ttl = time.Duration(engine.BlockDelay) * time.Millisecond * 4
//
//		p.log.Trace("Fetching full commits", "count", MaxHeaderFetch, "from", from, "peer", p.id)
//		go p.peer.RequestCommitsByNumber(from, MaxHeaderFetch, 0, false)
//
//	}
//	// Start pulling the header chain skeleton until all is done
//	getHeaders(from)
//fetch:
//	for {
//		select {
//		case packet := <-d.commitCh:
//			// Make sure the active peer is giving us the skeleton headers
//			if packet.PeerId() != p.id {
//				log.Debug("FetchIslandCommits Received skeleton from incorrect peer", "peer", packet.PeerId())
//				return nil, errors.New("FetchIslandCommits Received skeleton from incorrect peer")
//			}
//			headerReqTimer.UpdateSince(request)
//			timeout.Stop()
//
//			// If no more headers are inbound, notify the content fetchers and return
//			if packet.Items() == 0 {
//				p.log.Debug("FetchIslandCommits No more commits available")
//				return nil, errors.New("FetchIslandCommits No more commits available")
//			}
//			headers := packet.(*commitPack).commits
//
//			// Insert all the new headers and fetch the next batch
//			if len(headers) > 0 {
//				p.log.Trace("FetchIslandCommits Scheduling new commits", "count", len(headers), "from", from)
//				commitBlocks = append(commitBlocks, headers...)
//				from += uint64(len(headers))
//			}
//			if from >= height {
//				log.Info("拿齐了所有的commit块", "from", from, "height", height, "最新高度", commitBlocks[len(commitBlocks)-1:][0].NumberU64())
//				break fetch
//			}
//			getHeaders(from)
//
//		case <-timeout.C:
//			if d.dropPeer == nil {
//				// The dropPeer method is nil when `--copydb` is used for a local copy.
//				// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
//				p.log.Warn("FetchIslandCommits Downloader wants to drop peer, but peerdrop-function is not set", "peer", p.id)
//				return nil, errors.New("FetchIslandCommits Downloader wants to drop peer, but peerdrop-function is not set")
//			}
//			// Header retrieval timed out, consider the peer bad and drop
//			p.log.Debug("FetchIslandCommits commit request timed out", "elapsed", ttl)
//			headerTimeoutMeter.Mark(1)
//			d.dropPeer(p.id)
//
//			// Finish the sync gracefully instead of dumping the gathered data though
//			return nil, errBadPeer
//		}
//	}
//	return commitBlocks, nil
//}

func (d *Downloader) fetchCommits(p *peerConnection, from uint64, pivot uint64, errFetchCh, errCh chan error, fetch int32) error {
	p.log.Debug("Directing commit downloads", "origin", from)
	defer func() {
		p.log.Debug("commit download terminated")
		atomic.StoreInt32(&fetch, 1)
	}()
	if from == 0 {
		//防止第一次同步的时候从0开始取
		from = 1
	}
	// Create a timeout timer, and the associated header fetcher
	request := time.Now()       // time of the last skeleton fetch request
	timeout := time.NewTimer(0) // timer to dump a non-responsive active peer
	<-timeout.C                 // timeout channel should be initially empty
	defer timeout.Stop()

	var ttl time.Duration
	getHeaders := func(from uint64) {
		request = time.Now()

		//ttl = time.Duration(engine.BlockDelay) * time.Millisecond * 4
		ttl = d.requestTTL()
		timeout.Reset(ttl)

		p.log.Info("Fetching full commits", "count", MaxHeaderFetch, "from", from, "peer", p.id, "ttl", ttl)
		go p.peer.RequestCommitsByNumber(from, MaxHeaderFetch, 0, false)

	}
	// Start pulling the header chain skeleton until all is done
	getHeaders(from)
	log.Info("查看通道数量", "len", len(d.commitCh))
	for {
		select {
		case packet := <-d.commitCh:
			// Make sure the active peer is giving us the skeleton headers
			if packet.PeerId() != p.id {
				//如果不是同一个ip来的信息,直接丢掉
				log.Error("Received skeleton from incorrect peer", "peer", packet.PeerId())
				continue
			}
			headerReqTimer.UpdateSince(request)
			timeout.Stop()
			log.Info("收到了 commit")
			// If no more headers are inbound, notify the content fetchers and return
			if packet.Items() == 0 {
				p.log.Debug("No more commits available")
				return nil
			}
			headers := packet.(*commitPack).commits

			// Insert all the new headers and fetch the next batch
			if len(headers) > 0 {
				p.log.Info("Scheduling new commits", "count", len(headers), "from", from)
				d.processCommitCh <- headers
				from += uint64(len(headers))
			}
			getHeaders(from)
		case err := <-errCh: //processCommits 出错,这里接收
			log.Error("接收到了error", "err", err)
			return err

		case <-timeout.C:
			if d.dropPeer == nil {
				// The dropPeer method is nil when `--copydb` is used for a local copy.
				// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
				p.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", p.id)
				return nil
			}
			errFetchCh <- errors.New(fmt.Sprintf("fetch timeout,ttl=%v", ttl)) //通知process退出
			// Header retrieval timed out, consider the peer bad and drop
			p.log.Debug("commit request timed out", "elapsed", ttl)
			headerTimeoutMeter.Mark(1)
			d.dropPeer(p.id)

			// Finish the sync gracefully instead of dumping the gathered data though
			return errBadPeer
		}
	}
}

//PDX
func (d *Downloader) processCommits(p *peerConnection, wait *sync.WaitGroup, errFetchCh, errCh chan error, origin originAssertionSum, assertionSum CommitAssertionSum,
	height uint64, fetch int32) error {
	timeOut := time.NewTicker(5 * time.Minute)
	defer func() {
		log.Debug("processCommits  terminated")
		timeOut.Stop()
		wait.Done()

	}()

	//timeout := time.NewTimer(100 * time.Second)
	first := true //只有第一次才进行才走
	for {
		log.Info("processCommits1")
		currentCommitBlock := d.Commitchain.CurrentBlock()
		_, commitExtra := types2.CommitExtraDecode(currentCommitBlock)
		if !first && len(d.processCommitCh) == 0 && height <= currentCommitBlock.NumberU64() {
			log.Info("CommitCh为空")
			break
		}
		//fetch方法退出,并且通道里的commit块已经处理完毕,直接退出方法
		if len(d.processCommitCh) == 0 && atomic.LoadInt32(&fetch) == 1 {
			log.Info("获取commit完成,CommitCh为空")
			break
		}
		//第一次不更新祖先
		if !first {
			origin.originNum = currentCommitBlock.Number()
			origin.originAssertion = commitExtra.AssertionSum
		}
		blocks := make([]*types.Block, 0)
		log.Info("processCommits2", "len(d.processCommitCh)", len(d.processCommitCh), "height", height,
			"currentCommitBlock.NumberU64()", currentCommitBlock.NumberU64())

		select {
		case blocksCh := <-d.processCommitCh:
			// handle commit blocks
			blocks = append(blocks, blocksCh...)
			log.Debug("downloader collect commit blocks", "len", len(blocks), "祖先", origin.originNum)
			//验证assertion数量,和验证委员会的数
			if assertionSum.CommitHeight != nil {
				err := d.verifySyncCommit(blocks, origin, assertionSum)
				if err != nil {
					log.Error("verifySyncCommit fail", "err", err)
					errCh <- err
					return err
				}
				if first {
					d.CommitSyncIsland(origin.originNum.Uint64()) //第一次后面不删除删除本地commit
				}
			}
			first = false

			err := d.processSyncCommit(p, blocks, errCh) //保存同步过来的commit
			if err != nil {
				log.Error("processSyncCommit fail", "err", err)
				errCh <- err
				return err
			}
		case <-timeOut.C:
			log.Warn("processSyncCommit timeout 退出")
			return nil

		case err := <-errFetchCh:
			log.Error("FetchCommit fail 退出 processCommits")
			return err

		}
	}
	return nil

}

func (d *Downloader) verifySyncCommit(blocks []*types.Block, origin originAssertionSum, assertionSum CommitAssertionSum) error {
	sum := origin.originAssertion //祖先块的assertion数量
	log.Info("祖先块的assertion数量", "数量", sum, "高度", origin.originNum)
	var consensusAddress map[common.Address]struct{}
	if consensusQuorum, ok := quorum.CommitHeightToConsensusQuorum.Get(origin.originNum.Uint64(), d.stateDB); ok {
		//取出当前委员会
		consensusAddress = consensusQuorum.CopyAddress()
	} else {
		consensusAddress = make(map[common.Address]struct{})
	}

	//取下一次更新的高度,由于是从中间开始,所以需要找到下次更新的时间
	commitExtra, _ := types2.CommitExtraDecode(blocks[0])
	log.Info("commitExtra.HistoryHeight", "commitExtra.HistoryHeight", commitExtra.HistoryHeight)

	if testQuorum, ok := quorum.UpdateQuorumSnapshots.GetUpdateQuorum(commitExtra.HistoryHeight, d.stateDB); !ok {
		quorum.VerifyUpdateQuorums = quorum.NewUpdateQuorum()
	} else {
		quorum.VerifyUpdateQuorums.CopyUpdateQuorum(testQuorum)
	}

	log.Info("quorum.VerifyUpdateQuorums", "高度", quorum.VerifyUpdateQuorums.NextUpdateHeight)
	for _, commitBlock := range blocks {
		_, commitExtra := types2.CommitExtraDecode(commitBlock)
		log.Info("开始验证commitBlock的assertion", "高度", commitBlock.NumberU64(), "hash", commitBlock.Hash(), "Evidences长度", len(commitExtra.Evidences), "island", commitExtra.Island,
			"委员会数量", len(consensusAddress))

		singleCommitAssertion := 0 //每个commit区块有效的assertion累计
		acceptedBlocksLen := len(commitExtra.AcceptedBlocks)
		acceptedBlocks := commitExtra.AcceptedBlocks
		for _, condensedEvidence := range commitExtra.Evidences {
			total := make([]common.Hash, 0)
			ExtraBlocksLen := len(condensedEvidence.ExtraBlocks)
			switch {
			case condensedEvidence.ExtraKind == types2.EVIDENCE_ADD_EXTRA && ExtraBlocksLen != 0:
				//恢复每个节点的path
				//total = append(acceptedBlocks[:acceptedBlocksLen-ExtraBlocksLen],condensedEvidence.ExtraBlocks...)
				total = append(total, acceptedBlocks[:acceptedBlocksLen-ExtraBlocksLen]...)
				total = append(total, condensedEvidence.ExtraBlocks...)
				//log.Debug("ADD出现了path不一致的问题", "恢复的path", total)
			case condensedEvidence.ExtraKind == types2.EVIDENCE_DEL_EXTRA:
				//多余的区块Hash+标准的BlockPath
				total = append(total, acceptedBlocks...)
				total = append(total, condensedEvidence.ExtraBlocks...)
				//log.Debug("DEL出现了path不一致的问题", "恢复的path", total)
			case ExtraBlocksLen == 0 && condensedEvidence.ExtraKind != types2.EVIDENCE_EMP_EXTRA:
				//path完全一致
				total = append(total, acceptedBlocks...)
			//log.Debug("没有出现path不一致的问题", "恢复的path", total)
			default:
				//log.Info("非委员会成员,不需要添加", "add", condensedEvidence.Address())
			}
			total = append(total, condensedEvidence.ParentCommitHash)
			//验证每一个节点的签名
			add := condensedEvidence.Address()
			if !utopia.Perf {
				var err error
				add, _, err = engine.VerifySignAssertBlock(total, condensedEvidence.Signature)
				if err != nil {
					return errors.New("VerifySignAssertBlock fail")
				}

				//log.Info("验证commit完成", "高度", commitBlock.NumberU64(), "sum", sum)
			} else {
				log.Trace("perf mode no verify assert block path")
			}

			//验证单个区块的assertion是否大于委员会的2/3
			if _, ok := consensusAddress[add]; ok {
				singleCommitAssertion++
				if origin.originNum.Cmp(commitBlock.Number()) != 0 {
					sum.Add(sum, big.NewInt(1))
				}
			}
		}
		//更新委员会

		needNode := int(math.Ceil(float64(len(consensusAddress)) * 2 / 3))
		updateQuorum(commitBlock, commitExtra, consensusAddress)
		log.Info("consensusAddress", "长度", len(consensusAddress))
		//不是岛屿,assertion数量应该大于2/3
		if !commitExtra.Island && commitBlock.NumberU64() != 1 {
			switch singleCommitAssertion >= needNode {
			case true:
				log.Info("Sync commit verify succeed", "height", commitBlock.Number(), "needNode", needNode)

			default:
				log.Error("Sync commit verify fail", "height", commitBlock.Number(), "singleCommitAssertion", singleCommitAssertion,
					"needNode", needNode)
				return errors.New("sync commit verify fail")
			}

			//验证单个块的assertion数量
			switch {
			case sum.Cmp(commitExtra.AssertionSum) == 0:
				log.Info("aasertion验证成功", "sum", sum, "commitExtra.AssertionSum", commitExtra.AssertionSum)

			default:
				log.Error("aasertion验证失败", "sum", sum, "commitExtra.AssertionSum", commitExtra.AssertionSum)
				return errors.New("aasertion验证失败")
			}
		}
		//验证收到的commit块的真实性
		if assertionSum.CommitHeight.Cmp(big.NewInt(0)) != 0 {
			if commitBlock.Number().Cmp(assertionSum.CommitHeight) == 0 {
				switch {
				case sum.Cmp(assertionSum.AssertionSum) == 0:
					log.Info("收到的commitAssertion验证成功", "sum", sum, "assertionSum.AssertionSum", assertionSum.AssertionSum)
				default:
					log.Error("收到的commitAasertion验证失败", "sum", sum, "assertionSum.AssertionSum", assertionSum.AssertionSum)
					return errors.New("收到的commitAasertion验证失败")
				}
			}

		}

	}

	return nil

}

func updateQuorum(commitBlock *types.Block, commitExtra types2.CommitExtra, consensusAddress map[common.Address]struct{}) {
	currentNum := commitBlock.NumberU64() - 1
	commitBlockHash := commitBlock.Hash()
	switch {
	case currentNum <= quorum.LessCommit || engine.PerQuorum:
		//正常更新委员会 配置了engine.PerQuorum 每次都更新
		for _, add := range commitExtra.MinerDeletions {
			delete(consensusAddress, add)
		}
		for _, add := range commitExtra.MinerAdditions {
			consensusAddress[add] = struct{}{}
		}
		if currentNum == quorum.LessCommit && !engine.PerQuorum {
			//下次委员会更新高度 配置了engine.PerQuorum 不进行下次委员会的计算
			quorum.VerifyUpdateQuorums.CalculateNextUpdateHeight(commitBlockHash)
		}
	default:
		////增加预备委员会列表
		//quorum.VerifyPrepareQuorums.SetPrepareQuorum(commitBlockHash, commitExtra.MinerAdditions, nil)
		////清理预备委员会列表
		//quorum.VerifyPrepareQuorums.DelPrepareQuorum(commitBlockHash, commitExtra.MinerDeletions, nil)

		//清理当前委员会
		for _, add := range commitExtra.MinerDeletions {
			delete(consensusAddress, add)
		}
		if commitBlock.Number().Cmp(quorum.VerifyUpdateQuorums.NextUpdateHeight) == 0 {
			//计算下次委员会更新时间
			quorum.VerifyUpdateQuorums.CalculateNextUpdateHeight(commitBlockHash)
			log.Info("Verify计算下次委员会更新时间", "本次高度", commitBlock.NumberU64(), "下次更新高度", quorum.VerifyUpdateQuorums.NextUpdateHeight)
			//计算本次要加的委员会的成员
			set := quorum.SortPrepareQuorum(commitExtra.MinerAdditions, commitBlockHash, nil)
			//更新委员会
			for _, add := range set {
				consensusAddress[add] = struct{}{}
			}

		}

	}

}

func (d *Downloader) processSyncCommit(p *peerConnection, blocks []*types.Block, errCh chan error) error {
	for _, block := range blocks {
		log.Info("开始处理同步commit块", "高度", block.NumberU64())
		//time.Sleep(1 * time.Second)
		err := d.Worker.ProcessCommitBlock(block, true)
		switch err {
		case nil:
			log.Info("commit保存完毕", "commit高度", block.NumberU64())

		case engine.CommitHeightToLow:
			err = nil
			continue

		default:
			//收集错误

			log.Error("downloader processCommits error", "err", err)
			return err

		}

		_, commitExtra := types2.CommitExtraDecode(block)
		//根据blockPath来拿取
		//d.syncInBlockPathFetchNormal(commitExtra,p)
		//找到这个commit块对应的第一normal的高度 用当前commit中最新的normal高度减去blockPath的长度
		origin := commitExtra.NewBlockHeight.Uint64() - uint64(len(commitExtra.AcceptedBlocks))
		if len(commitExtra.AcceptedBlocks) == 0 {
			continue
		}

		err = d.syncCommitPathWithPeer(p, origin, commitExtra.AcceptedBlocks, commitExtra.NewBlockHeight.Uint64())
		if err != nil {
			log.Error("syncCommitPathWithPeer fail ", "err", err)
		}
		currentNum := d.blockchain.CurrentBlock().NumberU64()
		log.Info("检查normal高度是否到达", "currentNum", currentNum, "commitExtra.NewBlockHeight.Uint64()", commitExtra.NewBlockHeight.Uint64())
		if currentNum != commitExtra.NewBlockHeight.Uint64() {
			err = errors.New("同步normal高度出现异常,回滚normal,同步失败")
		}
		if err != nil {
			//收集错误
			//errCh <- err
			log.Error("downloader syncCommitPathWithPeer error", "err", err)
			d.Worker.RollBackCommit(block)
			return err
		}
	}
	return nil

}

func (d *Downloader) syncInBlockPathFetchNormal(commitExtra types2.CommitExtra, p *peerConnection) {
	log.Info("要发送的path", "path", commitExtra.AcceptedBlocks)
	p.peer.SendBlockPath(commitExtra.AcceptedBlocks)

	select {
	case blocks := <-d.SyncInBlockPathCh:
		log.Info("收到了根据commitPath中来的normal块")
		_, err := d.blockchain.InsertChain(blocks)
		if err != nil {
			log.Error("出错啦", "err", err)
		}

	}

}

//PDX
//func (d *Downloader) ProcessCommitsAssociated(wait *sync.WaitGroup) error {
//	timeOut := time.NewTimer(5 * time.Second)
//	for {
//		select {
//		case blocks := <-d.ProcessCommitChAssociated:
//			log.Debug("downloader processCommits process commit blocks", "len", len(blocks))
//			// handle commit blocks
//		Associated:
//			for {
//				AssociatedCHTimer := time.NewTimer(time.Second * 1)
//				select {
//				case <-d.AssociatedCH:
//					log.Debug("收到了AssociatedCH")
//					break Associated
//				case <-AssociatedCHTimer.C:
//					log.Debug("半生超时")
//					break Associated
//				}
//			}
//
//			if !atomic.CompareAndSwapInt32(&d.AssociatedSynchronising, 0, 1) {
//				return errBusy
//			}
//			for _, block := range blocks {
//				if engine.AssociatedCommitDeRepetition.Add(block.NumberU64(), block.Hash()) {
//					self := d.Commitchain.SearchingBlock(block.Hash(), block.NumberU64())
//					if self != nil {
//						log.Debug("半生commit已经存过", "num", block.NumberU64(), "Hash", block.Hash())
//						continue
//					}
//				} else {
//					log.Debug("半生commit已经在缓存中", "num", block.NumberU64(), "Hash", block.Hash())
//					continue
//				}
//				log.Info("半生Block", "高度", block.NumberU64())
//				err := d.Worker.ProcessCommitBlock(block, true)
//				if err != nil {
//					log.Error("半生Commit保存失败", "err", err)
//				}
//			}
//			atomic.StoreInt32(&d.AssociatedSynchronising, 0)
//			log.Info("ProcessCommitsAssociated半生保存完成", "len(d.ProcessCommitChAssociated)", len(d.ProcessCommitChAssociated))
//		case <-timeOut.C:
//			log.Info("没有要接收的commit块")
//		}
//		if len(d.ProcessCommitChAssociated) == 0 {
//			wait.Done()
//			return nil
//		}
//	}
//}

// fillHeaderSkeleton concurrently retrieves headers from all our available peers
// and maps them to the provided skeleton header chain.
//
// Any partial results from the beginning of the skeleton is (if possible) forwarded
// immediately to the header processor to keep the rest of the pipeline full even
// in the case of header stalls.
//
// The method returns the entire filled skeleton and also the number of headers
// already forwarded for processing.
func (d *Downloader) fillHeaderSkeleton(from uint64, skeleton []*types.Header) ([]*types.Header, int, error) {
	d.queue.ScheduleSkeleton(from, skeleton)

	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*headerPack)
			return d.queue.DeliverHeaders(pack.peerID, pack.headers, d.headerProcCh)
		}
		expire   = func() map[string]int { return d.queue.ExpireHeaders(d.requestTTL()) }
		throttle = func() bool { return false }
		reserve  = func(p *peerConnection, count int) (*fetchRequest, bool, error) {
			return d.queue.ReserveHeaders(p, count), false, nil
		}
		fetch    = func(p *peerConnection, req *fetchRequest) error { return p.FetchHeaders(req.From, MaxHeaderFetch) }
		capacity = func(p *peerConnection) int { return p.HeaderCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetHeadersIdle(accepted) }
	)
	err := d.fetchParts(errCancelHeaderFetch, d.headerCh, deliver, d.queue.headerContCh, expire,
		d.queue.PendingHeaders, d.queue.InFlightHeaders, throttle, reserve,
		nil, fetch, d.queue.CancelHeaders, capacity, d.peers.HeaderIdlePeers, setIdle, "headers")

	log.Debug("Skeleton fill terminated", "err", err)

	filled, proced := d.queue.RetrieveHeaders()
	return filled, proced, err
}

// fetchBodies iteratively downloads the scheduled block bodies, taking any
// available peers, reserving a chunk of blocks for each, waiting for delivery
// and also periodically checking for timeouts.
func (d *Downloader) fetchBodies(from uint64) error {
	log.Debug("Downloading block bodies", "origin", from)

	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*bodyPack)
			return d.queue.DeliverBodies(pack.peerID, pack.transactions, pack.uncles)
		}
		expire = func() map[string]int { return d.queue.ExpireBodies(d.requestTTL()) }
		fetch  = func(p *peerConnection, req *fetchRequest) error {
			log.Info("去找那个peer获取bodie", "id", p.id)
			return p.FetchBodies(req)
		}
		capacity = func(p *peerConnection) int { return p.BlockCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetBodiesIdle(accepted) }
	)
	err := d.fetchParts(errCancelBodyFetch, d.bodyCh, deliver, d.bodyWakeCh, expire,
		d.queue.PendingBlocks, d.queue.InFlightBlocks, d.queue.ShouldThrottleBlocks, d.queue.ReserveBodies,
		d.bodyFetchHook, fetch, d.queue.CancelBodies, capacity, d.peers.BodyIdlePeers, setIdle, "bodies")

	log.Debug("Block body download terminated", "err", err)
	return err
}

// fetchReceipts iteratively downloads the scheduled block receipts, taking any
// available peers, reserving a chunk of receipts for each, waiting for delivery
// and also periodically checking for timeouts.
func (d *Downloader) fetchReceipts(from uint64) error {
	log.Debug("Downloading transaction receipts", "origin", from)

	var (
		deliver = func(packet dataPack) (int, error) {
			pack := packet.(*receiptPack)
			return d.queue.DeliverReceipts(pack.peerID, pack.receipts)
		}
		expire   = func() map[string]int { return d.queue.ExpireReceipts(d.requestTTL()) }
		fetch    = func(p *peerConnection, req *fetchRequest) error { return p.FetchReceipts(req) }
		capacity = func(p *peerConnection) int { return p.ReceiptCapacity(d.requestRTT()) }
		setIdle  = func(p *peerConnection, accepted int) { p.SetReceiptsIdle(accepted) }
	)
	err := d.fetchParts(errCancelReceiptFetch, d.receiptCh, deliver, d.receiptWakeCh, expire,
		d.queue.PendingReceipts, d.queue.InFlightReceipts, d.queue.ShouldThrottleReceipts, d.queue.ReserveReceipts,
		d.receiptFetchHook, fetch, d.queue.CancelReceipts, capacity, d.peers.ReceiptIdlePeers, setIdle, "receipts")

	log.Debug("Transaction receipt download terminated", "err", err)
	return err
}

// fetchParts iteratively downloads scheduled block parts, taking any available
// peers, reserving a chunk of fetch requests for each, waiting for delivery and
// also periodically checking for timeouts.
//
// As the scheduling/timeout logic mostly is the same for all downloaded data
// types, this method is used by each for data gathering and is instrumented with
// various callbacks to handle the slight differences between processing them.
//
// The instrumentation parameters:
//  - errCancel:   error type to return if the fetch operation is cancelled (mostly makes logging nicer)
//  - deliveryCh:  channel from which to retrieve downloaded data packets (merged from all concurrent peers)
//  - deliver:     processing callback to deliver data packets into type specific download queues (usually within `queue`)
//  - wakeCh:      notification channel for waking the fetcher when new tasks are available (or sync completed)
//  - expire:      task callback method to abort requests that took too long and return the faulty peers (traffic shaping)
//  - pending:     task callback for the number of requests still needing download (detect completion/non-completability)
//  - inFlight:    task callback for the number of in-progress requests (wait for all active downloads to finish)
//  - throttle:    task callback to check if the processing queue is full and activate throttling (bound memory use)
//  - reserve:     task callback to reserve new download tasks to a particular peer (also signals partial completions)
//  - fetchHook:   tester callback to notify of new tasks being initiated (allows testing the scheduling logic)
//  - fetch:       network callback to actually send a particular download request to a physical remote peer
//  - cancel:      task callback to abort an in-flight download request and allow rescheduling it (in case of lost peer)
//  - capacity:    network callback to retrieve the estimated type-specific bandwidth capacity of a peer (traffic shaping)
//  - idle:        network callback to retrieve the currently (type specific) idle peers that can be assigned tasks
//  - setIdle:     network callback to set a peer back to idle and update its estimated capacity (traffic shaping)
//  - kind:        textual label of the type being downloaded to display in log mesages
func (d *Downloader) fetchParts(errCancel error, deliveryCh chan dataPack, deliver func(dataPack) (int, error), wakeCh chan bool,
	expire func() map[string]int, pending func() int, inFlight func() bool, throttle func() bool, reserve func(*peerConnection, int) (*fetchRequest, bool, error),
	fetchHook func([]*types.Header), fetch func(*peerConnection, *fetchRequest) error, cancel func(*fetchRequest), capacity func(*peerConnection) int,
	idle func() ([]*peerConnection, int), setIdle func(*peerConnection, int), kind string) error {

	// Create a ticker to detect expired retrieval tasks
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	update := make(chan struct{}, 1)
	sum := 0
	timer := time.NewTimer(4 * time.Minute)
	//每次去找peer获取都纪录下来
	fetchPeer := make(map[string]*fetchRequest)
	// Prepare the queue and fetch block parts until the block header fetcher's done
	finished := false
	for {
		select {
		case <-d.cancelCh:
			return errCancel
		case <-timer.C:
			return errors.New("同步超时")
		case packet := <-deliveryCh:
			if kind == "bodies" {
				log.Info("收到了同步的bodies", "id", packet.PeerId())
			}
			delete(fetchPeer, packet.PeerId()) //删除已经收到的
			timer.Reset(4 * time.Minute)
			if sum >= 100 {
				return errors.New("同步的bodies失败次数过多")
			}
			// If the peer was previously banned and failed to deliver its pack
			// in a reasonable time frame, ignore its message.
			if peer := d.peers.Peer(packet.PeerId()); peer != nil {
				// Deliver the received chunk of data and check chain validity
				accepted, err := deliver(packet)
				if err == errInvalidChain {
					return err
				}
				// Unless a peer delivered something completely else than requested (usually
				// caused by a timed out request which came through in the end), set it to
				// idle. If the delivery's stale, the peer should have already been idled.
				if err != errStaleDelivery {
					setIdle(peer, accepted)
				}
				// Issue a log to the user to see what's going on
				switch {
				case err == nil && packet.Items() == 0:
					//数据获取一定次数后,直接结束同步
					sum++
					peer.log.Info("Requested data not delivered", "type", kind, "sum", sum)
				case err == nil:
					sum = 0
					peer.log.Info("Delivered new batch of data", "type", kind, "count", packet.Stats())
				default:
					peer.log.Info("Failed to deliver retrieved data", "type", kind, "err", err)
					return err
				}
			}
			// Blocks assembled, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case cont := <-wakeCh:
			// The header fetcher sent a continuation flag, check if it's done
			if !cont {
				finished = true
			}
			// Headers arrive, try to update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-ticker.C:
			// Sanity check update the progress
			select {
			case update <- struct{}{}:
			default:
			}

		case <-update:
			// Short circuit if we lost all our peers
			if d.peers.Len() == 0 {
				return errNoPeers
			}
			// Check for fetch request timeouts and demote the responsible peers
			for pid, fails := range expire() {
				if peer := d.peers.Peer(pid); peer != nil {
					// If a lot of retrieval elements expired, we might have overestimated the remote peer or perhaps
					// ourselves. Only reset to minimal throughput but don't drop just yet. If even the minimal times
					// out that sync wise we need to get rid of the peer.
					//
					// The reason the minimum threshold is 2 is because the downloader tries to estimate the bandwidth
					// and latency of a peer separately, which requires pushing the measures capacity a bit and seeing
					// how response times reacts, to it always requests one more than the minimum (i.e. min 2).
					if fails > 2 {
						peer.log.Info("Data delivery timed out", "type", kind)
						setIdle(peer, 0)
					} else {
						peer.log.Debug("Stalling delivery, dropping", "type", kind)
						if d.dropPeer == nil {
							// The dropPeer method is nil when `--copydb` is used for a local copy.
							// Timeouts can occur if e.g. compaction hits at the wrong time, and can be ignored
							peer.log.Warn("Downloader wants to drop peer, but peerdrop-function is not set", "peer", pid)
						} else {
							d.dropPeer(pid)
						}
					}
				}
			}

			// If there's nothing more to fetch, wait or terminate

			if pending() == 0 {
				log.Info("下载完成信息", "!inFlight()", !inFlight(), "finished", finished, "type", kind)
				if !inFlight() && finished {
					log.Debug("Data fetching completed", "type", kind)
					return nil
				}
				//查询请求peer是否还在peerSet中
				log.Info("peerSet长度", "d.PeerSet().peers", len(d.PeerSet().peers), "fetchPeer", len(fetchPeer))
				for string, _ := range d.PeerSet().peers {

					log.Info("查询peerSet", "PeerSet", string)
				}
				for id, request := range fetchPeer {
					log.Info("查询ID", "id", id, "d.PeerSet().Peer(id)", d.PeerSet().Peer(id))
					if d.PeerSet().Peer(id) == nil {
						//如果不在,从新获取数据
						log.Info("发现断链节点", "id", id)
						//d.dropPeer(id) //删除没有收到信息的节点
						newPeer := d.PeerSet().RandPeer()
						request.Time = time.Now() //重置请求时间
						delete(fetchPeer, id)     //删除旧的
						d.queue.lock.Lock()
						switch {
						case kind == "bodies":
							delete(d.queue.blockPendPool, id)
							d.queue.blockPendPool[newPeer.id] = request
						case kind == "receipts":
							delete(d.queue.receiptPendPool, id)
							d.queue.receiptPendPool[newPeer.id] = request
						}
						d.queue.lock.Unlock()
						fetchPeer[newPeer.id] = request //添加新的
						log.Info("查询断链后的重新发送的数据", "pendPool", len(d.queue.blockPendPool),
							"fetchPeer", len(fetchPeer), "NewID", newPeer.id)
						fetch(newPeer, request)
					}
				}
				break
			}
			// Send a download request to all idle peers, until throttled
			progressed, throttled, running := false, false, inFlight()
			idles, total := idle()

			for _, peer := range idles {
				// Short circuit if throttling activated
				if throttle() {
					throttled = true
					break
				}
				// Short circuit if there is no more available task.
				if pending() == 0 {
					break
				}
				// Reserve a chunk of fetches for a peer. A nil can mean either that
				// no more headers are available, or that the peer is known not to
				// have them.
				request, progress, err := reserve(peer, capacity(peer))
				if err != nil {
					return err
				}
				if progress {
					progressed = true
				}
				if request == nil {
					continue
				}
				if request.From > 0 {
					peer.log.Info("Requesting new batch of data", "type", kind, "from", request.From)
				} else {
					peer.log.Info("Requesting new batch of data", "type", kind, "count", len(request.Headers), "from", request.Headers[0].Number)
				}
				// Fetch the chunk and make sure any errors return the hashes to the queue
				if fetchHook != nil {
					fetchHook(request.Headers)
				}
				//纪录每个id的请求
				log.Info("纪录发送的id", "id", peer.id)
				fetchPeer[peer.id] = request
				if err := fetch(peer, request); err != nil {
					// Although we could try and make an attempt to fix this, this error really
					// means that we've double allocated a fetch task to a peer. If that is the
					// case, the internal state of the downloader and the queue is very wrong so
					// better hard crash and note the error instead of silently accumulating into
					// a much bigger issue.
					panic(fmt.Sprintf("%v: %s fetch assignment failed", peer, kind))
				}

				running = true
			}
			// Make sure that we have peers available for fetching. If all peers have been tried
			// and all failed throw an error
			if !progressed && !throttled && !running && len(idles) == total && pending() > 0 {
				return errPeersUnavailable
			}
		}
	}
}

// processHeaders takes batches of retrieved headers from an input channel and
// keeps processing and scheduling them into the header chain and downloader's
// queue until the stream ends or a failure occurs.
func (d *Downloader) processHeaders(origin uint64, pivot uint64, td *big.Int) error {
	// Keep a count of uncertain headers to roll back
	rollback := []*types.Header{}
	defer func() {
		if len(rollback) > 0 {
			// Flatten the headers and roll them back
			hashes := make([]common.Hash, len(rollback))
			for i, header := range rollback {
				hashes[i] = header.Hash()
			}
			lastHeader, lastFastBlock, lastBlock := d.lightchain.CurrentHeader().Number, common.Big0, common.Big0
			if d.mode != LightSync {
				lastFastBlock = d.blockchain.CurrentFastBlock().Number()
				lastBlock = d.blockchain.CurrentBlock().Number()
			}
			d.lightchain.Rollback(hashes)
			curFastBlock, curBlock := common.Big0, common.Big0
			if d.mode != LightSync {
				curFastBlock = d.blockchain.CurrentFastBlock().Number()
				curBlock = d.blockchain.CurrentBlock().Number()
			}
			log.Warn("Rolled back headers", "count", len(hashes),
				"header", fmt.Sprintf("%d->%d", lastHeader, d.lightchain.CurrentHeader().Number),
				"fast", fmt.Sprintf("%d->%d", lastFastBlock, curFastBlock),
				"block", fmt.Sprintf("%d->%d", lastBlock, curBlock))
		}
	}()

	// Wait for batches of headers to process
	gotHeaders := false

	for {
		select {
		case <-d.cancelCh:
			return errCancelHeaderProcessing

		case headers := <-d.headerProcCh:
			// Terminate header processing if we synced up
			if len(headers) == 0 {
				// Notify everyone that headers are fully processed
				for _, ch := range []chan bool{d.bodyWakeCh, d.receiptWakeCh} {
					select {
					case ch <- false:
					case <-d.cancelCh:
					}
				}
				// If no headers were retrieved at all, the peer violated its TD promise that it had a
				// better chain compared to ours. The only exception is if its promised blocks were
				// already imported by other means (e.g. fecher):
				//
				// R <remote peer>, L <local node>: Both at block 10
				// R: Mine block 11, and propagate it to L
				// L: Queue block 11 for import
				// L: Notice that R's head and TD increased compared to ours, start sync
				// L: Import of block 11 finishes
				// L: Sync begins, and finds common ancestor at 11
				// L: Request new headers up from 11 (R's TD was higher, it must have something)
				// R: Nothing to give
				if d.mode != LightSync {
					head := d.blockchain.CurrentBlock()
					if !gotHeaders && td.Cmp(d.blockchain.GetTd(head.Hash(), head.NumberU64())) > 0 {
						return errStallingPeer
					}
				}
				// If fast or light syncing, ensure promised headers are indeed delivered. This is
				// needed to detect scenarios where an attacker feeds a bad pivot and then bails out
				// of delivering the post-pivot blocks that would flag the invalid content.
				//
				// This check cannot be executed "as is" for full imports, since blocks may still be
				// queued for processing when the header download completes. However, as long as the
				// peer gave us something useful, we're already happy/progressed (above check).
				if d.mode == FastSync || d.mode == LightSync {
					head := d.lightchain.CurrentHeader()
					if td.Cmp(d.lightchain.GetTd(head.Hash(), head.Number.Uint64())) > 0 {
						return errStallingPeer
					}
				}
				// Disable any rollback and return
				rollback = nil
				return nil
			}
			// Otherwise split the chunk of headers into batches and process them
			gotHeaders = true

			for len(headers) > 0 {
				// Terminate if something failed in between processing chunks
				select {
				case <-d.cancelCh:
					return errCancelHeaderProcessing
				default:
				}
				// Select the next chunk of headers to import
				limit := maxHeadersProcess
				if limit > len(headers) {
					limit = len(headers)
				}
				chunk := headers[:limit]

				// In case of header only syncing, validate the chunk immediately
				if d.mode == FastSync || d.mode == LightSync {
					// Collect the yet unknown headers to mark them as uncertain
					unknown := make([]*types.Header, 0, len(headers))
					for _, header := range chunk {
						if !d.lightchain.HasHeader(header.Hash(), header.Number.Uint64()) {
							unknown = append(unknown, header)
						}
					}
					// If we're importing pure headers, verify based on their recentness
					frequency := fsHeaderCheckFrequency
					if chunk[len(chunk)-1].Number.Uint64()+uint64(fsHeaderForceVerify) > pivot {
						frequency = 1
					}
					if n, err := d.lightchain.InsertHeaderChain(chunk, frequency); err != nil {
						// If some headers were inserted, add them too to the rollback list
						if n > 0 {
							rollback = append(rollback, chunk[:n]...)
						}
						log.Debug("Invalid header encountered", "number", chunk[n].Number, "hash", chunk[n].Hash(), "err", err)
						return errInvalidChain
					}
					// All verifications passed, store newly found uncertain headers
					rollback = append(rollback, unknown...)
					if len(rollback) > fsHeaderSafetyNet {
						rollback = append(rollback[:0], rollback[len(rollback)-fsHeaderSafetyNet:]...)
					}
				}
				// Unless we're doing light chains, schedule the headers for associated content retrieval
				if d.mode == FullSync || d.mode == FastSync {
					// If we've reached the allowed number of pending headers, stall a bit
					for d.queue.PendingBlocks() >= maxQueuedHeaders || d.queue.PendingReceipts() >= maxQueuedHeaders {
						select {
						case <-d.cancelCh:
							return errCancelHeaderProcessing
						case <-time.After(time.Second):
						}
					}
					// Otherwise insert the headers for content retrieval
					inserts := d.queue.Schedule(chunk, origin)
					if len(inserts) != len(chunk) {
						log.Debug("Stale headers")
						return errBadPeer
					}
				}
				headers = headers[limit:]
				origin += uint64(limit)
			}

			// Update the highest block number we know if a higher one is found.
			d.syncStatsLock.Lock()
			if d.syncStatsChainHeight < origin {
				d.syncStatsChainHeight = origin - 1
			}
			d.syncStatsLock.Unlock()

			// Signal the content downloaders of the availablility of new tasks
			for _, ch := range []chan bool{d.bodyWakeCh, d.receiptWakeCh} {
				select {
				case ch <- true:
				default:
				}
			}
		}
	}
}

// processFullSyncContent takes fetch results from the queue and imports them into the chain.
func (d *Downloader) processFullSyncContent() error {
	for {
		results := d.queue.Results(true)
		if len(results) == 0 {
			return nil
		}
		if d.chainInsertHook != nil {
			d.chainInsertHook(results)
		}

		//blocks := types.NewBlockWithHeader(results[0].Header).WithBody(results[0].Transactions, results[0].Uncles)

		//var b bool
		////删除commit
		//if atomic.CompareAndSwapInt32(&d.IsSyncNormalDeleteCommit, 0, 1) {
		//
		//	b = d.SyncNormalDeleteCommit(blocks)
		//}
		//
		//if b {
		//	log.Info("需要发送信息")
		//	if len(d.AssociatedCH) != 0 {
		//		<-d.AssociatedCH
		//	}
		//	d.AssociatedCH <- b //删除完毕  可以保存commit了
		//	log.Info("AssociatedCH发送完毕")
		//}

		if err := d.importBlockResults(results); err != nil {
			return err
		}
	}
}

func (d *Downloader) importBlockResults(results []*fetchResult) error {
	// Check for any early termination requests
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	default:
	}
	// Retrieve the a batch of results to import
	first, last := results[0].Header, results[len(results)-1].Header
	log.Info("Inserting downloaded chain", "items", len(results),
		"firstnum", first.Number, "firsthash", first.Hash(),
		"lastnum", last.Number, "lasthash", last.Hash(),
	)
	blocks := make([]*types.Block, len(results))
	for i, result := range results {
		blocks[i] = types.NewBlockWithHeader(result.Header).WithBody(result.Transactions, result.Uncles)
	}

	if index, err := d.blockchain.InsertChain(blocks); err != nil {
		log.Error("Downloaded item processing failed Inserting downloaded chain", "number", results[index].Header.Number, "hash", results[index].Header.Hash(), "err", err)
		return errInvalidChain
	}
	return nil
}

// processFastSyncContent takes fetch results from the queue and writes them to the
// database. It also controls the synchronisation of state nodes of the pivot block.
func (d *Downloader) processFastSyncContent(latest *types.Header) error {
	// Start syncing state of the reported head block. This should get us most of
	// the state of the pivot block.
	stateSync := d.syncState(latest.Root)
	defer stateSync.Cancel()
	go func() {
		if err := stateSync.Wait(); err != nil && err != errCancelStateFetch {
			d.queue.Close() // wake up WaitResults
		}
	}()
	// Figure out the ideal pivot block. Note, that this goalpost may move if the
	// sync takes long enough for the chain head to move significantly.
	pivot := uint64(0)
	if height := latest.Number.Uint64(); height > uint64(fsMinFullBlocks) {
		pivot = height - uint64(fsMinFullBlocks)
	}
	// To cater for moving pivot points, track the pivot block and subsequently
	// accumulated download results separately.
	var (
		oldPivot *fetchResult   // Locked in pivot block, might change eventually
		oldTail  []*fetchResult // Downloaded content after the pivot
	)
	log.Info("同步的最新高度和pivot", "最新高度", latest.Number.Uint64(), "pivot", pivot)
	for {
		// Wait for the next batch of downloaded data to be available, and if the pivot
		// block became stale, move the goalpost
		results := d.queue.Results(oldPivot == nil) // Block if we're not monitoring pivot staleness
		if len(results) != 0 {
			fetchResults := results[len(results)-1:][0]

			log.Info("results的最新高度", "最新高度", fetchResults.Header.Number)
		}
		if len(results) == 0 {
			// If pivot sync is done, stop
			if oldPivot == nil {
				return stateSync.Cancel()
			}
			// If sync failed, stop
			select {
			case <-d.cancelCh:
				return stateSync.Cancel()
			default:
			}
		}
		if d.chainInsertHook != nil {
			d.chainInsertHook(results)
		}
		if oldPivot != nil {
			results = append(append([]*fetchResult{oldPivot}, oldTail...), results...)
		}
		// Split around the pivot block and process the two sides via fast/full sync
		if atomic.LoadInt32(&d.committed) == 0 {
			latest = results[len(results)-1].Header
			if height := latest.Number.Uint64(); height > pivot+2*uint64(fsMinFullBlocks) {
				log.Warn("Pivot became stale, moving", "old", pivot, "new", height-uint64(fsMinFullBlocks))
				pivot = height - uint64(fsMinFullBlocks)
			}
		}
		P, beforeP, afterP := splitAroundPivot(pivot, results)
		if beforeP != nil {
			log.Info("beforeP处理", "before长度", len(beforeP), "before最后高度", beforeP[len(beforeP)-1:][0].Header.Number)
		}

		if err := d.commitFastSyncData(beforeP, stateSync); err != nil {
			return err
		}
		if P != nil {
			// If new pivot block found, cancel old state retrieval and restart
			if oldPivot != P {
				stateSync.Cancel()

				stateSync = d.syncState(P.Header.Root)
				log.Info("处理P点", "高度", P.Header.Number)
				defer stateSync.Cancel()
				go func() {
					if err := stateSync.Wait(); err != nil && err != errCancelStateFetch {
						d.queue.Close() // wake up WaitResults
					}
				}()
				oldPivot = P
			}
			// Wait for completion, occasionally checking for pivot staleness
			select {
			case <-stateSync.done:
				if stateSync.err != nil {
					return stateSync.err
				}
				log.Info("开始处理P点", "高度", P.Header.Number)
				if err := d.commitPivotBlock(P); err != nil {
					return err
				}
				log.Info("P点处理完毕", "高度", P.Header.Number)

				oldPivot = nil

			case <-time.After(time.Second):
				oldTail = afterP
				continue
			}
		}
		// Fast sync done, pivot commit done, full import
		if afterP != nil {
			log.Info("afterP处理", "afterP长度", len(afterP))

		}
		if err := d.importBlockResults(afterP); err != nil {
			return err
		}
	}
}

func splitAroundPivot(pivot uint64, results []*fetchResult) (p *fetchResult, before, after []*fetchResult) {
	for _, result := range results {
		num := result.Header.Number.Uint64()
		switch {
		case num < pivot:
			before = append(before, result)
		case num == pivot:
			p = result
		default:
			after = append(after, result)
		}
	}
	return p, before, after
}

func (d *Downloader) commitFastSyncData(results []*fetchResult, stateSync *stateSync) error {
	// Check for any early termination requests
	if len(results) == 0 {
		return nil
	}
	select {
	case <-d.quitCh:
		return errCancelContentProcessing
	case <-stateSync.done:
		if err := stateSync.Wait(); err != nil {
			return err
		}
	default:
	}
	// Retrieve the a batch of results to import
	first, last := results[0].Header, results[len(results)-1].Header
	log.Debug("Inserting fast-sync blocks", "items", len(results),
		"firstnum", first.Number, "firsthash", first.Hash(),
		"lastnumn", last.Number, "lasthash", last.Hash(),
	)
	blocks := make([]*types.Block, len(results))
	receipts := make([]types.Receipts, len(results))
	for i, result := range results {
		blocks[i] = types.NewBlockWithHeader(result.Header).WithBody(result.Transactions, result.Uncles)
		receipts[i] = result.Receipts
	}
	if index, err := d.blockchain.InsertReceiptChain(blocks, receipts); err != nil {
		log.Debug("Downloaded item processing failed", "number", results[index].Header.Number, "hash", results[index].Header.Hash(), "err", err)
		return errInvalidChain
	}
	return nil
}

func (d *Downloader) commitPivotBlock(result *fetchResult) error {
	block := types.NewBlockWithHeader(result.Header).WithBody(result.Transactions, result.Uncles)
	log.Debug("Committing fast sync pivot as new head", "number", block.Number(), "hash", block.Hash())
	if _, err := d.blockchain.InsertReceiptChain([]*types.Block{block}, []types.Receipts{result.Receipts}); err != nil {
		return err
	}
	if err := d.blockchain.FastSyncCommitHead(block.Hash()); err != nil {
		return err
	}
	atomic.StoreInt32(&d.committed, 1)
	return nil
}

// DeliverHeaders injects a new batch of block headers received from a remote
// node into the download schedule.
func (d *Downloader) DeliverHeaders(id string, headers []*types.Header) (err error) {
	return d.deliver(id, d.headerCh, &headerPack{id, headers}, headerInMeter, headerDropMeter)
}

// DeliverBodies injects a new batch of block bodies received from a remote node.
func (d *Downloader) DeliverBodies(id string, transactions [][]*types.Transaction, uncles [][]*types.Header) (err error) {
	return d.deliver(id, d.bodyCh, &bodyPack{id, transactions, uncles}, bodyInMeter, bodyDropMeter)
}

// DeliverReceipts injects a new batch of receipts received from a remote node.
func (d *Downloader) DeliverReceipts(id string, receipts [][]*types.Receipt) (err error) {
	return d.deliver(id, d.receiptCh, &receiptPack{id, receipts}, receiptInMeter, receiptDropMeter)
}

// DeliverNodeData injects a new batch of node state data received from a remote node.
func (d *Downloader) DeliverNodeData(id string, data [][]byte) (err error) {
	return d.deliver(id, d.stateCh, &statePack{id, data}, stateInMeter, stateDropMeter)
}

// deliver injects a new batch of data received from a remote node.
func (d *Downloader) deliver(id string, destCh chan dataPack, packet dataPack, inMeter, dropMeter metrics.Meter) (err error) {
	// Update the delivery metrics for both good and failed deliveries
	inMeter.Mark(int64(packet.Items()))
	defer func() {
		if err != nil {
			dropMeter.Mark(int64(packet.Items()))
		}
	}()
	// Deliver or abort if the sync is canceled while queuing
	d.cancelLock.RLock()
	cancel := d.cancelCh
	d.cancelLock.RUnlock()
	if cancel == nil {
		return errNoSyncActive
	}
	select {
	case destCh <- packet:
		return nil
	case <-cancel:
		return errNoSyncActive
	}
}

//PDX
func (d *Downloader) DeliverCommits(id string, commits []*types.Block) (err error) {
	d.commitCh <- &commitPack{id, commits}
	log.Debug("deliver commits to commitCh", "len", len(commits))
	return nil
}

// qosTuner is the quality of service tuning loop that occasionally gathers the
// peer latency statistics and updates the estimated request round trip time.
func (d *Downloader) qosTuner() {
	for {
		// Retrieve the current median RTT and integrate into the previoust target RTT
		rtt := time.Duration((1-qosTuningImpact)*float64(atomic.LoadUint64(&d.rttEstimate)) + qosTuningImpact*float64(d.peers.medianRTT()))
		atomic.StoreUint64(&d.rttEstimate, uint64(rtt))

		// A new RTT cycle passed, increase our confidence in the estimated RTT
		conf := atomic.LoadUint64(&d.rttConfidence)
		conf = conf + (1000000-conf)/2
		atomic.StoreUint64(&d.rttConfidence, conf)

		// Log the new QoS values and sleep until the next RTT
		log.Debug("Recalculated downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
		select {
		case <-d.quitCh:
			return
		case <-time.After(rtt):
		}
	}
}

// qosReduceConfidence is meant to be called when a new peer joins the downloader's
// peer set, needing to reduce the confidence we have in out QoS estimates.
func (d *Downloader) qosReduceConfidence() {
	// If we have a single peer, confidence is always 1
	peers := uint64(d.peers.Len())
	if peers == 0 {
		// Ensure peer connectivity races don't catch us off guard
		return
	}
	if peers == 1 {
		atomic.StoreUint64(&d.rttConfidence, 1000000)
		return
	}
	// If we have a ton of peers, don't drop confidence)
	if peers >= uint64(qosConfidenceCap) {
		return
	}
	// Otherwise drop the confidence factor
	conf := atomic.LoadUint64(&d.rttConfidence) * (peers - 1) / peers
	if float64(conf)/1000000 < rttMinConfidence {
		conf = uint64(rttMinConfidence * 1000000)
	}
	atomic.StoreUint64(&d.rttConfidence, conf)

	rtt := time.Duration(atomic.LoadUint64(&d.rttEstimate))
	log.Debug("Relaxed downloader QoS values", "rtt", rtt, "confidence", float64(conf)/1000000.0, "ttl", d.requestTTL())
}

// requestRTT returns the current target round trip time for a download request
// to complete in.
//
// Note, the returned RTT is .9 of the actually estimated RTT. The reason is that
// the downloader tries to adapt queries to the RTT, so multiple RTT values can
// be adapted to, but smaller ones are preferred (stabler download stream).
func (d *Downloader) requestRTT() time.Duration {
	return time.Duration(atomic.LoadUint64(&d.rttEstimate)) * 9 / 10
}

// requestTTL returns the current timeout allowance for a single download request
// to finish under.
func (d *Downloader) requestTTL() time.Duration {
	var (
		rtt  = time.Duration(atomic.LoadUint64(&d.rttEstimate))
		conf = float64(atomic.LoadUint64(&d.rttConfidence)) / 1000000.0
	)
	ttl := time.Duration(ttlScaling) * time.Duration(float64(rtt)/conf)
	if ttl > ttlLimit {
		ttl = ttlLimit
	}
	ttl = ttl * 4
	if ttl <= 30 {
		ttl = 30
	}
	//ttl := time.Second * 20
	return ttl
}

func (d *Downloader) VerifyQualification(head *types.Header) bool {
	//先判断收到块的作者在不在委员会中
	commitHeight := d.Commitchain.CurrentBlock().NumberU64()
	consensusQuorum, ok := quorum.CommitHeightToConsensusQuorum.Get(commitHeight, d.stateDB)
	if !ok {
		log.Warn("同步验证区块,获取委员会失败")
		return true
	}
	//if consensusQuorum.Len()==1&&consensusQuorum.Contains(d.Worker.Signer()){ //委员会里只有自己,接收其他区块
	if consensusQuorum.Len() == 1 { //委员会里只有自己,接收其他区块

		return true
	}
	if _, ok := consensusQuorum.Hmap[head.Coinbase.String()]; ok {
		log.Info("同步收到的块是委员会中的成员", "收到区块的地址", head.Coinbase.String())
		return true
	}
	land, _ := engine.LocalLandSetMap.LandMapGet(commitHeight, d.stateDB)

	if land.IslandState && len(land.IslandQuorum) == 0 {
		log.Warn("岛屿状态IslandQuorum为空,直接进行同步")
		return true
	}

	if land.IslandState {
		for _, add := range land.IslandQuorum {
			if add == head.Coinbase.String() {
				log.Info("同步收到的块是分叉前委员会中的成员", "收到区块的地址", head.Coinbase.String())
				return true
			}
		}
	}

	log.Warn("同步区块验证失败,收到的区块无效", "当前commit高度", commitHeight)
	return false

}

func (d *Downloader) CommitSyncIsland(origin uint64) {

	if atomic.LoadInt32(&d.AssociatedSynchronising) == 1 {
		log.Warn("SyncIsland或者SyncNormalDeleteCommit半生commit正在同步")
		return
	}
	log.Info("CommitSyncIslandSyncIsland开始删除commit", "current", d.Commitchain.CurrentBlock().NumberU64())

	currentCommitNum := d.Commitchain.CurrentBlock().NumberU64()
	if currentCommitNum == 0 {
		return
	}

	if commit := d.Commitchain.GetBlockByNum(currentCommitNum + 1); commit != nil {
		//在insert的时候有可能是没有更新current,所以查找下
		currentCommitNum = commit.NumberU64()
	}

	for i := origin; i <= currentCommitNum; i++ {
		if i == 0 {
			continue
		}
		block := d.Commitchain.GetBlockByNum(i)
		log.Info("要删除的高度", "commit祖先的高度", origin, "当前要删的高度", i, "要删除到哪里", currentCommitNum)
		//先删除 在同步
		if block == nil {
			continue
		}
		//先删除 在同步
		engine.CommitDeRepetition.Del(block.NumberU64())           //commit删除缓存
		engine.AssociatedCommitDeRepetition.Del(block.NumberU64()) //commit删除缓存
		core.DeleteCommitBlock(d.stateDB, block.Hash(), block.NumberU64(), *d.Commitchain.CommitChain())
		engine.CommitBlockQueued.DelToProcessCommitBlockQueueMap(i) //删除toProcessCommit缓存
		quorum.CommitHeightToConsensusQuorum.Del(block.NumberU64(), d.stateDB)
		qualification.CommitHeight2NodeDetailSetCache.Del(block.NumberU64(), d.stateDB)
		engine.MultiSign.DelMultiSign() //清除多签

	}

	//如果当前的commit高度已经比裂脑前的高度小,就用当前高度
	if origin != 0 {
		origin -= 1
	}
	localCommit := d.Commitchain.GetBlockByNum(origin)
	rawdb.WriteHeadCommitBlockHash(d.stateDB, localCommit.Hash()) //恢复LastCommitBlock
	core.SetCurrentCommitBlock(d.Commitchain.CommitChain(), localCommit)
	currentBlockNumber := d.Commitchain.CurrentBlock().NumberU64()
	blockExtra, _ := types2.CommitExtraDecode(localCommit)
	//下次更新委员会的时间
	if testQuorum, ok := quorum.UpdateQuorumSnapshots.GetUpdateQuorum(blockExtra.HistoryHeight, d.stateDB); !ok {
		quorum.UpdateQuorums = quorum.NewUpdateQuorum()
	} else {
		//复制bigInt
		quorum.UpdateQuorums.CopyUpdateQuorum(testQuorum)
	}

	log.Info("quorum.UpdateQuorums", "当前高度", localCommit.NumberU64(), "下次更新高度", quorum.UpdateQuorums.NextUpdateHeight)
	if localCommit.Number().Cmp(quorum.UpdateQuorums.NextUpdateHeight) == 0 {
		//如果下次更新委员会的高度等于当前commit高度,再用当前高度取一次更新委员会的高度
		if testQuirum, ok := quorum.UpdateQuorumSnapshots.GetUpdateQuorum(localCommit.Number(), d.stateDB); !ok {
			quorum.UpdateQuorums = quorum.NewUpdateQuorum()
		} else {
			//复制bigInt
			quorum.UpdateQuorums.CopyUpdateQuorum(testQuirum)
		}
		log.Info("quorum.UpdateQuorums第二次取", "当前高度", localCommit, "下次更新高度", quorum.UpdateQuorums.NextUpdateHeight)

	}

	log.Info("currentCommit高度", "commit高度", currentBlockNumber, "回滚后下次更新委员会的高度", quorum.UpdateQuorums.NextUpdateHeight, "blockExtra.HistoryHeight", blockExtra.HistoryHeight)

}
