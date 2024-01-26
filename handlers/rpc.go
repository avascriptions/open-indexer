package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"google.golang.org/protobuf/proto"
	"log"
	"open-indexer/model"
	"open-indexer/model/serialize"
	"open-indexer/utils"
	"regexp"
	"strings"
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
		ErrorHandler: func(c fiber.Ctx, err error) error {
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

	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", rpcHost, rpcPort)))
}

func StopRpc() {
	if app != nil {
		app.Shutdown()
		app = nil
	}
}

//func notFound(c fiber.Ctx) error {
//	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
//		"code": "404",
//		"msg":  "rpc not found",
//	})
//}

func isNumeric(s string) bool {
	re := regexp.MustCompile(`^\d+(\.\d+)?$`)
	return re.MatchString(s)
}

func wrapResult(c fiber.Ctx) error {
	err := c.Next()
	body := string(c.Response().Body())
	if !(strings.HasPrefix(body, "{") || strings.HasPrefix(body, "[") || isNumeric(body) || body == "true" || body == "false") {
		body = "\"" + body + "\""
	}
	resp := fmt.Sprintf(`{"code":200,"data":%s}`, body)
	c.Set("Content-type", "application/json; charset=utf-8")
	c.Status(fiber.StatusOK).SendString(resp)
	return err
}

func getTokens(c fiber.Ctx) error {
	allTokens := make([]*model.Token, 0, len(tokens))
	for _, token := range tokens {
		allTokens = append(allTokens, token)
	}
	return c.JSON(allTokens)
}

func getToken(c fiber.Ctx) error {
	tick := strings.ToLower(c.Params("tick"))
	token, ok := tokens[tick]
	if !ok {
		return errors.New("token not found")
	}
	return c.JSON(token)
}

func getTokenHolders(c fiber.Ctx) error {
	tick := strings.ToLower(c.Params("tick"))
	holders, ok := tokenHolders[tick]
	if !ok {
		return errors.New("token not found")
	}
	return c.JSON(holders)
}

func getAddress(c fiber.Ctx) error {
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

func getRecordsByBlock(c fiber.Ctx) error {
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

	var result []*serialize.ProtoRecord
	var count uint32
	for ok := iter.Seek([]byte(fromKey)); ok; ok = iter.Next() {
		keys := strings.Split(string(iter.Key()), "-")
		block := utils.ParseInt64(keys[1])
		if block > toBlock {
			break
		}
		record := &serialize.ProtoRecord{}
		err := proto.Unmarshal(iter.Value(), record)
		if err != nil {
			return errors.New("read token error")
		}
		result = append(result, record)
		if count++; count >= 10000 {
			// max item is 10000
			break
		}
	}
	iter.Release()

	return c.JSON(result)
}

//func getRecordsByAddress(c fiber.Ctx) error {
//	return nil
//}

func getRecordsByTxId(c fiber.Ctx) error {
	// todo: ?
	return nil
}

func createSnapshot(c fiber.Ctx) error {
	block := c.Query("block", "0")
	if block == "0" {
		createSnapshotFlag = true
	} else {
		createSnapshotBlock = uint64(utils.ParseInt64(block))
	}
	return nil
}
