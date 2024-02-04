package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
	"open-indexer/model"
	"open-indexer/model/serialize"
	"open-indexer/utils"
	"regexp"
	"strings"
	"time"
)

var app *fiber.App

func StartRpc() {

	rpcCfg := cfg.Section("mongo")
	rpcHost := rpcCfg.Key("host").String()
	rpcPort := rpcCfg.Key("port").MustUint(3030)

	if rpcHost == "*" {
		rpcHost = ""
	}

	fiberCfg := fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": fiber.StatusInternalServerError,
				"msg":  err.Error()},
			)
		},
	}

	app = fiber.New(fiberCfg)
	app.Use(cors.New())
	//app.Use(notFound)

	api := app.Group("/v1", wrapResult)

	api.Get("/tokens/", getTokens)
	api.Get("/token/:tick", getToken)
	api.Get("/token/:tick/holders", getTokenHolders)
	api.Get("/address/:addr", getAddress)
	api.Get("/address/:addr/:tick", getAddress)
	api.Get("/records", getRecordsByBlock)
	//api.Get("/records-by-address/:address", getRecordsByAddress)
	api.Get("/records-by-txid/:txid", getRecordsByTxId)
	api.Get("/snapshot/create", createSnapshot)

	app.Listen(fmt.Sprintf("%s:%d", rpcHost, rpcPort))
}

func StopRpc() {
	if app != nil {
		app.ShutdownWithTimeout(time.Duration(5) * time.Second) // 5s
		app = nil
	}

	StopSuccessCount++
	logger.Println("rpc stopped")
}

func isNumeric(s string) bool {
	re := regexp.MustCompile(`^\d+(\.\d+)?$`)
	return re.MatchString(s)
}

func wrapResult(c *fiber.Ctx) error {
	if db == nil {
		return errors.New("database is closed")
	} else {
		err := c.Next()
		if err != nil {
			return err
		}
	}
	body := string(c.Response().Body())
	resp := fmt.Sprintf(`{"code":200,"data":%s}`, body)
	c.Set("Content-type", "application/json; charset=utf-8")
	c.Status(fiber.StatusOK).SendString(resp)
	return nil
}

func getTokens(c *fiber.Ctx) error {
	allTokens := make([]*model.Token, 0, len(tokens))
	for _, token := range tokens {
		allTokens = append(allTokens, token)
	}
	return c.JSON(allTokens)
}

func getToken(c *fiber.Ctx) error {
	tick := strings.ToLower(c.Params("tick"))
	token, ok := tokens[tick]
	if !ok {
		return errors.New("token not found")
	}
	return c.JSON(token)
}

func getTokenHolders(c *fiber.Ctx) error {
	tick := strings.ToLower(c.Params("tick"))
	holders, ok := tokenHolders[tick]
	if !ok {
		return errors.New("token not found")
	}
	return c.JSON(holders)
}

func getAddress(c *fiber.Ctx) error {
	addr := strings.ToLower(c.Params("addr"))
	tick := strings.ToLower(c.Params("tick"))
	addrBalances, ok := balances[addr]
	if !ok {
		return c.SendString("[]")
	}
	if tick != "" {
		balance, ok := addrBalances[tick]
		if ok {
			return c.SendString(balance.String())
		} else {
			return errors.New("this address doesn't have this token")
		}
	}
	return c.JSON(addrBalances)
}

func getRecordsByBlock(c *fiber.Ctx) error {
	if db == nil {
		return errors.New("database is closed")
	}
	// read tokens
	fromBlock := utils.ParseInt64(c.Query("fromBlock", "0"))
	toBlock := utils.ParseInt64(c.Query("toBlock", "0"))
	if fromBlock <= 0 {
		return errors.New("fromBlock parameter error")
	}
	if toBlock <= 0 || toBlock < fromBlock {
		return errors.New("toBlock parameter error")
	}

	fromKey := fmt.Sprintf("r-%d-0", fromBlock)
	//toKey := fmt.Sprintf("r-%d", toBlock)
	iter := db.NewIterator(nil, nil)

	var records []*model.Asc20
	var count uint32
	for ok := iter.Seek([]byte(fromKey)); ok; ok = iter.Next() {
		keys := strings.Split(string(iter.Key()), "-")
		block := utils.ParseInt64(keys[1])
		if block > toBlock {
			break
		}
		protoRecord := &serialize.ProtoRecord{}
		err := proto.Unmarshal(iter.Value(), protoRecord)
		if err != nil {
			return errors.New("read token error")
		}
		record := model.Asc20FromProto(protoRecord)
		records = append(records, record)
		if count++; count >= 10000 {
			// max item is 10000
			break
		}
	}
	iter.Release()

	return c.JSON(records)
}

//func getRecordsByAddress(c fiber.Ctx) error {
//	return nil
//}

func getRecordsByTxId(c *fiber.Ctx) error {
	txid := strings.ToLower(c.Params("txid"))
	if !strings.HasPrefix(txid, "0x") || len(txid) != 66 {
		return errors.New("incorrect txid format")
	}
	intBytes, err := db.Get([]byte("h-"+txid), nil)
	if err != nil {
		return errors.New("txid not found")
	}
	block := utils.BytesToUint64(intBytes)
	iter := db.NewIterator(util.BytesPrefix([]byte(fmt.Sprintf("r-%d-", block))), nil)
	var records []*model.Asc20
	for iter.Next() {
		protoRecord := &serialize.ProtoRecord{}
		err = proto.Unmarshal(iter.Value(), protoRecord)
		if err != nil {
			return err
		}
		record := model.Asc20FromProto(protoRecord)
		if record.Hash == txid {
			records = append(records, record)
		}
	}
	return c.JSON(records)
}

func createSnapshot(c *fiber.Ctx) error {
	block := c.Query("block", "0")
	if block == "0" {
		createSnapshotFlag = true
	} else {
		createSnapshotBlock = uint64(utils.ParseInt64(block))
	}
	return c.JSON("ok")
}
