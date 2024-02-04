package handlers

import (
	"bufio"
	"google.golang.org/protobuf/proto"
	"io"
	"open-indexer/model"
	"open-indexer/model/serialize"
	"open-indexer/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func InitFromSnapshot() {
	if snapFile == "" {
		return
	}
	if strings.HasSuffix(snapFile, ".txt") {
		readSnapshotFromText()
		return
	}

	// read from binary
	file, err := os.Open(snapFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fs, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buffer := make([]byte, fs.Size())
	_, err = io.ReadFull(file, buffer)
	if err != nil {
		panic(err)
	}

	protoSnapshot := &serialize.Snapshot{}
	err = proto.Unmarshal(buffer, protoSnapshot)
	if err != nil {
		panic(err)
	}

	// save data
	syncFromBlock = protoSnapshot.Block
	inscriptionNumber = protoSnapshot.Number
	asc20RecordId = protoSnapshot.RecordId

	var tokenCount uint64
	var listCount uint64
	var holderCount uint64

	for idx := range protoSnapshot.Tokens {
		token := model.TokenFromProto(protoSnapshot.Tokens[idx])

		lowerTick := strings.ToLower(token.Tick)
		tokens[lowerTick] = token
		tokensByHash[token.Hash] = token

		tokenHolders[lowerTick] = make(map[string]*model.DDecimal)

		updatedTokens[lowerTick] = true

		tokenCount++
	}

	for idx := range protoSnapshot.Lists {
		list := model.ListFromProto(protoSnapshot.Lists[idx])
		lists[list.InsId] = list

		updatedLists[list.InsId] = true

		listCount++
	}

	for idx := range protoSnapshot.Balances {
		userBalances := protoSnapshot.Balances[idx]
		address := utils.BytesToHexStr(userBalances.Address)
		if address == "0x" {
			continue
		}
		balances[address] = make(map[string]*model.DDecimal)
		for idx2 := range userBalances.Balances {
			tickBalance := userBalances.Balances[idx2]

			lowerTick := strings.ToLower(tickBalance.Tick)
			if lowerTick == "" {
				continue
			}
			balance, _, _ := model.NewDecimalFromString(tickBalance.Amount)

			balances[address][lowerTick] = balance
			tokenHolders[lowerTick][address] = balance

			updatedBalances[address+lowerTick] = tickBalance.Amount

			holderCount++
		}
	}

	logger.Printf("init from snapshot, block %d, inscriptionNumber %d, asc20RecordId %d, tokens %d lists %d holders %d", syncFromBlock, inscriptionNumber, asc20RecordId, tokenCount, listCount, holderCount)

	// todo: set
	//var updatedTokens
	//var updatedBalances = make(map[string]string)
	//var updatedLists = make(map[string]bool)

	saveToStorage(syncFromBlock)

	syncFromBlock++
}

func readSnapshotFromText() {
	file, err := os.Open(snapFile)
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
			syncFromBlock = uint64(utils.ParseInt64(line))
			if scanner.Scan() {
				inscriptionNumber = uint64(utils.ParseInt64(scanner.Text()))
			}
			if scanner.Scan() {
				asc20RecordId = uint64(utils.ParseInt64(scanner.Text()))
			}
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

	logger.Printf("init from snapshot, block %d, inscriptionNumber %d, asc20RecordId %d, tokens %d lists %d holders %d", syncFromBlock, inscriptionNumber, asc20RecordId, tokenCount, listCount, holderCount)
	saveToStorage(syncFromBlock)

	syncFromBlock++
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
	token.Progress = uint32(utils.ParseInt32(row[6]))
	token.Holders = uint32(utils.ParseInt32(row[7]))
	token.Trxs = uint32(utils.ParseInt32(row[8]))
	token.CreatedAt = uint64(utils.ParseInt64(row[9]))
	token.CompletedAt = uint64(utils.ParseInt64(row[10]))
	token.Hash = row[11]

	lowerTick := strings.ToLower(token.Tick)
	tokens[lowerTick] = &token
	tokensByHash[token.Hash] = &token

	tokenHolders[lowerTick] = make(map[string]*model.DDecimal)

	updatedTokens[lowerTick] = true
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

	updatedLists[list.InsId] = true
}

func readBalance(line string) {
	row := strings.Split(line, ",")

	if len(row) != 3 {
		panic("invalid balance format:" + line)
	}

	tick := strings.Replace(row[0], "[*_*]", ",", -1)
	lowerTick := strings.ToLower(tick)
	address := row[1]
	balance, _, _ := model.NewDecimalFromString(row[2])

	tokenHolders[tick][address] = balance

	if _, ok := balances[address]; !ok {
		balances[address] = make(map[string]*model.DDecimal)
	}
	balances[address][tick] = balance

	updatedBalances[address+lowerTick] = row[2]
}

func snapshot(block uint64) {
	start := time.Now().UnixMilli()
	logger.Println("save snapshot at ", block)
	snapshotFile := "./snapshots/snap-" + strconv.FormatUint(block, 10) + ".bin"
	file, err := os.OpenFile(snapshotFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}

	msgTokens := make([]*serialize.ProtoToken, len(tokens))
	msgLists := make([]*serialize.ProtoList, len(lists))
	msgBalances := make([]*serialize.UserBalance, len(balances))

	// save tokens
	var idx = uint64(0)
	for _, token := range tokens {
		msgTokens[idx] = token.ToProtoToken()
		idx++
	}

	// save lists
	idx = 0
	for _, list := range lists {
		msgLists[idx] = list.ToProtoList()
		idx++
	}

	// save balance
	idx = 0
	var balanceCount = uint64(0)
	for address, userBalances := range balances {
		msgUserBalances := make([]*serialize.TickBalance, len(userBalances))
		var idx2 = uint64(0)
		for tick, balance := range userBalances {
			if balance.Sign() == 0 {
				continue
			}
			msgUserBalances[idx2] = &serialize.TickBalance{
				Tick:   tick,
				Amount: balance.String(),
			}
			idx2++
			balanceCount++
		}
		msgBalances[idx] = &serialize.UserBalance{
			Address:  utils.HexStrToBytes(address),
			Balances: msgUserBalances,
		}
		idx++
	}

	msgSnapshot := &serialize.Snapshot{
		Block:    block,
		Number:   inscriptionNumber,
		RecordId: asc20RecordId,
		Tokens:   msgTokens,
		Lists:    msgLists,
		Balances: msgBalances,
	}
	snapBytes, err := proto.Marshal(msgSnapshot)
	if err != nil {
		panic("snapshot error: " + err.Error())
	}
	file.Write(snapBytes)

	costs := time.Now().UnixMilli() - start
	logger.Println("save", len(tokens), " tokens, ", len(lists), " lists, ", balanceCount, " balances successfully costs ", costs, "ms at ", block)
}
