package engine

import (
	"pdx-chain/common"
	core_types "pdx-chain/core/types"
	"pdx-chain/ethdb"
	"pdx-chain/log"
	"pdx-chain/quorum"
	"pdx-chain/utopia/engine/qualification"
	"pdx-chain/utopia/types"
)

func UpdateConsensusQuorum(w *utopiaWorker, commitExtra *types.CommitExtra, block *core_types.Block) bool {
	//更新活跃度
	commitHeight := block.NumberU64()

	currentCommit := w.blockchain.CommitChain.GetBlockByNum(block.NumberU64() - 1)
	var currentCommitHeight uint64
	//var currentCommitHash common.Hash
	if currentCommit != nil {
		currentCommitHeight = currentCommit.NumberU64()
		//currentCommitHash = currentCommit.Hash()
	} else {
		currentCommitHeight = 0
	}
	//更新委员会
	quorumAdditions := commitExtra.MinerAdditions
	commitHash := block.Hash()
	switch {
	case currentCommitHeight == 0:
		//第一次更新委员会
		firstMinerAdditions(quorumAdditions, w, commitHeight)
	case currentCommitHeight <= quorum.LessCommit || PerQuorum:
		log.Info("查询PerQuorum的数据", "PerQuorum", PerQuorum)
		//commit高度小于100,每次都更新委员会
		if !lessMinerAdditions(quorumAdditions, w, commitHeight, commitExtra) {
			log.Error("lessMinerAdditions fail")
			return false
		}
		if currentCommitHeight == quorum.LessCommit && quorum.UpdateQuorums.HistoryUpdateHeight.Uint64() == 0 &&
			!PerQuorum {
			//下次更新委员会的高度
			quorum.UpdateQuorums.CalculateNextUpdateHeight(commitHash)
			quorum.UpdateQuorums.CalculateNextAfterHeight(commitHash)
			quorum.UpdateQuorums.HistoryUpdateHeight = block.Number()
			if !quorum.UpdateQuorumSnapshots.SetUpdateQuorum(quorum.UpdateQuorums, w.utopia.db) {
				log.Error("UpdateQuorums.SetUpdateQuorum fail")
				return false
			}
			log.Info("开发模式关闭", "下次更新高度", quorum.UpdateQuorums.NextUpdateHeight)
		}
	default:
		//commit高度大于100,按照新规则进行更新委员会列表
		//增加委员会列表成员

		currentQuorum, ok := quorum.CommitHeightToConsensusQuorum.Get(commitHeight-1, w.utopia.db)
		if !ok {
			log.Error("quorum.CommitHeightToConsensusQuorum fail")
			return false
		}

		currentQuorumCopy := currentQuorum.Copy()
		//清理失去资格的委员会成员

		currentQuorumCopy = delConsensusQuorum(currentQuorumCopy, w, commitHeight, commitExtra.MinerDeletions)

		quorum.CommitHeightToConsensusQuorum.Set(commitHeight, currentQuorumCopy, w.utopia.db)

		log.Info("下次更新高度", "下次的高度", quorum.UpdateQuorums.NextUpdateHeight,
			"block.Number()", block.Number())

		if block.Number().Cmp(quorum.UpdateQuorums.NextUpdateHeight) == 0 {
			set := quorum.SortPrepareQuorum(commitExtra.MinerAdditions, commitHash, w.utopia.db)
			log.Info("开始更新本地委员会", "更新高度", quorum.UpdateQuorums.NextUpdateHeight)

			if set != nil {
				for _, add := range set {
					log.Info("新增成员", "add", add.String())
					currentQuorumCopy.Add(add.String(), add)
				}
				//更新当前高度委员会
				quorum.CommitHeightToConsensusQuorum.Set(commitHeight, currentQuorumCopy, w.utopia.db)
			}
		}
	}

	return true
}

