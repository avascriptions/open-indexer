package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"open-indexer/model"
	"open-indexer/model/fetch"
	"open-indexer/utils"
	"strconv"
	"sync"
	"time"
)

var req = resty.New().SetTimeout(3 * time.Second)
var fetchUrl = ""

var fetchDataBlock uint64
var fetchInterrupt bool
var lastBlockNumber uint64

var cachedTranscriptions = make(map[uint64][]*model.Transaction)
var cachedLogs = make(map[uint64][]*model.EvmLog)
var cachedBlockNumber uint64

func StartFetch() {
	//value, err := db.Get([]byte("h-data-block"), nil)
	//if err == nil {
	//	fetchDataBlock = utils.BytesToUint64(value)
	//} else {
	//  fetchDataBlock = syncFromBlock - 1
	//}
	fetchDataBlock = syncFromBlock - 1
	logger.Println("start fetch data from ", fetchDataBlock+1)

	if lastBlockNumber == 0 {
		var err error
		lastBlockNumber, err = fetchLastBlockNumber()
		if err != nil {
			panic("fetch last block number error")
		}
		logger.Println("fetch: last block number", lastBlockNumber)
	}

	// fetch
	fetchInterrupt = false
	for !fetchInterrupt {
		var trxsResp fetch.BlockResponse
		var logsResp fetch.LogsResponse
		fetchDataBlock++
		err := fetchData(fetchDataBlock, &trxsResp, &logsResp)
		if err != nil {
			logger.Println("fetch error:", err.Error())
			fetchDataBlock--
			time.Sleep(time.Duration(1) * time.Second)
		} else {
			err = saveData(&trxsResp, &logsResp)
			if err != nil {
				logger.Println("fetch: save error:", err.Error())
				QuitChan <- true
				break
			}
		}
	}

	StopSuccessCount++
	logger.Println("fetch stopped")
}

func StopFetch() {
	fetchInterrupt = true
	if DataSourceType != "rpc" {
		StopSuccessCount++
	}
}

func fetchData(blockNumber uint64, blockResp *fetch.BlockResponse, logsResp *fetch.LogsResponse) error {
	start := time.Now().UnixMilli()

	if blockNumber > lastBlockNumber {
		lastBlock, err := fetchLastBlockNumber()
		if err != nil {
			return err
		}
		lastBlockNumber = lastBlock
	}

	if blockNumber > lastBlockNumber {
		return errors.New("no new blocks to be fetched")
	}

	var wg sync.WaitGroup
	var err0 error
	var err1 error

	wg.Add(2)
	go func() {
		err0 = fetchTransactions(blockNumber, blockResp)
		wg.Done()
	}()
	go func() {
		err1 = fetchContractLogs(blockNumber, logsResp)
		wg.Done()
	}()

	wg.Wait()

	if err0 != nil {
		return err0
	}
	if err1 != nil {
		return err1
	}
	costs := time.Now().UnixMilli() - start
	if costs > 200 {
		logger.Info("fetch data at #", blockNumber, " costs ", costs, " ms")
	}
	return nil
}

func fetchLastBlockNumber() (uint64, error) {
	//start := time.Now().UnixMilli()
	reqJson := fmt.Sprintf(`{"id": "indexer","jsonrpc": "2.0","method": "eth_blockNumber","params": []}`)
	resp, rerr := req.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(reqJson).
		Post(fetchUrl)
	if rerr != nil {
		logger.Info("fetch url error:", rerr)
		return 0, rerr
	}

	var response fetch.NumberResponse
	uerr := json.Unmarshal(resp.Body(), &response)
	if uerr != nil {
		logger.Info("json parse error: ", uerr)
		fmt.Println(string(resp.Body()))
		return 0, rerr
	}
	if response.Error.Code != 0 && response.Error.Message != "" {
		return 0, errors.New(fmt.Sprintf("fetch error code: %d, msg: %s", response.Error.Code, response.Error.Message))
	}
	if response.Id != "indexer" || response.JsonRpc != "2.0" {
		return 0, errors.New("fetch error data")
	}

	blockNumber := utils.HexToUint64(response.Result) - 2

	return blockNumber, nil
}

