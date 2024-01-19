package handlers

import (
	"bufio"
	"fmt"
	"open-indexer/model"
	"open-indexer/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func InitFromSnapshot(snapfile string) {
	file, err := os.Open(snapfile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	max := 100 * 1024 * 1024 // 100m
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	nowType := 0
	tokenCount := 0
	listCount := 0
	holderCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "-- ") {
			dataType := line[3:]
			if dataType == "block" {
				nowType = 1
			} else if dataType == "tokens" {
				nowType = 2
			} else if dataType == "lists" {
				nowType = 3
			} else if dataType == "balances" {
				nowType = 4
			}
			continue
		}
		if nowType == 1 {
			fetchFromBlock = uint64(utils.ParseInt64(line)) + 1
		} else if nowType == 2 {
			tokenCount++
			readToken(line)
		} else if nowType == 3 {
			listCount++
			readList(line)
		} else if nowType == 4 {
			holderCount++
			readBalance(line)
		}
	}
	logger.Printf("init from snapshot, block %d, tokens %d lists %d holders %d", fetchFromBlock-1, tokenCount, listCount, holderCount)
}

func readToken(line string) {
	row := strings.Split(line, ",")

	if len(row) != 12 {
		panic("invalid token format:" + line)
	}
	var token model.Token
	token.Tick = strings.Replace(row[0], "[*_*]", ",", -1)
	token.Number = uint64(utils.ParseInt64(row[1]))
	token.Precision = int(utils.ParseInt32(row[2]))
	token.Max, _, _ = model.NewDecimalFromString(row[3])
	token.Limit, _, _ = model.NewDecimalFromString(row[4])
	token.Minted, _, _ = model.NewDecimalFromString(row[5])
	token.Progress = utils.ParseInt32(row[6])
	token.Holders = utils.ParseInt32(row[7])
	token.Trxs = utils.ParseInt32(row[8])
	token.CreatedAt = uint64(utils.ParseInt64(row[9]))
	token.CompletedAt = uint64(utils.ParseInt64(row[10]))
	token.Hash = row[11]

	lowerTick := strings.ToLower(token.Tick)
	tokens[lowerTick] = &token
	tokensByHash[token.Hash] = &token

	tokenHolders[lowerTick] = make(map[string]*model.DDecimal)
}

func readList(line string) {
	row := strings.Split(line, ",")

	if len(row) != 6 {
		panic("invalid list format:" + line)
	}

	var list model.List
	list.InsId = row[0]
	list.Owner = row[1]
	list.Exchange = row[2]
	list.Tick = strings.Replace(row[3], "[*_*]", ",", -1)
	list.Amount, _, _ = model.NewDecimalFromString(row[4])
	list.Precision = int(utils.ParseInt32(row[5]))

	lists[list.InsId] = &list
}

func readBalance(line string) {
	row := strings.Split(line, ",")

	if len(row) != 3 {
		panic("invalid balance format:" + line)
	}

	tick := strings.Replace(row[0], "[*_*]", ",", -1)
	address := row[1]
	balance, _, _ := model.NewDecimalFromString(row[2])

	tokenHolders[tick][address] = balance

	if _, ok := balances[address]; !ok {
		balances[address] = make(map[string]*model.DDecimal)
	}
	balances[address][tick] = balance
}

func snapshot(block uint64) {
	start := time.Now().UnixMilli()
	logger.Println("save snapshot at ", block)
	snapshotFile := "./snapshots/snap-" + strconv.FormatUint(block, 10) + ".txt"
	file, err := os.OpenFile(snapshotFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	fmt.Fprintf(file, "-- block\n%d\n", block)
	fmt.Fprintf(file, "-- tokens\n")
	for _, token := range tokens {
		fmt.Fprintf(file, "%s,%d,%d,%s,%s,%s,%d,%d,%d,%d,%d,%s\n",
			strings.Replace(token.Tick, ",", "[*_*]", -1),
			token.Number,
			token.Precision,
			token.Max.String(),
			token.Limit.String(),
			token.Minted.String(),
			token.Progress,
			token.Holders,
			token.Trxs,
			token.CreatedAt,
			token.CompletedAt,
			token.Hash,
		)
	}

	// save lists
	fmt.Fprintf(file, "-- lists\n")
	for _, list := range lists {
		fmt.Fprintf(file, "%s,%s,%s,%s,%s,%d\n",
			list.InsId,
			list.Owner,
			list.Exchange,
			strings.Replace(list.Tick, ",", "[*_*]", -1),
			list.Amount.String(),
			list.Precision,
		)
	}

	// save balance
	fmt.Fprintf(file, "-- balances\n")

	var userCount = uint64(0)
	for tick, tickHolders := range tokenHolders {
		for address, balance := range tickHolders {
			if balance.Sign() == 0 {
				continue
			}
			userCount++
			fmt.Fprintf(file, "%s,%s,%s\n",
				strings.Replace(tick, ",", "[*_*]", -1),
				address,
				balance.String(),
			)
		}
	}
	costs := time.Now().UnixMilli() - start
	logger.Println("save", len(tokens), "tokens, ", userCount, " balances successfully costs ", costs, "ms at ", block)
}
