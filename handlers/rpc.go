package handlers

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"log"
	"open-indexer/model"
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
	log.Println("body")
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
