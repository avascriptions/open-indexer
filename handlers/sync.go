package handlers

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"open-indexer/model"
	"time"
)

// var dataStartBlock = 31918263
// var dataEndBlock = 39206794
var dataStartBlock uint64
var dataEndBlock uint64

var fetchBlock uint64
var fetchSize uint64

func init() {
	synCfg := cfg.Section("sync")
	dataStartBlock = synCfg.Key("start").MustUint64(0)
	dataEndBlock = synCfg.Key("end").MustUint64(0)
	fetchSize = synCfg.Key("size").MustUint64(1)

	fetchBlock = dataStartBlock

	if dataEndBlock > 0 && dataStartBlock > dataEndBlock {
		panic("block number error")
	}
}

func SyncFromMongo() error {
	for {
		finished, err := syncBlock()
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		//time.Sleep(500 * time.Millisecond)
	}
}

func syncBlock() (bool, error) {
	var trxs []*model.Transaction
	var logs []*model.EvmLog

	//38714454
	fetchToBlock := fetchBlock + fetchSize - 1
	if fetchBlock < 37400000 {
		fetchToBlock = fetchBlock + (fetchSize * 3000) - 1
	} else if fetchBlock < 37900000 {
		fetchToBlock = fetchBlock + (fetchSize * 1000) - 1
	} else if fetchBlock < 38714000 {
		fetchToBlock = fetchBlock + (fetchSize * 20) - 1
	} else {
		fetchToBlock = fetchBlock + (fetchSize * 5) - 1
	}
	if dataEndBlock > 0 && fetchToBlock > dataEndBlock {
		fetchToBlock = dataEndBlock
	}

	log.Printf("fetch %d to %d", fetchBlock, fetchToBlock)

	// read trxs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	collection := mongodb.Collection("transactions")
	cur, err := collection.Find(ctx, bson.D{{"block", bson.D{{"$gte", fetchBlock}, {"$lte", fetchToBlock}}}})
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
	collection = mongodb.Collection("evmlogs")
	cur, err = collection.Find(ctx, bson.D{{"block", bson.D{{"$gte", fetchBlock}, {"$lte", fetchToBlock}}}})
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

	records := MixRecords(trxs, logs)
	err = ProcessRecords(records)

	fetchBlock = fetchToBlock + 1

	return fetchToBlock == dataEndBlock, err
}
