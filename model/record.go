package model

type Record struct {
	IsLog            bool
	Block            uint64
	TransactionIndex uint32
	LogIndex         uint32
	Transaction      *Transaction
	EvmLog           *EvmLog
}
