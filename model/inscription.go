package model

type Inscription struct {
	Id          string
	Number      uint64
	From        string
	To          string
	Block       uint64
	Idx         uint32
	Timestamp   uint64
	ContentType string
	Content     string
}
