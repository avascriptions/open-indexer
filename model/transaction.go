package model

type Transaction struct {
	Id        string
	From      string
	To        string
	Block     uint64
	Idx       uint32
	Timestamp uint64
	Input     string
}
