package main

import (
	"flag"
	"open-indexer/handlers"
	"open-indexer/loader"
)

var (
	inputfile1 string
	inputfile2 string
	outputfile string
)

func init() {
	flag.StringVar(&inputfile1, "transactions", "./data/transactions.input.txt", "the filename of input data, default(./data/transactions.input.txt)")
	flag.StringVar(&inputfile2, "logs", "./data/logs.input.txt", "the filename of input data, default(./data/logs.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

	flag.Parse()
}

func main() {

	var logger = handlers.GetLogger()

	logger.Info("start index")

	trxs, err := loader.LoadTransactionData(inputfile1)
	if err != nil {
		logger.Fatalf("invalid input, %s", err)
	}

	logs, err := loader.LoadLogData(inputfile2)
	if err != nil {
		logger.Fatalf("invalid input, %s", err)
	}

	records := handlers.MixRecords(trxs, logs)

	err = handlers.ProcessRecords(records)
	if err != nil {
		logger.Fatalf("process error, %s", err)
	}

	logger.Info("successed")

	// print
	tokens, userBalances, tokenHolders := handlers.GetInfo()
	loader.DumpTickerInfoMap(outputfile, tokens, userBalances, tokenHolders)

}
