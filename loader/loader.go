package loader

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"open-indexer/model"
)

func LoadTransactionData(fname string) ([]*model.Transaction, error) {

	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var trxs []*model.Transaction
	scanner := bufio.NewScanner(file)
	max := 4 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf(line)
		fields := strings.Split(line, ",")

		if len(fields) != 7 {
			return nil, fmt.Errorf("invalid data format", len(fields))
		}

		var data model.Transaction

		data.Id = fields[0]
		data.From = fields[1]
		data.To = fields[2]

		block, err := strconv.ParseUint(fields[3], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Block = block

		idx, err := strconv.ParseUint(fields[4], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Block = block

		data.Idx = uint32(idx)

		blockTime, err := strconv.ParseUint(fields[5], 10, 32)
		if err != nil {
			return nil, err
		}
		data.Timestamp = blockTime
		data.Input = fields[6]

		trxs = append(trxs, &data)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return trxs, nil
}

func DumpTickerInfoMap(fname string,
	tokens map[string]*model.Token,
	userBalances map[string]map[string]*model.DDecimal,
	tokenHolders map[string]map[string]*model.DDecimal,
) {

	file, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		log.Fatalf("open block index file failed, %s", err)
		return
	}
	defer file.Close()

	var allTickers []string
	for ticker := range tokens {
		allTickers = append(allTickers, ticker)
	}
	sort.SliceStable(allTickers, func(i, j int) bool {
		return allTickers[i] < allTickers[j]
	})

	for _, ticker := range allTickers {
		info := tokens[ticker]

		fmt.Fprintf(file, "%s trxs: %d, total: %s, minted: %s, holders: %d\n",
			info.Tick,
			info.Trxs,
			info.Max.String(),
			info.Minted,
			len(tokenHolders[ticker]),
		)

		// holders
		var allHolders []string
		for holder := range tokenHolders[ticker] {
			allHolders = append(allHolders, holder)
		}
		sort.SliceStable(allHolders, func(i, j int) bool {
			return allHolders[i] < allHolders[j]
		})

		// holders
		for _, holder := range allHolders {
			balance := tokenHolders[ticker][holder]

			fmt.Fprintf(file, "%s %s  balance: %s, tokens: %d\n",
				info.Tick,
				holder,
				balance.String(),
				len(userBalances[holder]),
			)
		}
	}
}
