package handlers

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"open-indexer/model"
	"strings"
	"time"
)

// var dataStartBlock = 31918263
// var dataEndBlock = 39206794
var dataStartBlock uint64
var dataEndBlock uint64

var syncFromBlock uint64
var syncToBlock uint64
var fetchSize uint64

var latestBlock uint64
var createSnapshotFlag bool
var createSnapshotBlock uint64

var syncInterrupt bool

func initSync() {
	synCfg := cfg.Section("sync")
	dataStartBlock = synCfg.Key("start").MustUint64(0)
	dataEndBlock = synCfg.Key("end").MustUint64(0)
	fetchSize = synCfg.Key("size").MustUint64(1)

	syncFromBlock = dataStartBlock

	if dataEndBlock > 0 && dataStartBlock > dataEndBlock {
		panic("block number error")
	}
}

func StartSync() {
	for !syncInterrupt {
		finished, err := syncBlock()
		if err != nil {
			if strings.HasPrefix(err.Error(), "sync: no more new block") {
				logger.Println(err.Error() + ", wait 1s")
				time.Sleep(time.Duration(1) * time.Second)
				continue
			}
			logger.Errorln("sync error:", err)
			QuitChan <- true
			break
		}
		if finished {
			logger.Println("sync finished")
			QuitChan <- true
			break
		}
	}

	StopSuccessCount++
	logger.Println("sync stopped")
}

func StopSync() {
	syncInterrupt = true
}

func syncBlock() (bool, error) {

	syncToBlock = syncFromBlock + fetchSize - 1

	// Modify parameters for faster synchronization
	//if syncFromBlock < 37400000 {
	//	syncToBlock = syncFromBlock + 100000 - 1
	//} else if syncFromBlock < 37900000 {
	//	syncToBlock = syncFromBlock + 50000 - 1
	//} else if syncFromBlock < 38400000 {
	//	syncToBlock = syncFromBlock + 5000 - 1
	//} else if syncFromBlock < 38900000 {
	//	syncToBlock = syncFromBlock + 2000 - 1
	//} else if syncFromBlock < 40000000 {
	//	syncToBlock = syncFromBlock + 500 - 1
	//} else if syncFromBlock < 40560000 {
	//	syncToBlock = syncFromBlock + 1000 - 1
	//}

	if dataEndBlock > 0 && syncToBlock > dataEndBlock {
		syncToBlock = dataEndBlock
	}

	// read trxs
	if latestBlock == 0 || syncToBlock >= latestBlock {
		var err error
		latestBlock, err = getLatestBlock()
		if err != nil {
			return false, err
		}

		if latestBlock > syncFromBlock && latestBlock-syncFromBlock < 10 {
			// It's catching up. read it block by block.
			fetchSize = 1
		}

		if latestBlock < syncFromBlock-1 {
			return false, errors.New("the latest block is smaller than the current block")
		}
	}

	if syncToBlock > latestBlock {
		syncToBlock = latestBlock
	}

	if syncFromBlock > syncToBlock {
		checkSnapshot()
		return false, errors.New(fmt.Sprintf("sync: no more new block, block %d", syncToBlock))
	}

	start := time.Now().UnixMilli()
	logger.Printf("sync block from %d to %d", syncFromBlock, syncToBlock)
	trxs, err := getTransactions()
	if err != nil {
		return false, err
	}

	// read logs
	logs, err := getLogs()
	if err != nil {
		return false, err
	}

	records := mixRecords(trxs, logs)
	err = processRecords(records)
	if err != nil {
		return false, err
	}

	err = saveToStorage(syncToBlock)
	if err != nil {
		return false, err
	}

	checkSnapshot()

	costs := time.Now().UnixMilli() - start
	logger.Println("sync finished, costs ", costs, " ms")

	syncFromBlock = syncToBlock + 1

	return syncToBlock == dataEndBlock, err
}

func checkSnapshot() {
	// create a snapshot
	if createSnapshotFlag || syncToBlock == createSnapshotBlock {
		createSnapshotFlag = false
		if syncToBlock == createSnapshotBlock {
			createSnapshotBlock = 0
		}
		snapshot(syncToBlock)
	}
}

func getLatestBlock() (uint64, error) {
	if DataSourceType == "mongo" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		statusCollection := mongodb.Collection("status")
		result := statusCollection.FindOne(ctx, bson.D{})
		if result.Err() != nil {
			return 0, result.Err()
		}
		var status model.Status
		result.Decode(&status)

		return status.Block, nil
	} else {
		blockNumber := cachedBlockNumber
		if blockNumber == 0 {
			blockNumber = syncFromBlock - 1
		}
		return blockNumber, nil
	}
}

func getTransactions() ([]*model.Transaction, error) {
	var trxs []*model.Transaction
	if DataSourceType == "mongo" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		trxCollection := mongodb.Collection("transactions")
		cur, err := trxCollection.Find(ctx, bson.D{{"block", bson.D{{"$gte", syncFromBlock}, {"$lte", syncToBlock}}}})
		defer cur.Close(ctx)
		if err != nil {
			logger.Println(err)
			return nil, err
		}
		for cur.Next(ctx) {
			var result model.Transaction
			err := cur.Decode(&result)
			if err != nil {
				logger.Fatal(err)
			}
			trxs = append(trxs, &result)
		}
	} else {
		for block := syncFromBlock; block <= syncToBlock; block++ {
			_trxs, ok := cachedTranscriptions[block]
			if !ok {
				if block == syncFromBlock {
					return nil, errors.New(fmt.Sprintf("trxs not cached at %d", block))
				} else {
					break
				}
			} else {
				delete(cachedTranscriptions, block)
			}
			trxs = append(trxs, _trxs...)
		}
	}
	return trxs, nil
}

func getLogs() ([]*model.EvmLog, error) {
	var logs []*model.EvmLog
	if DataSourceType == "mongo" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logCollection := mongodb.Collection("evmlogs")
		cur, err := logCollection.Find(ctx, bson.D{{"block", bson.D{{"$gte", syncFromBlock}, {"$lte", syncToBlock}}}})
		defer cur.Close(ctx)
		if err != nil {
			return nil, err
		}
		for cur.Next(ctx) {
			var result model.EvmLog
			err := cur.Decode(&result)
			if err != nil {
				logger.Fatal(err)
			}
			logs = append(logs, &result)
		}
	} else {
		for block := syncFromBlock; block <= syncToBlock; block++ {
			_logs, ok := cachedLogs[block]
			if !ok {
				if block == syncFromBlock {
					return nil, errors.New(fmt.Sprintf("logs not cached at %d", block))
				} else {
					break
				}
			} else {
				delete(cachedLogs, block)
			}
			logs = append(logs, _logs...)
		}
	}
	return logs, nil
}
