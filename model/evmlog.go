package model

type EvmLog struct {
	Hash      string
	Address   string
	Topics    []string
	Data      string
	Block     uint64
	TrxIndex  uint32
	LogIndex  uint32
	Timestamp uint64
}