func UpdateQuorumsHeight(block *core_types.Block, db ethdb.Database) bool {
	if block.Number().Cmp(quorum.UpdateQuorums.NextUpdateHeight) == 0 {
		//如果到了更新委员会的
		quorum.UpdateQuorums.CalculateNextUpdateHeight(block.Hash())
		quorum.UpdateQuorums.CalculateNextAfterHeight(block.Hash())
		quorum.UpdateQuorums.HistoryUpdateHeight = block.Number()
		if !quorum.UpdateQuorumSnapshots.SetUpdateQuorum(quorum.UpdateQuorums, db) {
			log.Error("UpdateQuorums.SetUpdateQuorum fail")
			return false
		}

	}
	log.Info("开始储存", "下次更新高度", quorum.UpdateQuorums.NextUpdateHeight,
		"本次更新高度", quorum.UpdateQuorums.HistoryUpdateHeight)
	return true
}

func firstMinerAdditions(quorumAdditions []common.Address, w *utopiaWorker, commitHeight uint64) {
	// recv'd first commit block after genesis commit block
	var commitQuorum = quorum.NewNodeAddress()
	if len(quorumAdditions) > 0 {
		for _, address := range quorumAdditions {
			commitQuorum.Add(address.Hex(), address)
		}
	}
	quorum.CommitHeightToConsensusQuorum.Set(commitHeight, commitQuorum, w.utopia.db)
}

func lessMinerAdditions(quorumAdditions []common.Address, w *utopiaWorker, commitHeight uint64, commitExtra *types.CommitExtra) bool {
	preCommitHeight := commitHeight - 1
	if preCommitHeight == 0 {
		preCommitHeight = 1
	}
	// retrieve previous node set
	// ignore genesisCommitBlock
	currentQuorum, ok := quorum.CommitHeightToConsensusQuorum.Get(preCommitHeight, w.utopia.db)
	if !ok {
		// the current commit block does NOT have quorum
		return false
	} else {
		commitQuorum := currentQuorum.Copy()
		// combine (add/delete)
		delConsensusQuorum(commitQuorum, w, commitHeight, commitExtra.MinerDeletions)

		if len(quorumAdditions) > 0 {
			for _, address := range quorumAdditions {
				commitQuorum.Add(address.Hex(), address)
			}
		}

		//if ConsensusQuorumLimt != 100 && ConsensusQuorumLimt != 0 {
		//	nodesDetails, _ := qualification.CommitHeight2NodeDetailSetCache.Get(commitHeight, w.utopia.db)
		//
		//	nodes := make(qualification.ByQualificationIndex, 0)
		//	for address, _ := range commitQuorum.Hmap {
		//		nodeDetail := nodesDetails.Get(address)
		//		if nodeDetail != nil {
		//			nodes = append(nodes, nodeDetail)
		//		}
		//	}
		//
		//	sort.Sort(nodes)
		//
		//	flt := float64(ConsensusQuorumLimt) / float64(100)
		//
		//	index := int(math.Ceil(float64(len(nodes)) * flt))
		//
		//	nodes = nodes[:index]
		//
		//	for address, _ := range commitQuorum.Hmap {
		//		nodesDetail := nodesDetails.Get(address)
		//		if !contains(nodes, nodesDetail) {
		//			commitQuorum.Del(address)
		//		}
		//	}
		//}

		quorum.CommitHeightToConsensusQuorum.Set(commitHeight, commitQuorum, w.utopia.db)

	}
	return true
}

func delConsensusQuorum(currentQuorum *quorum.NodeAddress, w *utopiaWorker, commitHeight uint64, quorumDeletions []common.Address) *quorum.NodeAddress {
	//删除委员会,并且清理活跃度
	if len(quorumDeletions) > 0 {
		for _, address := range quorumDeletions {
			currentQuorum.Del(address.Hex())
			currentHeightNodeDetails, ok := qualification.CommitHeight2NodeDetailSetCache.Get(commitHeight, w.utopia.db)
			if ok {
				nodeDetail := currentHeightNodeDetails.Get(address.Hex())
				if nodeDetail != nil {
					nodeDetail = CleanUpNodeDetailInfo(nodeDetail, commitHeight)
					currentHeightNodeDetails.Add(address.String(), nodeDetail)
					qualification.CommitHeight2NodeDetailSetCache.Set(commitHeight, currentHeightNodeDetails, w.utopia.db)
				}
			}
		}
	}
	return currentQuorum

}
