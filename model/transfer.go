package model

type Transfer struct {
	Id        uint64
	Number    uint64
	Hash      string
	From      string
	To        string
	Block     uint64
	Idx       uint32
	Timestamp uint64
}
