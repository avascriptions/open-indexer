package handlers

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"open-indexer/model"
	"time"
)

// var dataStartBlock = 31918263
// var dataEndBlock = 39206794
var dataStartBlock uint64
var dataEndBlock uint64

var fetchFromBlock uint64
var fetchToBlock uint64
var fetchSize uint64

var latestBlock uint64
var createSnapshotFlag bool
var createSnapshotBlock uint64

func initSync() {
	synCfg := cfg.Section("sync")
	dataStartBlock = synCfg.Key("start").MustUint64(0)
	dataEndBlock = synCfg.Key("end").MustUint64(0)
	fetchSize = synCfg.Key("size").MustUint64(1)

	fetchFromBlock = dataStartBlock

	if dataEndBlock > 0 && dataStartBlock > dataEndBlock {
		panic("block number error")
	}
}

func SyncBlock() (bool, error) {
	var trxs []*model.Transaction
	var logs []*model.EvmLog

	fetchToBlock = fetchFromBlock + fetchSize - 1

	// Modify parameters for faster synchronization
	//if fetchFromBlock < 37400000 {
	//	fetchToBlock = fetchFromBlock + 100000 - 1
	//} else if fetchFromBlock < 37900000 {
	//	fetchToBlock = fetchFromBlock + 50000 - 1
	//} else if fetchFromBlock < 38400000 {
	//	fetchToBlock = fetchFromBlock + 5000 - 1
	//} else if fetchFromBlock < 38900000 {
	//	fetchToBlock = fetchFromBlock + 2000 - 1
	//} else if fetchFromBlock < 40000000 {
	//	fetchToBlock = fetchFromBlock + 500 - 1
	//} else if fetchFromBlock < 40560000 {
	//	fetchToBlock = fetchFromBlock + 1000 - 1
	//}

	if dataEndBlock > 0 && fetchToBlock > dataEndBlock {
		fetchToBlock = dataEndBlock
	}

	// read trxs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if latestBlock == 0 || fetchToBlock >= latestBlock {
		statusCollection := mongodb.Collection("status")
		result := statusCollection.FindOne(ctx, bson.D{})
		if result.Err() != nil {
			return false, result.Err()
		}
		var status model.Status
		result.Decode(&status)
		latestBlock = status.Block

		if latestBlock > fetchFromBlock && latestBlock-fetchFromBlock < 10 {
			// It's catching up. read it block by block.
			fetchSize = 1
		}
	}

	if fetchToBlock > latestBlock {
		fetchToBlock = latestBlock
	}

	if fetchFromBlock > fetchToBlock {
		return false, errors.New(fmt.Sprintf("no more new block, block %d", fetchToBlock))
	}

	log.Printf("fetch %d to %d", fetchFromBlock, fetchToBlock)
	trxCollection := mongodb.Collection("transactions")
	cur, err := trxCollection.Find(ctx, bson.D{{"block", bson.D{{"$gte", fetchFromBlock}, {"$lte", fetchToBlock}}}})
	defer cur.Close(ctx)
	if err != nil {
		logger.Println(err)
		return false, err
	}
	for cur.Next(ctx) {
		var result model.Transaction
		err := cur.Decode(&result)
		if err != nil {
			logger.Fatal(err)
		}
		trxs = append(trxs, &result)
	}

	// read logs
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logCollection := mongodb.Collection("evmlogs")
	cur, err = logCollection.Find(ctx, bson.D{{"block", bson.D{{"$gte", fetchFromBlock}, {"$lte", fetchToBlock}}}})
	defer cur.Close(ctx)
	if err != nil {
		return false, err
	}
	for cur.Next(ctx) {
		var result model.EvmLog
		err := cur.Decode(&result)
		if err != nil {
			logger.Fatal(err)
		}
		logs = append(logs, &result)
	}

	records := mixRecords(trxs, logs)
	err = processRecords(records)
	if err != nil {
		return false, err
	}

	err = saveToStorage(fetchToBlock)
	if err != nil {
		return false, err
	}

	// create a snapshot
	if createSnapshotFlag || fetchToBlock == createSnapshotBlock {
		createSnapshotFlag = false
		if fetchToBlock == createSnapshotBlock {
			createSnapshotBlock = 0
		}
		snapshot(fetchToBlock)
	}

	fetchFromBlock = fetchToBlock + 1

	return fetchToBlock == dataEndBlock, err
}

func Snapshot() {
	snapshot(fetchToBlock)
}
