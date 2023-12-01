package main

import (
	"flag"
	"open-indexer/handlers"
	"open-indexer/loader"
)

var (
	inputfile  string
	outputfile string
)

func init() {
	flag.StringVar(&inputfile, "input", "./data/asc20.input.txt", "the filename of input data, default(./data/asc20.input.txt)")
	flag.StringVar(&outputfile, "output", "./data/asc20.output.txt", "the filename of output result, default(./data/asc20.output.txt)")

	flag.Parse()
}

func main() {

	var logger = handlers.GetLogger()

	logger.Info("start index")

	trxs, err := loader.LoadTransactionData(inputfile)
	if err != nil {
		logger.Fatalf("invalid input, %s", err)
	}

	err = handlers.ProcessUpdateARC20(trxs)
	if err != nil {
		logger.Fatalf("process error, %s", err)
	}

	logger.Info("successed")

	// print
	loader.DumpTickerInfoMap(outputfile, handlers.Tokens, handlers.UserBalances, handlers.TokenHolders)

}
