package chain

import (
	"chaos/api/log"
	"chaos/api/model"
	"chaos/api/system"
	"context"
	"sync"
	"time"
)

var topupTxQueue = system.NewRichQueue[QueuePassObject]()
var startOnce sync.Once

func AppendTopupTx(chainID uint64, txHash string, mainID uint64, refFlowID uint64, op int) {
	topupTxQueue.Enqueue(QueuePassObject{
		ChainID:   chainID,
		TxHash:    txHash,
		MainID:    mainID,
		RefFlowID: refFlowID,
		Op:        op,
	})
}

var topupHashChannel = make(chan *TopupTxInfo, 500)

func StartTopupConsumer(ctx context.Context) {
	startOnce.Do(func() {
		go func() {
			defer func() {
				log.Info("Topup consumer goroutine shutting down...")
			}()
			topupTxQueue.ConsumerWithContext(ctx, 1, func(chainHash QueuePassObject, lock *sync.WaitGroup) {
				// parts := strings.SplitN(chainHash, "-", 2)
				// if len(parts) != 2 {
				// 	log.Errorf("bad key: %s", chainHash)
				// 	return
				// }
				// cid, err := strconv.ParseUint(parts[0], 10, 64)
				// if err != nil || len(parts[1]) != 66 || !strings.HasPrefix(parts[1], "0x") {
				// 	return
				// }

				var topupInfo *TopupTxInfo
				var err error
				var op = chainHash.Op
				switch op {
				case model.BalanceFlowOpRecharge:
					topupInfo, err = ParseTopupTx(chainHash.ChainID, chainHash.TxHash)
				case model.BalanceFlowOpFreeze:
					topupInfo, err = ParseOpenLockTx(chainHash.ChainID, chainHash.TxHash)
				case model.BalanceFlowOpWithdraw:
					topupInfo, err = ParseClaimLockedTx(chainHash.ChainID, chainHash.TxHash)
				case model.BalanceFlowOpUnfreeze:
					topupInfo, err = ParseCancelLockedTx(chainHash.ChainID, chainHash.TxHash)
					// case model.BalanceFlowOpWithdraw:
					// 	topupInfo, err = ParseTopupTx(chainHash.ChainID, chainHash.TxHash)
					// case model.BalanceFlowOpUnfreeze:
					// 	topupInfo, err = ParseTopupTx(chainHash.ChainID, chainHash.TxHash)
				}

				// topupInfo, err := ParseTopupTx(chainHash.ChainID, chainHash.TxHash)
				if err != nil {
					log.Error("parse topup tx failed: ", err, topupInfo)
					topupTxQueue.EnqueueWithDelay(chainHash, 10*time.Second)
					return
				}
				if topupInfo == nil {
					log.Error("unsupported tx object parse, so ignore it: ", chainHash.TxHash)
					return
				}
				if topupInfo.Status == "pending" {
					log.Info("topup tx pending, will retry after delay", chainHash.TxHash)
					topupTxQueue.EnqueueWithDelay(chainHash, 5*time.Second)
					return
				}
				topupInfo.MainID = chainHash.MainID
				topupInfo.RefFlowID = chainHash.RefFlowID
				topupInfo.Op = chainHash.Op
				topupInfo.ChainID = chainHash.ChainID
				publishUpdate(topupInfo)
			})
		}()
	})
}

func UpdateTopupTx() <-chan *TopupTxInfo {
	return topupHashChannel
}

func publishUpdate(info *TopupTxInfo) {
	select {
	case topupHashChannel <- info:
	default:
		log.Warnf("topup updates channel full; please handle tx in time manually: %s", info.TxHash)
	}
}
