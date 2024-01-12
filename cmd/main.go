package main

import (
	"flag"
	"open-indexer/handlers"
	"open-indexer/loader"
)

var (
	// inputfile1 string
	// inputfile2 string
	outputfile string
)

func init() {
	// flag.StringVar(&inputfile1, "transactions", "./data/transactions.input.txt", "the filename of input data, default(./data/transactions.input.txt)")
	// flag.StringVar(&inputfile2, "logs", "./data/logs.input.txt", "the filename of input data, default(./data/logs.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

	flag.Parse()
}

func main() {

	var logger = handlers.GetLogger()

	logger.Info("start index")

	var err error
	err = handlers.SyncFromMongo()
	if err != nil {
		logger.Fatalf("sync error, %s", err)
	}

	logger.Info("start print")
	// print
	tokens, userBalances, tokenHolders := handlers.GetInfo()
	loader.DumpTickerInfoMap(outputfile, tokens, userBalances, tokenHolders)

	logger.Info("successed")
}