func fetchTransactions(blockNumber uint64, response *fetch.BlockResponse) (err error) {
	//start := time.Now().UnixMilli()
	block := strconv.FormatUint(blockNumber, 16)
	reqJson := fmt.Sprintf(`{"id": "indexer","jsonrpc": "2.0","method": "eth_getBlockByNumber","params": ["0x%s", %t]}`, block, true)
	resp, rerr := req.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(reqJson).
		Post(fetchUrl)
	if rerr != nil {
		logger.Info("fetch url error:", rerr)
		err = rerr
		return
	}

	uerr := json.Unmarshal(resp.Body(), &response)
	if uerr != nil {
		logger.Info("json parse error: ", uerr)
		fmt.Println(string(resp.Body()))
		err = uerr
		return
	}
	if response.Error.Code != 0 && response.Error.Message != "" {
		err = errors.New(fmt.Sprintf("fetch error code: %d, msg: %s", response.Error.Code, response.Error.Message))
		return
	}
	if response.Id != "indexer" || response.JsonRpc != "2.0" || response.Result.Hash == "" {
		err = errors.New("fetch error data")
		return
	}

	//costs := time.Now().UnixMilli() - start
	//if costs > 200 {
	//	logger.Info("fetch trxs at #", blockNumber, ", costs ", costs, " ms")
	//}
	return
}

func fetchContractLogs(blockNumber uint64, response *fetch.LogsResponse) (err error) {
	//start := time.Now().UnixMilli()
	block := strconv.FormatUint(blockNumber, 16)
	reqJson := fmt.Sprintf(`{"id": "indexer","jsonrpc": "2.0","method": "eth_getLogs","params": [{"fromBlock": "0x%s","toBlock": "0x%s"}]}`, block, block)
	resp, rerr := req.R().EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(reqJson).
		Post(fetchUrl)
	if rerr != nil {
		logger.Info("fetch url error:", rerr)
		err = rerr
		return
	}

	uerr := json.Unmarshal(resp.Body(), &response)
	if uerr != nil {
		logger.Info("json parse error: ", uerr)
		fmt.Println(string(resp.Body()))
		err = uerr
		return
	}
	if response.Error.Code != 0 && response.Error.Message != "" {
		err = errors.New(fmt.Sprintf("fetch error code: %d, msg: %s", response.Error.Code, response.Error.Message))
		return
	}
	if response.Id != "indexer" || response.JsonRpc != "2.0" {
		err = errors.New("fetch error data")
		return
	}

	//costs := time.Now().UnixMilli() - start
	//if costs > 200 {
	//	logger.Info("fetch logs at #", blockNumber, " costs ", costs, " ms")
	//}
	return
}

func saveData(blockResp *fetch.BlockResponse, logsResp *fetch.LogsResponse) error {
	var err error

	var blockNumber = utils.HexToUint64(blockResp.Result.Number)
	var timestamp = utils.HexToUint64(blockResp.Result.Timestamp)

	// save trxs
	trxs := make([]*model.Transaction, len(blockResp.Result.Transactions))
	for ti := range blockResp.Result.Transactions {
		_trx := blockResp.Result.Transactions[ti]
		trx := &model.Transaction{
			Id:        _trx.Hash,
			From:      _trx.From,
			To:        _trx.To,
			Block:     blockNumber,
			Idx:       utils.HexToUint32(_trx.TransactionIndex),
			Timestamp: timestamp,
			Input:     _trx.Input,
		}
		trxs[ti] = trx
	}
	cachedTranscriptions[blockNumber] = trxs

	// save logs
	logs := make([]*model.EvmLog, len(logsResp.Result))
	for li := range logsResp.Result {
		_log := logsResp.Result[li]
		log := &model.EvmLog{
			Hash:      _log.TransactionHash,
			Address:   _log.Address,
			Topics:    _log.Topics,
			Data:      _log.Data,
			Block:     blockNumber,
			TrxIndex:  utils.HexToUint32(_log.TransactionIndex),
			LogIndex:  utils.HexToUint32(_log.LogIndex),
			Timestamp: timestamp,
		}
		logs[li] = log
	}
	cachedLogs[blockNumber] = logs

	cachedBlockNumber = blockNumber

	// Only saved to memory for now, considering leveldb in the future

	//batch := new(leveldb.Batch)
	// save block height
	//batch.Put([]byte("h-data-block"), utils.Uint64ToBytes(blockNumber))
	//
	//// batch write
	//err = db.Write(batch, nil)
	//if err != nil {
	//	logger.Fatal(err)
	//}

	return err
}
