package handlers

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
	"open-indexer/model"
	"open-indexer/model/serialize"
	"open-indexer/utils"
	"strings"
)

func initFromStorage() error {
	// read block
	value, err := db.Get([]byte("h-block"), nil)
	if err != nil {
		if strings.Index(err.Error(), "not found") >= 0 {
			// empty database
			logger.Println("start with empty database")
			return nil
		}
		return err
	}

	blockNumber := utils.BytesToUint64(value)
	if blockNumber == 0 {
		return errors.New("read block from database error")
	}
	syncFromBlock = blockNumber + 1

	value, err = db.Get([]byte("h-number"), nil)
	if err != nil {
		return err
	}
	inscriptionNumber = utils.BytesToUint64(value)

	value, err = db.Get([]byte("h-record-id"), nil)
	if err != nil {
		return err
	}
	asc20RecordId = utils.BytesToUint64(value)

	logger.Printf("block: %d, number: %d, asc20 id: %d", syncFromBlock, inscriptionNumber, asc20RecordId)

	//if syncFromBlock > 0 {
	//	return errors.New("test here")
	//}

	// read tokens
	iter := db.NewIterator(util.BytesPrefix([]byte("t-")), nil)
	for iter.Next() {
		protoToken := &serialize.ProtoToken{}
		err = proto.Unmarshal(iter.Value(), protoToken)
		if err != nil {
			return err
		}
		token := model.TokenFromProto(protoToken)

		lowerTick := strings.ToLower(token.Tick)
		tokens[lowerTick] = token
		tokensByHash[token.Hash] = token
		tokenHolders[lowerTick] = make(map[string]*model.DDecimal)
	}
	iter.Release()

	// read lists
	iter = db.NewIterator(util.BytesPrefix([]byte("l-")), nil)
	for iter.Next() {
		protoList := &serialize.ProtoList{}
		err = proto.Unmarshal(iter.Value(), protoList)
		if err != nil {
			return err
		}
		list := model.ListFromProto(protoList)

		lists[list.InsId] = list
	}
	iter.Release()

	// read balances
	iter = db.NewIterator(util.BytesPrefix([]byte("b-")), nil)
	for iter.Next() {
		bkey := string(iter.Key())[2:]
		address := bkey[0:42]
		tick := bkey[42:]
		balance, _, err := model.NewDecimalFromString(string(iter.Value()))
		if err != nil {
			return err
		}

		tokenHolders[tick][address] = balance

		if _, ok := balances[address]; !ok {
			balances[address] = make(map[string]*model.DDecimal)
		}
		balances[address][tick] = balance
	}
	iter.Release()

	logger.Println("current block", syncFromBlock)

	return nil
}

/***
 * Need to write data to storage, next version implement write to local database and initialize from local database,
 * currently only initialize from snapshot
 */
func saveToStorage(blockHeight uint64) error {
	var count = 0
	var err error
	batch := new(leveldb.Batch)

	// save tokens
	for tick, _ := range updatedTokens {
		count++
		token := tokens[tick]
		bytes, err := proto.Marshal(token.ToProtoToken())
		if err != nil {
			logger.Errorln("serialize token error", err.Error())
			return err
		}
		key := fmt.Sprintf("t-%s", token.Tick)
		batch.Put([]byte(key), bytes)
	}
	if count > 0 {
		logger.Println("saved", count, "tokens successfully at ", blockHeight)
	}
	updatedTokens = make(map[string]bool)

	// save balances
	count = 0
	for bkey, balance := range updatedBalances {
		//owner := bkey[0:42]
		//tick := bkey[42:]
		key := fmt.Sprintf("b-%s", bkey)
		if balance == "0" {
			batch.Delete([]byte(key))
		} else {
			batch.Put([]byte(key), []byte(balance))
		}
		//logger.Println(key, balance)
		count++
	}
	if count > 0 {
		logger.Println("saved", count, "balances successfully at ", blockHeight)
	}
	updatedBalances = make(map[string]string)

	// save lists
	count = 0
	for insId, isAdd := range updatedLists {
		key := fmt.Sprintf("l-%s", insId)
		if isAdd {
			list := lists[insId]
			bytes, err := proto.Marshal(list.ToProtoList())
			if err != nil {
				logger.Errorln("serialize list error", err.Error())
				return err
			}
			batch.Put([]byte(key), bytes)
		} else {
			batch.Delete([]byte(key))
		}
		count++
	}
	if count > 0 {
		logger.Println("saved", len(updatedLists), "lists successfully at ", blockHeight)
	}
	updatedLists = make(map[string]bool)

	// save asc-20 records
	var lastBlock = uint64(0)
	var blockIndex = uint32(0)
	for asc20Idx := range asc20Records {
		asc20Record := asc20Records[asc20Idx]
		bytes, err := proto.Marshal(asc20Record.ToProtoRecord())
		if err != nil {
			logger.Errorln("serialize token error", err.Error())
			return err
		}
		if lastBlock != asc20Record.Block {
			lastBlock = asc20Record.Block
			blockIndex = 0
		} else {
			blockIndex++
		}
		key := fmt.Sprintf("r-%d-%d", asc20Record.Block, blockIndex)
		batch.Put([]byte(key), bytes)

		key = fmt.Sprintf("h-%s", asc20Record.Hash)
		batch.Put([]byte(key), utils.Uint64ToBytes(asc20Record.Block))
	}
	if count > 0 {
		logger.Println("saved", len(asc20Records), "records successfully at ", blockHeight)
	}
	asc20Records = make([]*model.Asc20, 0)

	// save block height
	batch.Put([]byte("h-block"), utils.Uint64ToBytes(blockHeight))
	// inscription number
	batch.Put([]byte("h-number"), utils.Uint64ToBytes(inscriptionNumber))
	// asc20 id
	batch.Put([]byte("h-record-id"), utils.Uint64ToBytes(asc20RecordId))

	// batch write
	err = db.Write(batch, nil)
	if err != nil {
		logger.Fatal(err)
	}

	return err
}
